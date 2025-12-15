package bitcointalk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/store"

	"github.com/PuerkitoBio/goquery"
)

type BitcointalkHandle struct {
	url    string
	filter map[string]struct{}
	limit  int
	mu     sync.Mutex
	active bool
	notified store.NotifyCache
	C       chan string
	cancel context.CancelFunc
	Logger logger.Log
	Config *config.Config
}

func NewBitcointalkHandle(cfg *config.Config, logger logger.Log) *BitcointalkHandle {
	return &BitcointalkHandle{
		url:    cfg.Bitcointalk.Url,
		limit:  cfg.Bitcointalk.Limit,
		filter: make(map[string]struct{}),
		mu:     sync.Mutex{},
		notified: store.NewLRU(50),
		active: false,
		C: 	 make(chan string, 10),
		Logger: logger,
		Config: cfg,
	}
}

func (t *BitcointalkHandle) StartMonitor() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active {
		t.Logger.Info("Bitcointalk 监控已在运行中，跳过重复启动")
		return errors.New("Bitcointalk 监控已在运行中，跳过重复启动")
	}
	// 过滤列表更新
	t.monitor()
	t.active = true
	go t.runMonitor()
	return nil
}

func (t *BitcointalkHandle) StopMonitor() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.active {
		t.cancel()
		t.active = false
		t.Logger.Info("Bitcointalk 监控已停止")
	} else {
		t.Logger.Info("Bitcointalk 监控未运行，跳过停止操作")
	}
}

func (t *BitcointalkHandle) runMonitor() {
	ctx, cf := context.WithCancel(context.Background())

	interval := time.Duration(t.Config.Bitcointalk.Interval) * time.Second
	t.cancel = cf
	ticker := time.NewTicker(interval)
    defer ticker.Stop()
	t.Logger.Info("监控循环启动","间隔: ", interval)
	for {
        select {
        case <-ticker.C:
            // 收到计时器信号，执行监控动作
			t.monitor()
		case <-ctx.Done():
			t.Logger.Info("收到终止信号，Bitcointalk 监控循环停止")
		}
	}
}

func (b *BitcointalkHandle) monitor() {
	r, err := http.Get(b.url)
	if err != nil {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("body解析失败")
		return
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		log.Println("html解析失败")
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
							b.C <- "Bitcointalk 新帖推送:\n主 题: *" + text + "*\n回复: *" + reply + "*\n点击: *" + views + "*\n直达链接: " + url
						}
					}
				}			
			}
		}
	})
}




