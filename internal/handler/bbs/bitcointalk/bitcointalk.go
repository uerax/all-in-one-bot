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
	mu       sync.Mutex
	active   bool
	notified store.NotifyCache
	C        chan models.Message
	cancel   context.CancelFunc
	Logger   logger.Log
	Config   *config.Config
}

func NewBitcointalkHandle(cfg *config.Config, logger logger.Log, c chan models.Message) *BitcointalkHandle {
	return &BitcointalkHandle{
		url:      cfg.Bitcointalk.Url,
		limit:    cfg.Bitcointalk.Limit,
		filter:   make(map[string]struct{}),
		mu:       sync.Mutex{},
		notified: store.NewLRU(50),
		active:   false,
		C:        c,
		Logger:   logger,
		Config:   cfg,
	}
}

func (t *BitcointalkHandle) StartMonitor(ctx context.Context, chatID int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active {
		return ErrAlreadyRunning
	}
	// 过滤列表更新
	t.monitor(chatID)
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

	ctx, cf := context.WithCancel(ctx)
	interval := time.Duration(t.Config.Bitcointalk.Interval) * time.Second
	t.cancel = cf
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	t.Logger.Info("监控循环启动", "间隔: ", interval)
	for {
		select {
		case <-ctx.Done():
			t.Logger.Info("收到终止信号，Bitcointalk 监控循环停止")
			return
		case <-ticker.C:
			// 收到计时器信号，执行监控动作
			t.monitor(chatID)
		}
	}
}

func (b *BitcointalkHandle) monitor(chatID int64) {
	r, err := http.Get(b.url)
	if err != nil {
		b.Logger.Error("请求Bitcointalk失败", "error:", err)
		return
	}
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
			if !b.notified.LRUExists(topic) {
				b.notified.LRUAdd(topic)
				for k := range b.filter {
					if strings.Contains(strings.ToLower(text), strings.ToLower(k)) {
						return
					}
				}
				if b.active {
					reply := strings.TrimSpace(s.Find("td").Eq(4).Text())
					views := strings.TrimSpace(s.Find("td").Eq(5).Text())
					rpy, _ := strconv.Atoi(reply)
					if rpy < 5 {
						url, exists := td.Attr("href")
						if exists {
							b.C <- models.Message{
								ChatID: chatID,
								Text:   "Bitcointalk 新帖推送:\n主 题: *" + text + "*\n回复: *" + reply + "*\n点击: *" + views + "*\n直达链接: " + url,
							}
						}
					}
				}
			}
		}
	})
}
