package nodeseek

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/store"
)

var ErrNotRunning = errors.New("Nodeseek 监控未运行")
var ErrAlreadyRunning = errors.New("Nodeseek 监控已在运行中")

type Nodeseek struct {
	mu       sync.Mutex
	C        chan models.Message
	active   bool
	tag      int64
	url      string
	interval int
	keyword  map[string]struct{}
	db       store.Store
	client   *http.Client
	cancel   context.CancelFunc
	Config   config.Nodeseek
	Log      logger.Log
}

func NewNodeseek(db store.Store, c chan models.Message, cfg config.Nodeseek, logger logger.Log) *Nodeseek {
	return &Nodeseek{
		C:        c,
		url:      cfg.Url,
		interval: cfg.Interval,
		mu:       sync.Mutex{},
		active:   false,
		keyword:  make(map[string]struct{}),
		db:       db,
		Config:   cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Log: logger,
	}
}

func (h *Nodeseek) StartMonitor(ctx context.Context, chatID int64) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.active {
		return ErrAlreadyRunning
	}
	h.active = true
	ctx, cf := context.WithCancel(ctx)
	h.cancel = cf
	go h.runMonitor(ctx, chatID)
	return nil
}

func (h *Nodeseek) StopMonitor() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.active {
		return ErrNotRunning
	}
	h.cancel()
	h.active = false
	return nil
}

func (h *Nodeseek) syncKeyword() {
	h.mu.Lock()
	defer h.mu.Unlock()
	kw, err := h.db.Set("nodeseek", "keyword")
	if err != nil {
		h.Log.Error("Nodeseek 同步 Keyword 失败", "error:", err)
	} else {
		h.keyword = kw
	}
}

func (f *Nodeseek) dailySync(ctx context.Context) {
	go func() {
		for {
			timer := time.NewTimer(time.Hour * 24)
			f.Log.Info("已开启 Keyword 每24小时定时同步")

			select {
			case <-timer.C:
				f.Log.Info("触发 Keyword 定时同步", "时间:", time.Now().Format("2006-01-02 15:04:05"))
				f.syncKeyword()
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
	}()
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Guid        int64  `xml:"guid"`
}
type Channel struct {
	Item []*Item `xml:"item"`
}
type NodeseekResp struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel *Channel `xml:"channel"`
}

func (h *Nodeseek) runMonitor(ctx context.Context, chatID int64) {

	interval := time.Duration(h.Config.Interval) * time.Second
	ticker := time.NewTicker(interval)
	h.syncKeyword()
	h.dailySync(ctx)
	defer ticker.Stop()
	h.Log.Info("Nodeseek 监控循环启动", "间隔: ", interval)
	for {
		select {
		case <-ctx.Done():
			h.Log.Info("收到终止信号，Nodeseek 监控循环停止")
			return
		case <-ticker.C:
			// 收到计时器信号，执行监控动作
			h.monitor(chatID)
		}
	}
}

func (h *Nodeseek) monitor(chatID int64) {

	r, err := h.client.Get(h.url)
	bbs := NodeseekResp{}
	if err != nil {
		return
	}
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	err = xml.Unmarshal(b, &bbs)
	if err != nil {
		return
	}
	if bbs.Channel == nil || len(bbs.Channel.Item) == 0 {
		return
	}
	msg := ""
	tag := h.tag

	for _, v := range bbs.Channel.Item {
		if tag >= v.Guid {
			continue
		}
		if tag < v.Guid {
			tag = v.Guid
		}
		v.Title = strings.ToLower(strings.TrimSpace(v.Title))
		for k := range h.keyword {
			if strings.Contains(v.Title, k) {
				msg += fmt.Sprintf("[%s](%s)\n", v.Title, v.Link)
				break
			}
		}
	}

	if tag > h.tag {
		h.tag = tag
	}

	if msg != "" {
		h.C <- models.Message{ChatID: chatID, Text: "*NodeSeek新帖:*\n" + msg}
	}
}
