package crocodile

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	cg "github.com/uerax/all-in-one-bot/lite/internal/handler/crypto/coingecko"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/store"
)

// Item 表示一个监控币种。
type Item struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Rule string `json:"rule,omitempty"`
}

// RuleConfig 存储量能扫描的阈值配置。
type RuleConfig struct {
	Lookback          int
	YesterdayMultiple float64
	AverageMultiple   float64
}

// Signal 存储触发信号的详细信息。
type Signal struct {
	Item
	RuleConfig
	Time                 time.Time
	Close                float64
	Volume               float64
	PreviousVolume       float64
	PreviousAverageVol   float64
	YesterdayRatio       float64
	AverageRatio         float64
}

type klineSource interface {
	GetDailyKline(coin string) ([]cg.DailyKline, error)
}

type klineCacheItem struct {
	klines    []cg.DailyKline
	updatedAt time.Time
}

// Crocodile 是量能监控引擎。
type Crocodile struct {
	mu          sync.Mutex
	source      klineSource
	ruleConfig  RuleConfig
	lastTrigger map[string]string // "coinID:rule" -> "2006-01-02"
	klineCache  map[string]klineCacheItem
	ctx         context.Context
	cancel      context.CancelFunc
	db          store.Store
	ch          chan<- models.Message
	client      *http.Client
	log         logger.Log
	cfg         config.Crocodile
}

func NewCrocodile(db store.Store, source klineSource, ch chan<- models.Message, cfg config.Crocodile, log logger.Log) *Crocodile {
	return &Crocodile{
		db:          db,
		source:      source,
		ruleConfig: RuleConfig{
			Lookback:          cfg.Lookback,
			YesterdayMultiple: cfg.YesterdayMultiple,
			AverageMultiple:   cfg.AverageMultiple,
		},
		lastTrigger: make(map[string]string),
		klineCache:  make(map[string]klineCacheItem),
		ch:          ch,
		client:      &http.Client{Timeout: 15 * time.Second},
		log:         log,
		cfg:         cfg,
	}
}

// Monitor 开启每日 UTC 00:05 量能扫描。
func (c *Crocodile) Monitor(chatID int64) {
	c.mu.Lock()
	if c.cancel != nil {
		c.mu.Unlock()
		c.ch <- models.Message{ChatID: chatID, Text: "Crocodile 监控已在运行中"}
		return
	}
	ctx, cf := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cf
	c.mu.Unlock()

	go func() {
		c.log.Info("Crocodile 监控已启动")
		c.check(chatID)
		for {
			next := durationUntilNextRun()
			timer := time.NewTimer(next)
			select {
			case <-ctx.Done():
				timer.Stop()
				c.log.Info("Crocodile 监控已停止")
				return
			case <-timer.C:
				c.check(chatID)
			}
		}
	}()
}

// Stop 停止监控。
func (c *Crocodile) Stop(chatID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel == nil {
		c.ch <- models.Message{ChatID: chatID, Text: "Crocodile 监控未运行"}
		return
	}
	c.cancel()
	c.cancel = nil
	c.ch <- models.Message{ChatID: chatID, Text: "Crocodile 监控已关闭"}
}

// Handle 手动触发一次扫描（只读查询，不写去重状态，设置 30s 超时上下文）。
func (c *Crocodile) Handle(chatID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	signals, errList := c.scan(ctx)
	if len(signals) == 0 {
		reply := "扫描完成，未触发量能信号"
		if len(errList) > 0 {
			reply += ":\n" + strings.Join(errList[:min(3, len(errList))], "\n")
		}
		c.ch <- models.Message{ChatID: chatID, Text: reply}
		return
	}
	sort.Slice(signals, func(i, j int) bool { return signals[i].Name < signals[j].Name })
	c.sendSignals(chatID, signals)
	if len(errList) > 0 {
		c.ch <- models.Message{ChatID: chatID, Text: "扫描期间部分错误:\n" + strings.Join(errList[:min(3, len(errList))], "\n")}
	}
}

// ListMonitor 展示当前监控列表。
func (c *Crocodile) ListMonitor(chatID int64) {
	items, err := c.loadList()
	if err != nil || len(items) == 0 {
		c.ch <- models.Message{ChatID: chatID, Text: "监控列表为空，请使用 /crocodile_add 添加币种"}
		return
	}
	var sb strings.Builder
	sb.WriteString("*Crocodile 监控列表*\n")
	for _, item := range items {
		link := fmt.Sprintf("https://www.coingecko.com/en/coins/%s", item.ID)
		sb.WriteString(fmt.Sprintf("`%s`: [%s](%s)\n", item.Name, item.ID, link))
	}
	c.ch <- models.Message{ChatID: chatID, Text: sb.String(), Kind: models.KindMarkdown}
}

