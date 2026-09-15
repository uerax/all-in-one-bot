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

// Item 表示一个 DEX 监控币种。
type Item struct {
	Network string `json:"network"`           // 链名，如 "base", "eth", "solana"
	Name    string `json:"name"`              // 代币名称/符号，如 "auki", "wquil"
	Address string `json:"address,omitempty"` // 可选：代币合约地址或池子地址
	Rule    string `json:"rule,omitempty"`    // 规则
	ID      string `json:"id,omitempty"`      // 兼容旧字段
}

// QueryString 返回用于向 DEX Provider 查询的入参字符串。
func (item Item) QueryString() string {
	if item.Address != "" {
		if item.Network != "" {
			return item.Network + ":" + item.Address
		}
		return item.Address
	}
	if item.Network != "" {
		return item.Network + ":" + item.Name
	}
	if item.ID != "" {
		return item.ID
	}
	return item.Name
}

// Key 返回币种唯一标识键。
func (item Item) Key() string {
	if item.Network != "" && item.Name != "" {
		return strings.ToLower(item.Network + ":" + item.Name)
	}
	if item.Address != "" {
		return strings.ToLower(item.Address)
	}
	if item.Name != "" {
		return strings.ToLower(item.Name)
	}
	return strings.ToLower(item.ID)
}

// DexLink 生成指向 GeckoTerminal 的真实池子或网络搜索链接。
func (item Item) DexLink() string {
	if item.Network != "" && item.Address != "" {
		return fmt.Sprintf("https://www.geckoterminal.com/%s/pools/%s", item.Network, item.Address)
	}
	if item.Network != "" {
		return fmt.Sprintf("https://www.geckoterminal.com/%s/pools?search=%s", item.Network, item.Name)
	}
	return fmt.Sprintf("https://www.geckoterminal.com/search?query=%s", item.Name)
}

func formatPrice(price float64) string {
	if price >= 1.0 {
		return fmt.Sprintf("$%.2f", price)
	}
	if price >= 0.01 {
		return fmt.Sprintf("$%.4f", price)
	}
	if price >= 0.0001 {
		return fmt.Sprintf("$%.6f", price)
	}
	if price > 0 {
		return fmt.Sprintf("$%.8f", price)
	}
	return "$0.00"
}

func formatChange(pct float64) string {
	if pct > 0 {
		return fmt.Sprintf("+%.2f%%", pct)
	}
	return fmt.Sprintf("%.2f%%", pct)
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
	PriceChangePct     float64
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
	mu          sync.RWMutex
	items       []Item
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
	lookback := cfg.Lookback
	if lookback <= 0 {
		lookback = 7
	}
	yesterday := cfg.YesterdayMultiple
	if yesterday <= 0 {
		yesterday = 2.0
	}
	average := cfg.AverageMultiple
	if average <= 0 {
		average = 3.0
	}

	c := &Crocodile{
		db:     db,
		source: source,
		ruleConfig: RuleConfig{
			Lookback:          lookback,
			YesterdayMultiple: yesterday,
			AverageMultiple:   average,
		},
		lastTrigger: make(map[string]string),
		klineCache:  make(map[string]klineCacheItem),
		ch:          ch,
		client:      &http.Client{Timeout: 15 * time.Second},
		log:         log,
		cfg:         cfg,
	}
	c.initList()
	return c
}

// initList 在启动时仅同步一次数据：优先从本地/远程 Store 加载，载入内存并在首次拉取时保存至本地持久化。
func (c *Crocodile) initList() {
	if c.db == nil {
		return
	}
	var items []Item
	if err := c.db.Load("crocodile", "list", &items); err != nil {
		c.log.Error("crocodile 初始化加载监控列表失败", "error", err)
		return
	}
	c.mu.Lock()
	c.items = items
	c.mu.Unlock()

	// 首次启动若本地尚无文件，Save 确保本地持久化建立，后续重启彻底不再走远端
	if len(items) > 0 {
		if err := c.db.Save("crocodile", "list", items); err != nil {
			c.log.Warn("crocodile 初始化本地持久化失败", "error", err)
		}
	}
}

