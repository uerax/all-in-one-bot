package bitcointalk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/store"

	"github.com/PuerkitoBio/goquery"
)

var ErrAlreadyRunning = errors.New("Bitcointalk 监控已在运行中")
var ErrNotRunning = errors.New("Bitcointalk 监控未运行")

type BitcointalkHandle struct {
	url      string
	filter   map[string]struct{}
	limit    int
	mu       sync.RWMutex
	active   bool
	notified store.LRU
	db       store.Store
	C        chan models.Message
	client   *http.Client
	cancel   context.CancelFunc
	Logger   logger.Log
	Config   *config.Bitcointalk
}

func NewBitcointalkHandle(db store.Store, cfg *config.Bitcointalk, logger logger.Log, c chan models.Message) *BitcointalkHandle {
	return &BitcointalkHandle{
		url:    cfg.Url,
		limit:  cfg.Limit,
		filter: make(map[string]struct{}),
		db:     db,
		active: false,
		client: &http.Client{Timeout: 10 * time.Second},
		C:      c,
		Logger: logger,
		Config: cfg,
	}
}

// syncFilterLocked 更新过滤列表，调用方必须已持有写锁。
func (t *BitcointalkHandle) syncFilterLocked() {
	if m, err := t.db.Set("bitcointalk", "filter"); err != nil {
		t.Logger.Error("bitcointalk 同步过滤列表失败", "error:", err)
	} else {
		t.filter = m
	}
}

func (t *BitcointalkHandle) syncFilter() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.syncFilterLocked()
}

func (f *BitcointalkHandle) dailySync(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Hour * 24)
		f.Logger.Info("已开启 filter 每24小时定时同步")
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				f.Logger.Info("触发 filter 定时同步", "时间:", time.Now().Format("2006-01-02 15:04:05"))
				f.syncFilter()
			}
		}
	}()
}

func (t *BitcointalkHandle) StartMonitor(ctx context.Context, chatID int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active {
		return ErrAlreadyRunning
	}
	t.notified = store.NewLRUCache(t.limit)
	// Bug 1 fix: 在持锁期间调用不加锁的内部方法，避免死锁
	t.syncFilterLocked()
	// Bug 2 fix: 在持锁期间创建 cancel，防止 StopMonitor 在 goroutine 启动前拿到 nil cancel
	ctx, cf := context.WithCancel(ctx)
	t.cancel = cf
	t.active = true
	go t.runMonitor(ctx, chatID)
	return nil
}

func (t *BitcointalkHandle) StopMonitor() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return ErrNotRunning
	}

	t.cancel()
	t.active = false
	return nil
}

func (t *BitcointalkHandle) runMonitor(ctx context.Context, chatID int64) {
	t.dailySync(ctx)
	interval := time.Duration(t.Config.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	t.Logger.Info("监控循环启动", "间隔: ", interval)
	t.monitor(chatID)
	for {
		select {
		case <-ctx.Done():
			t.Logger.Info("收到终止信号，Bitcointalk 监控循环停止")
			return
		case <-ticker.C:
			t.monitor(chatID)
		}
	}
}

func (b *BitcointalkHandle) monitor(chatID int64) {
	r, err := b.client.Get(b.url)
	if err != nil {
		b.Logger.Error("请求Bitcointalk失败", "error:", err)
		return
	}
	// Bug 3 fix: 确保 response body 被关闭，防止连接泄漏
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		b.Logger.Error("body解析失败", "error:", err)
		return
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		b.Logger.Error("body解析失败", "error:", err)
		return
	}
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		td := s.Find("td").Eq(2).Find("span").Find("a")
		topic := ""
		link, exists := td.Attr("href")
		if exists && strings.Contains(link, "topic=") {
			idx := strings.Index(link, "topic=")
			topic = link[idx+6:]
		}
		if td.Text() != "" {
			text := strings.TrimSpace(td.Text())
			if !b.notified.Seen(topic) {
				// Bug 4 fix: 读 filter 时加读锁，防止与 dailySync 并发写产生数据竞争
				b.mu.RLock()
				filtered := false
				for k := range b.filter {
					if strings.Contains(strings.ToLower(text), strings.ToLower(k)) {
						filtered = true
						break
					}
				}
				b.mu.RUnlock()
				if filtered {
					return
				}
				if b.active {
					reply := strings.TrimSpace(s.Find("td").Eq(4).Text())
					views := strings.TrimSpace(s.Find("td").Eq(5).Text())
					rpy, _ := strconv.Atoi(reply)
					if rpy < 5 {
						url, exists := td.Attr("href")
						if exists {
							// Bug 5 fix: 非阻塞发送，channel 满时跳过本条推送
							select {
							case b.C <- models.Message{
								ChatID: chatID,
								Text:   "Bitcointalk 新帖推送:\n主 题: *" + text + "*\n回复: *" + reply + "*\n点击: *" + views + "*\n直达链接: " + url,
							}:
							default:
								b.Logger.Warn("消息队列已满，跳过本条推送", "topic:", topic)
							}
						}
					}
				}
			}
		}
	})
}
