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
	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
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

// Signal 存储币种扫描诊断与告警信号信息。
type Signal struct {
	Item
	RuleConfig
	Time               time.Time
	Close              float64
	Volume             float64
	PreviousVolume     float64
	PreviousAverageVol float64
	YesterdayRatio     float64
	AverageRatio       float64
	Triggered          bool
}

type klineSource interface {
	GetDailyKline(coin string) ([]provider.DailyKline, error)
}

type klineCacheItem struct {
	klines    []provider.DailyKline
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
		db:     db,
		source: source,
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

// Handle 手动触发一次只读量能诊断扫描（/crocodile_check），返回包含所有币种量能现状的详细诊断报告。
func (c *Crocodile) Handle(chatID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	signals, errList := c.scan(ctx)

	c.mu.Lock()
	rc := c.ruleConfig
	c.mu.Unlock()

	var triggered []Signal
	var untriggered []Signal

	for _, sig := range signals {
		if sig.Triggered {
			triggered = append(triggered, sig)
		} else {
			untriggered = append(untriggered, sig)
		}
	}

	sort.Slice(triggered, func(i, j int) bool { return triggered[i].Name < triggered[j].Name })
	sort.Slice(untriggered, func(i, j int) bool { return untriggered[i].Name < untriggered[j].Name })

	var sb strings.Builder
	nowUTC := time.Now().UTC().Format("2006-01-02 15:04:05")
	sb.WriteString("🐊 *Crocodile 量能诊断报告* (`/crocodile_check`)\n")
	fmt.Fprintf(&sb, "扫描时间: `%s UTC`\n", nowUTC)
	fmt.Fprintf(&sb, "规则配置: 回望 `%d`天 | 昨日倍数 `%.1fx` | 均值倍数 `%.1fx`\n", rc.Lookback, rc.YesterdayMultiple, rc.AverageMultiple)
	fmt.Fprintf(&sb, "扫描总数: `%d` | 触发告警: `%d`\n\n", len(signals), len(triggered))

	if len(triggered) > 0 {
		sb.WriteString("*【🟢 触发量能信号币种】*\n")
		for _, sig := range triggered {
			link := fmt.Sprintf("https://www.coingecko.com/en/coins/%s", sig.ID)
			fmt.Fprintf(&sb, "🟢 [%s](%s) (`%s`)\n", escapeMarkdown(sig.Name), link, sig.ID)
			fmt.Fprintf(&sb, "  收盘价: `$%.4f` | 今日量能: `%.2f`\n", sig.Close, sig.Volume)
			fmt.Fprintf(&sb, "  昨日倍数: *%.2fx* (阈值 %.1fx) | 均值倍数: *%.2fx* (阈值 %.1fx)\n\n",
				sig.YesterdayRatio, sig.YesterdayMultiple, sig.AverageRatio, sig.AverageMultiple)
		}
	}

	if len(untriggered) > 0 {
		sb.WriteString("*【⚪ 未触发币种量能现状】*\n")
		for _, sig := range untriggered {
			fmt.Fprintf(&sb, "⚪ `%s`: 收盘 `$%.4f` | 昨日倍数: `%.2fx` | 均值倍数: `%.2fx`\n",
				sig.Name, sig.Close, sig.YesterdayRatio, sig.AverageRatio)
		}
		sb.WriteString("\n")
	}

	if len(errList) > 0 {
		sb.WriteString("*【⚠️ 扫描异常说明】*\n")
		limit := min(5, len(errList))
		for _, errStr := range errList[:limit] {
			fmt.Fprintf(&sb, "- %s\n", errStr)
		}
	}

	c.ch <- models.Message{
		ChatID: chatID,
		Text:   strings.TrimSpace(sb.String()),
		Kind:   models.KindMarkdown,
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

func (c *Crocodile) getCachedKline(coinID string) ([]provider.DailyKline, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.klineCache[coinID]
	if !ok {
		return nil, false
	}
	// 由于每天 UTC 00:05 评估的是上一个完整 UTC 自然日的闭合 K 线，在同一 UTC 日期内该闭合数据不会改变。
	// 双重安全校验：既要求属于同一个 UTC 自然日，又限制单次缓存最长不超过 12 小时（防止跨天边界遗留）。
	todayUTC := time.Now().UTC().Format("2006-01-02")
	if item.updatedAt.UTC().Format("2006-01-02") == todayUTC && time.Since(item.updatedAt) < 12*time.Hour {
		return item.klines, true
	}
	return nil, false
}

func (c *Crocodile) setCachedKline(coinID string, klines []provider.DailyKline) {
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
		if !sig.Triggered {
			continue
		}
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
// 自动剔除未闭合的“今天（UTC 当天）”K线，确保对比对象为【已闭合的上一自然日 Full 24h K线】与【前天及前 N 天均值】。
func evaluate(item Item, klines []provider.DailyKline, rc RuleConfig) (*Signal, error) {
	if len(klines) == 0 {
		return nil, fmt.Errorf("kline 数据为空")
	}

	// 剔除尚未闭合的当前 UTC 当天 K 线（例如在 UTC 00:05 运行时，当天的柱子仅有 5 分钟成交量）
	todayUTC := time.Now().UTC().Format("2006-01-02")
	if klines[len(klines)-1].Timestamp.UTC().Format("2006-01-02") == todayUTC {
		klines = klines[:len(klines)-1]
	}

	need := rc.Lookback + 1
	if len(klines) < need {
		return nil, fmt.Errorf("完全闭合的 kline 数据不足 (%d < %d)", len(klines), need)
	}

	recent := klines[len(klines)-need:]
	targetDay := recent[len(recent)-1] // 【刚刚闭合的上一自然日 Full 24h K线】
	prevDay := recent[len(recent)-2]   // 【前天（再前一个自然日）Full 24h K线】

	var sumVol float64
	for _, k := range recent[:rc.Lookback] {
		sumVol += k.Volume
	}
	avg := sumVol / float64(rc.Lookback) // 包含前天在内，此前 Lookback (例如 5) 天的平均成交量

	if avg == 0 || prevDay.Volume == 0 {
		return nil, nil
	}

	yRatio := targetDay.Volume / prevDay.Volume
	aRatio := targetDay.Volume / avg
	triggered := yRatio >= rc.YesterdayMultiple && aRatio >= rc.AverageMultiple

	return &Signal{
		Item:               item,
		RuleConfig:         rc,
		Time:               targetDay.Timestamp,
		Close:              targetDay.Close,
		Volume:             targetDay.Volume,
		PreviousVolume:     prevDay.Volume,
		PreviousAverageVol: avg,
		YesterdayRatio:     yRatio,
		AverageRatio:       aRatio,
		Triggered:          triggered,
	}, nil
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