// getItems 获取当前内存中的监控币种切片拷贝。
func (c *Crocodile) getItems() []Item {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res := make([]Item, len(c.items))
	copy(res, c.items)
	return res
}

// Monitor 开启每日 UTC 00:01 量能扫描。
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

	c.ch <- models.Message{ChatID: chatID, Text: "Crocodile 监控已启动，将在每日 UTC 00:01 自动执行扫描"}

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
	// 动态分配上下文超时：默认至少 180 秒（3 分钟），或按监控条目数量（每条 3 秒 + 30 秒缓冲）动态放宽，确保平滑扫描完整执行
	timeout := 180 * time.Second
	if items, err := c.loadList(); err == nil && len(items) > 0 {
		needed := time.Duration(len(items)*3+30) * time.Second
		if needed > timeout {
			timeout = needed
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
			link := sig.DexLink()
			netStr := strings.ToUpper(sig.Network)
			if netStr == "" {
				netStr = "DEX"
			}
			fmt.Fprintf(&sb, "🟢 [%s](%s) (`%s`)\n", escapeMarkdown(sig.Name), link, netStr)
			fmt.Fprintf(&sb, "  收盘价: `%s` (`%s`) | 昨日量能: `%.2f`\n", formatPrice(sig.Close), formatChange(sig.PriceChangePct), sig.Volume)
			fmt.Fprintf(&sb, "  昨日倍数: *%.2fx* (阈值 %.1fx) | 均值倍数: *%.2fx* (阈值 %.1fx)\n\n",
				sig.YesterdayRatio, sig.YesterdayMultiple, sig.AverageRatio, sig.AverageMultiple)
		}
	}

	if len(untriggered) > 0 {
		sb.WriteString("*【⚪ 未触发币种量能现状】*\n")
		for _, sig := range untriggered {
			fmt.Fprintf(&sb, "⚪ `%s`: 收盘 `%s` (`%s`) | 昨日倍数: `%.2fx` | 均值倍数: `%.2fx`\n",
				sig.Name, formatPrice(sig.Close), formatChange(sig.PriceChangePct), sig.YesterdayRatio, sig.AverageRatio)
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
	sb.WriteString("🐊 *Crocodile DEX 监控列表*\n")
	for _, item := range items {
		link := item.DexLink()
		net := strings.ToUpper(item.Network)
		if net == "" {
			net = "DEX"
		}
		fmt.Fprintf(&sb, "- `[%s]` [%s](%s)\n", net, escapeMarkdown(item.Name), link)
	}
	c.ch <- models.Message{ChatID: chatID, Text: sb.String(), Kind: models.KindMarkdown}
}

// AddMonitor 添加一个监控币种（支持 network, name, address）。
func (c *Crocodile) AddMonitor(chatID int64, network, name, address string) {
	network = strings.ToLower(strings.TrimSpace(network))
	name = strings.TrimSpace(name)
	address = strings.TrimSpace(address)
	if network == "" && name == "" {
		c.ch <- models.Message{ChatID: chatID, Text: "参数有误: 请输入链名和代币名称，例如: base auki"}
		return
	}
	if name == "" {
		name = network
		network = "base"
	}
	item := Item{Network: network, Name: name, Address: address}
	if err := c.addItem(item); err != nil {
		c.ch <- models.Message{ChatID: chatID, Text: "添加失败: " + err.Error()}
		return
	}
	c.ch <- models.Message{
		ChatID: chatID,
		Text:   fmt.Sprintf("已添加 DEX 监控: `[%s]` *%s*", strings.ToUpper(network), escapeMarkdown(name)),
		Kind:   models.KindMarkdown,
	}
}

// DeleteMonitor 从监控列表中移除指定币种。
func (c *Crocodile) DeleteMonitor(chatID int64, target string) {
	target = strings.TrimSpace(target)
	if target == "" {
		c.ch <- models.Message{ChatID: chatID, Text: "参数有误: 币种名称不能为空"}
		return
	}
	deleted, deletedLabel, err := c.deleteItem(target)
	if err != nil {
		c.ch <- models.Message{ChatID: chatID, Text: "删除失败: " + err.Error()}
		return
	}
	if !deleted {
		c.ch <- models.Message{ChatID: chatID, Text: fmt.Sprintf("未在监控列表中找到币种: `%s`", target), Kind: models.KindMarkdown}
		return
	}

	c.ch <- models.Message{ChatID: chatID, Text: fmt.Sprintf("已删除 DEX 监控: `%s`", deletedLabel), Kind: models.KindMarkdown}
}

// RuleTip 返回当前规则配置说明。
func (c *Crocodile) RuleTip(chatID int64) {
	c.mu.Lock()
	rc := c.ruleConfig
	c.mu.Unlock()
	c.ch <- models.Message{
		ChatID: chatID,
		Text: fmt.Sprintf("*当前规则配置*\n回望天数: `%d`\n昨日倍数: `%.1f`\n均值倍数: `%.1f`\n\n请输入新参数: `lookback yesterday avg` 例如: `7 2 3`",
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
	// 由于每天 UTC 00:01 评估的是上一个完整 UTC 自然日的闭合 K 线，在同一 UTC 日期内该闭合数据不会改变。
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

		query := item.QueryString()
		cacheKey := item.Key()
		klines, ok := c.getCachedKline(cacheKey)
		if !ok {
			var err error
			klines, err = c.source.GetDailyKline(query)
			// 智能重试：排除 404 等客户端错误，仅在网络/超时/限流时等待 2.5 秒后重试一次
			if err != nil && !strings.Contains(err.Error(), "404") {
				c.log.Warn("crocodile GetDailyKline 首次失败，重试", "query", query, "error", err)
				select {
				case <-ctx.Done():
					errList = append(errList, "扫描提前取消")
					return signals, errList
				case <-time.After(2500 * time.Millisecond):
				}
				klines, err = c.source.GetDailyKline(query)
			}
			if err == nil {
				c.setCachedKline(cacheKey, klines)
			} else {
				errList = append(errList, fmt.Sprintf("%s: %v", query, err))
				continue
			}
		}

		if i < len(items)-1 {
			select {
			case <-ctx.Done():
				errList = append(errList, "扫描提前取消")
				return signals, errList
			case <-time.After(2500 * time.Millisecond):
			}
		}

		sig, err := evaluate(item, klines, rc)
		if err != nil {
			errList = append(errList, fmt.Sprintf("%s: %v", item.Name, err))
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
		key := sig.Item.Key() + ":volume_spike"
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
		link := sig.DexLink()
		netStr := strings.ToUpper(sig.Network)
		if netStr == "" {
			netStr = "DEX"
		}
		text := fmt.Sprintf(
			"🐊 *[%s]* [%s](%s) 触发日线放量突破信号\n"+
				"日期: `%s`\n"+
				"收盘价: `%s` (`%s`)\n"+
				"昨日量能倍数: `%.2fx` (阈值 %.1fx)\n"+
				"均值量能倍数: `%.2fx` (阈值 %.1fx)",
			netStr, escapeMarkdown(sig.Name), link,
			sig.Time.Format("2006-01-02"),
			formatPrice(sig.Close), formatChange(sig.PriceChangePct),
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

	// 防御性排序：确保 klines 严格按照时间戳升序排列（从旧到新，最后一个元素为最新）
	sort.Slice(klines, func(i, j int) bool {
		return klines[i].Timestamp.Before(klines[j].Timestamp)
	})

	// 剔除尚未闭合的当前 UTC 当天 K 线（例如在 UTC 00:01 运行时，当天的柱子仅有 1 分钟成交量）
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
	avg := sumVol / float64(rc.Lookback) // 包含前天在内，此前 Lookback 天的平均成交量

	if avg == 0 || prevDay.Volume == 0 {
		return nil, nil
	}

	yRatio := targetDay.Volume / prevDay.Volume
	aRatio := targetDay.Volume / avg
	triggered := yRatio >= rc.YesterdayMultiple && aRatio >= rc.AverageMultiple

	var changePct float64
	if prevDay.Close > 0 {
		changePct = (targetDay.Close - prevDay.Close) / prevDay.Close * 100
	}

	return &Signal{
		Item:               item,
		RuleConfig:         rc,
		Time:               targetDay.Timestamp,
		Close:              targetDay.Close,
		PriceChangePct:     changePct,
		Volume:             targetDay.Volume,
		PreviousVolume:     prevDay.Volume,
		PreviousAverageVol: avg,
		YesterdayRatio:     yRatio,
		AverageRatio:       aRatio,
		Triggered:          triggered,
	}, nil
}

// durationUntilNextRun 计算到下一个 UTC 00:01 的等待时长。
func durationUntilNextRun() time.Duration {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), 0, 1, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}

// loadList 获取当前监控列表（优先返回内存列表；若内存未初始化则从 store 补充加载一次）。
func (c *Crocodile) loadList() ([]Item, error) {
	c.mu.RLock()
	hasItems := len(c.items) > 0
	c.mu.RUnlock()
	if hasItems {
		return c.getItems(), nil
	}

	// 内存尚未初始化时（如直接通过 struct 实例化的测试），从 store 补充加载
	if c.db != nil {
		var items []Item
		if err := c.db.Load("crocodile", "list", &items); err == nil && len(items) > 0 {
			c.mu.Lock()
			if len(c.items) == 0 {
				c.items = items
			}
			c.mu.Unlock()
			return c.getItems(), nil
		}
	}
	return c.getItems(), nil
}

// addItem 将新条目添加并保存到内存及 db Store。
func (c *Crocodile) addItem(item Item) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 若内存尚未加载，尝试从 store 加载
	if len(c.items) == 0 && c.db != nil {
		var items []Item
		if err := c.db.Load("crocodile", "list", &items); err == nil {
			c.items = items
		}
	}

	found := false
	for i, e := range c.items {
		if strings.EqualFold(e.Key(), item.Key()) || (strings.EqualFold(e.Name, item.Name) && strings.EqualFold(e.Network, item.Network)) {
			c.items[i] = item
			found = true
			break
		}
	}
	if !found {
		c.items = append(c.items, item)
	}
	if c.db == nil {
		return fmt.Errorf("db store 未初始化")
	}
	return c.db.Save("crocodile", "list", c.items)
}

// deleteItem 从监控列表中移除指定币种（支持按 name, network:name 或 address 匹配）。
func (c *Crocodile) deleteItem(target string) (bool, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 若内存尚未加载，尝试从 store 加载
	if len(c.items) == 0 && c.db != nil {
		var items []Item
		if err := c.db.Load("crocodile", "list", &items); err == nil {
			c.items = items
		}
	}

	newList := make([]Item, 0, len(c.items))
	found := false
	deletedLabel := ""
	for _, item := range c.items {
		if !found && (strings.EqualFold(item.Name, target) || strings.EqualFold(item.Key(), target) || strings.EqualFold(item.ID, target) || (item.Address != "" && strings.EqualFold(item.Address, target))) {
			found = true
			net := strings.ToUpper(item.Network)
			if net == "" {
				net = "DEX"
			}
			deletedLabel = fmt.Sprintf("[%s] %s", net, item.Name)
			// 清理该币种在内存中的 kline 缓存与去重触发记录
			delete(c.klineCache, item.Key())
			for k := range c.lastTrigger {
				if strings.HasPrefix(k, item.Key()+":") {
					delete(c.lastTrigger, k)
				}
			}
			continue
		}
		newList = append(newList, item)
	}
	if !found {
		return false, "", nil
	}

	c.items = newList

	if c.db == nil {
		return false, "", fmt.Errorf("db store 未初始化")
	}
	if err := c.db.Save("crocodile", "list", c.items); err != nil {
		return false, "", err
	}
	return true, deletedLabel, nil
}