// AddMonitor 添加一个监控币种。
func (c *Crocodile) AddMonitor(chatID int64, id, name string) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" {
		c.ch <- models.Message{ChatID: chatID, Text: "参数有误: 币种 ID 不能为空"}
		return
	}
	if name == "" {
		name = id
	}
	if err := c.addItem(Item{ID: id, Name: name}); err != nil {
		c.ch <- models.Message{ChatID: chatID, Text: "添加失败: " + err.Error()}
		return
	}
	c.ch <- models.Message{ChatID: chatID, Text: fmt.Sprintf("已添加监控: `%s` (%s)", id, name), Kind: models.KindMarkdown}
}

// RuleTip 返回当前规则配置说明。
func (c *Crocodile) RuleTip(chatID int64) {
	c.mu.Lock()
	rc := c.ruleConfig
	c.mu.Unlock()
	c.ch <- models.Message{
		ChatID: chatID,
		Text: fmt.Sprintf("*当前规则配置*\n回望天数: `%d`\n今日倍数: `%.1f`\n均值倍数: `%.1f`\n\n请输入新参数: `lookback today avg` 例如: `5 3 2`",
			rc.Lookback, rc.YesterdayMultiple, rc.AverageMultiple),
		Kind: models.KindMarkdown,
	}
}

// UpdateRule 更新规则阈值。
func (c *Crocodile) UpdateRule(chatID int64, lookback int, yesterday, average float64) {
	if lookback <= 0 || yesterday <= 0 || average <= 0 {
		c.ch <- models.Message{ChatID: chatID, Text: "Crocodile 参数更新失败: 所有参数必须大于 0"}
		return
	}
	c.mu.Lock()
	c.ruleConfig = RuleConfig{
		Lookback:          lookback,
		YesterdayMultiple: yesterday,
		AverageMultiple:   average,
	}
	c.mu.Unlock()
	c.ch <- models.Message{
		ChatID: chatID,
		Text:   fmt.Sprintf("规则已更新: 回望=%d 今日倍数=%.1f 均值倍数=%.1f", lookback, yesterday, average),
	}
}

func (c *Crocodile) getCachedKline(coinID string) ([]cg.DailyKline, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.klineCache[coinID]
	if ok && time.Since(item.updatedAt) < 10*time.Minute {
		return item.klines, true
	}
	return nil, false
}

func (c *Crocodile) setCachedKline(coinID string, klines []cg.DailyKline) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.klineCache[coinID] = klineCacheItem{
		klines:    klines,
		updatedAt: time.Now(),
	}
}

// scan 执行一次只读量能扫描：拉取 K 线、评估信号、报告错误。
// 支持通过 Context 取消，内置 10 分钟短效缓存与智能重试机制。
func (c *Crocodile) scan(ctx context.Context) ([]Signal, []string) {
	items, err := c.loadList()
	if err != nil || len(items) == 0 {
		return nil, []string{"监控列表为空，请先使用 /crocodile_add 添加币种"}
	}

	c.mu.Lock()
	rc := c.ruleConfig
	c.mu.Unlock()

	var signals []Signal
	var errList []string

	for i, item := range items {
		select {
		case <-ctx.Done():
			errList = append(errList, "扫描提前取消")
			return signals, errList
		default:
		}

		klines, ok := c.getCachedKline(item.ID)
		if !ok {
			var err error
			klines, err = c.source.GetDailyKline(item.ID)
			// 智能重试：排除 404 等客户端错误，仅在网络/超时/限流时 500ms 重试一次
			if err != nil && !strings.Contains(err.Error(), "404") {
				c.log.Warn("crocodile GetDailyKline 首次失败，重试", "id", item.ID, "error", err)
				select {
				case <-ctx.Done():
					errList = append(errList, "扫描提前取消")
					return signals, errList
				case <-time.After(500 * time.Millisecond):
				}
				klines, err = c.source.GetDailyKline(item.ID)
			}
			if err == nil {
				c.setCachedKline(item.ID, klines)
			} else {
				errList = append(errList, fmt.Sprintf("%s: %v", item.ID, err))
				continue
			}
		}

		if i < len(items)-1 {
			select {
			case <-ctx.Done():
				errList = append(errList, "扫描提前取消")
				return signals, errList
			case <-time.After(150 * time.Millisecond):
			}
		}

		sig, err := evaluate(item, klines, rc)
		if err != nil {
			errList = append(errList, fmt.Sprintf("%s: %v", item.ID, err))
			continue
		}
		if sig != nil {
			signals = append(signals, *sig)
		}
	}
	return signals, errList
}

// check 定时扫描入口（Monitor 每日触发）：scan + 对已通知信号去重写入。
// 无信号时静默，仅在存在错误时反馈。
func (c *Crocodile) check(chatID int64) {
	c.mu.Lock()
	ctx := c.ctx
	c.mu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}

	signals, errList := c.scan(ctx)

	var notified []Signal
	for i := range signals {
		sig := &signals[i]
		key := sig.Item.ID + ":volume_spike"
		day := sig.Time.Format("2006-01-02")
		c.mu.Lock()
		already := c.lastTrigger[key] == day
		if !already {
			c.lastTrigger[key] = day
			// 容量上限保护
			if len(c.lastTrigger) > 1000 {
				for k := range c.lastTrigger {
					delete(c.lastTrigger, k)
					break
				}
			}
		}
		c.mu.Unlock()
		if !already {
			notified = append(notified, *sig)
		}
	}

	if len(notified) == 0 {
		if len(errList) > 0 {
			c.ch <- models.Message{ChatID: chatID, Text: "扫描完成，未触发量能信号:\n" + strings.Join(errList[:min(3, len(errList))], "\n")}
		}
		return
	}

	sort.Slice(notified, func(i, j int) bool { return notified[i].Name < notified[j].Name })
	c.sendSignals(chatID, notified)

	if len(errList) > 0 {
		c.ch <- models.Message{ChatID: chatID, Text: "扫描期间部分错误:\n" + strings.Join(errList[:min(3, len(errList))], "\n")}
	}
}

// sendSignals 将信号结果以 Markdown 格式发送。
func (c *Crocodile) sendSignals(chatID int64, signals []Signal) {
	for _, sig := range signals {
		link := fmt.Sprintf("https://www.coingecko.com/en/coins/%s", sig.ID)
		text := fmt.Sprintf(
			"🐊 [%s](%s) 触发买入信号\n"+
				"日期: `%s`\n"+
				"收盘价: `$%.4f`\n"+
				"今日量能倍数: `%.2fx` (阈值 %.1fx)\n"+
				"均值量能倍数: `%.2fx` (阈值 %.1fx)",
			escapeMarkdown(sig.Name), link,
			sig.Time.Format("2006-01-02"),
			sig.Close,
			sig.YesterdayRatio, sig.YesterdayMultiple,
			sig.AverageRatio, sig.AverageMultiple,
		)
		c.ch <- models.Message{ChatID: chatID, Text: text, Kind: models.KindMarkdown}
	}
}

// escapeMarkdown 转义 Telegram MarkdownV1 保留字符。
// 消息经 dispatcher 以 tb.ModeMarkdown (V1) 发送，仅需转义 _ * ` [；
// 转义 V2 专有字符 (- . () 等) 会在 V1 下原样输出 \- \. \(
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "`", "\\`", "[", "\\[",
	)
	return replacer.Replace(s)
}

// evaluate 对单个币种执行量能规则评估。
func evaluate(item Item, klines []cg.DailyKline, rc RuleConfig) (*Signal, error) {
	need := rc.Lookback + 1
	if len(klines) < need {
		return nil, fmt.Errorf("kline 数据不足 (%d < %d)", len(klines), need)
	}
	recent := klines[len(klines)-need:]
	latest := recent[len(recent)-1]
	yesterday := recent[len(recent)-2]
	var sumVol float64
	for _, k := range recent[:rc.Lookback] {
		sumVol += k.Volume
	}
	avg := sumVol / float64(rc.Lookback)
	if avg == 0 || yesterday.Volume == 0 {
		return nil, nil
	}
	yRatio := latest.Volume / yesterday.Volume
	aRatio := latest.Volume / avg
	if yRatio >= rc.YesterdayMultiple && aRatio >= rc.AverageMultiple {
		return &Signal{
			Item:               item,
			RuleConfig:         rc,
			Time:               latest.Time,
			Close:              latest.Close,
			Volume:             latest.Volume,
			PreviousVolume:     yesterday.Volume,
			PreviousAverageVol: avg,
			YesterdayRatio:     yRatio,
			AverageRatio:       aRatio,
		}, nil
	}
	return nil, nil
}

// durationUntilNextRun 计算到下一个 UTC 00:05 的等待时长。
func durationUntilNextRun() time.Duration {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}

// loadList 通过 db Store 加载监控列表。
func (c *Crocodile) loadList() ([]Item, error) {
	if c.db == nil {
		return nil, fmt.Errorf("db store 未初始化")
	}
	var items []Item
	if err := c.db.Load("crocodile", "list", &items); err != nil {
		return nil, err
	}
	return items, nil
}

// addItem 将新条目添加并保存到 db Store。
func (c *Crocodile) addItem(item Item) error {
	existing, _ := c.loadList()
	found := false
	for i, e := range existing {
		if e.ID == item.ID {
			existing[i] = item
			found = true
			break
		}
	}
	if !found {
		existing = append(existing, item)
	}
	if c.db == nil {
		return fmt.Errorf("db store 未初始化")
	}
	return c.db.Save("crocodile", "list", existing)
}
