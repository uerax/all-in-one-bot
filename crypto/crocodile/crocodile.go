package crocodile

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/all-in-one-bot/crypto"
	"github.com/uerax/goconf"
)

const (
	defaultRuleName       = "volume_spike"
	defaultListURL        = "https://raw.githubusercontent.com/uerax/all-in-one-bot/v2/crypto/crocodile/list.json"
	defaultLocalListPath  = "crypto/crocodile/list.json"
	defaultMonitorSeconds = 24 * 60 * 60
	defaultLookback       = 5
	defaultTodayMultiple  = 3
	defaultAvgMultiple    = 2
)

//go:embed list.json
var embeddedList []byte

type klineSource interface {
	GetDailyKline(coin string) ([]crypto.CoingeckoKline, error)
}

type Crocodile struct {
	interval    time.Duration
	source      klineSource
	load        func() ([]Item, error)
	listURL     string
	localPath   string
	rules       map[string]Rule
	ruleConfig  RuleConfig
	ch          chan<- common.AioEvent
	mu          sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	lastTrigger map[string]string
}

type Item struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Rule string `json:"rule"`
}

type Signal struct {
	Item                  Item
	Rule                  string
	RuleConfig            RuleConfig
	Time                  time.Time
	Close                 float64
	Volume                float64
	PreviousVolume        float64
	PreviousAverageVolume float64
	YesterdayRatio        float64
	AverageRatio          float64
}

type RuleConfig struct {
	Lookback          int
	YesterdayMultiple float64
	AverageMultiple   float64
}

type inspectionResult struct {
	Signal Signal
	Klines []crypto.CoingeckoKline
}

type Rule interface {
	Name() string
	Evaluate(item Item, klines []crypto.CoingeckoKline) (*Signal, error)
}

func NewCrocodile(ch ...chan<- common.AioEvent) *Crocodile {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return NewCrocodileWithSource(crypto.NewCoingecko(eventCh), eventCh)
}

func NewCrocodileWithSource(source klineSource, ch ...chan<- common.AioEvent) *Crocodile {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	interval := goconf.VarIntOrDefault(defaultMonitorSeconds, "crypto", "crocodile", "interval")
	c := &Crocodile{
		interval:    time.Duration(interval) * time.Second,
		source:      source,
		listURL:     goconf.VarStringOrDefault(defaultListURL, "crypto", "crocodile", "list"),
		localPath:   filepath.FromSlash(defaultLocalListPath),
		ruleConfig:  loadRuleConfig(),
		ch:          eventCh,
		lastTrigger: make(map[string]string),
	}
	c.load = c.loadList
	c.refreshRules()
	return c
}

func (t *Crocodile) Monitor() {
	t.mu.Lock()
	if t.cancel != nil {
		t.mu.Unlock()
		common.Send(t.ch, common.Markdown("crocodile监控已开启, 无需重复开启", true))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel
	t.mu.Unlock()

	common.Send(t.ch, common.Markdown(fmt.Sprintf("crocodile监控已开启, 轮询间隔 %s", t.interval), true))

	go t.loop(ctx)
}

func (t *Crocodile) Stop() {
	t.mu.Lock()
	if t.cancel == nil {
		t.mu.Unlock()
		common.Send(t.ch, common.Markdown("crocodile监控未开启", true))
		return
	}

	cancel := t.cancel
	t.cancel = nil
	t.ctx = nil
	t.mu.Unlock()

	cancel()
	common.Send(t.ch, common.Markdown("crocodile监控已关闭", true))
}

func (t *Crocodile) Handle() {
	t.check(true, false, false)
}

func (t *Crocodile) ListMonitor() {
	list, err := t.load()
	if err != nil {
		common.Send(t.ch, common.Markdown(fmt.Sprintf("crocodile读取监控列表失败: %v", err), true))
		return
	}

	common.Send(t.ch, common.Markdown(t.formatList(list), true))
}

func (t *Crocodile) AddMonitor(id, name string) {
	item := Item{Name: name, ID: id}
	if err := t.AddItem(item); err != nil {
		common.Send(t.ch, common.Markdown(fmt.Sprintf("crocodile添加监控失败: %v", err), true))
		return
	}

	item = normalizeItem(item)
	common.Send(t.ch, common.Markdown(fmt.Sprintf("crocodile已添加监控:\n%s\nID: `%s`", displayName(item), item.ID), true))
}

func (t *Crocodile) RuleTip() string {
	cfg := t.GetRuleConfig()
	return formatRuleDescription(cfg) + "\n\n请输入新参数: `lookback today avg`\n例如: `5 3 2`"
}

func (t *Crocodile) UpdateRule(args string) error {
	cfg, err := parseRuleConfig(args)
	if err != nil {
		return err
	}

	t.SetRuleConfig(cfg)
	common.Send(t.ch, common.Markdown("Crocodile参数已更新:\n"+formatRuleDescription(cfg), true))
	return nil
}

func formatRuleDescription(cfg RuleConfig) string {
	cfg = cfg.normalize()
	return fmt.Sprintf("当前触发规则是:\n交易量: 当天 > %.2fx 昨日, 当天 > %.2fx 平均\n平均数据取前 %d 天", cfg.YesterdayMultiple, cfg.AverageMultiple, cfg.Lookback)
}

func (t *Crocodile) loop(ctx context.Context) {
	defer func() {
		t.mu.Lock()
		if t.ctx == ctx {
			t.ctx = nil
			t.cancel = nil
		}
		t.mu.Unlock()
	}()

	// 1. 启动时当场执行一次
	t.check(true, false, true)

	// 2. 在 for 循环外部初始化 Timer (只分配一次内存)
	delay := durationUntilNextRun()
	timer := time.NewTimer(delay)

	// 确保外层函数退出时，清理掉 Timer
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			// context 被取消，直接退出即可，defer 会处理 timer.Stop()
			return

		case <-timer.C:
			// 3. 时间到了，执行扫描逻辑
			t.check(true, false, true)

			// 4. 计算明天的时间差，并复用（Reset）这个 timer
			timer.Reset(durationUntilNextRun())
		}
	}
}

func (t *Crocodile) check(notifyProblems, notifyNoHit, dedup bool) {

	list, err := t.load()
	if err != nil {
		log.Println("crocodile load list failed:", err)
		if notifyProblems {
			common.Send(t.ch, common.Markdown(fmt.Sprintf("crocodile读取list.json失败: %v", err), true))
		}
		return
	}

	if len(list) == 0 {
		if notifyProblems {
			common.Send(t.ch, common.Markdown("crocodile list.json为空, 请先配置需要监控的币种", true))
		}
		return
	}

	results := make([]inspectionResult, 0, len(list))
	errList := make([]string, 0)

	for _, item := range list {
		result, err := t.inspectWithKlines(item)
		if err != nil {
			log.Printf("crocodile inspect %s failed: %v", displayName(item), err)
			errList = append(errList, fmt.Sprintf("%s: %v", displayName(item), err))
			continue
		}
		if result == nil {
			continue
		}
		if dedup && !t.markTriggered(signalKey(result.Signal), result.Signal.Time.Format("2006-01-02")) {
			continue
		}
		results = append(results, *result)
	}

	if len(results) > 0 {
		sort.Slice(results, func(i, j int) bool {
			return strings.ToLower(displayName(results[i].Signal.Item)) < strings.ToLower(displayName(results[j].Signal.Item))
		})
		t.sendSignalMessages(results)
		t.sendCharts(results)
		return
	}

	if notifyProblems && len(errList) > 0 {
		common.Send(t.ch, common.Markdown(t.formatErrors(errList), true))
		return
	}

	_ = notifyNoHit
}

func (t *Crocodile) inspect(item Item) (*Signal, error) {
	result, err := t.inspectWithKlines(item)
	if err != nil || result == nil {
		return nil, err
	}
	return &result.Signal, nil
}

func (t *Crocodile) inspectWithKlines(item Item) (*inspectionResult, error) {
	item = normalizeItem(item)
	if item.ID == "" {
		return nil, errors.New("缺少coingecko id")
	}

	rule, err := t.rule(item.Rule)
	if err != nil {
		return nil, err
	}

	klines, err := t.source.GetDailyKline(item.ID)
	if err != nil {
		return nil, err
	}

	sig, err := rule.Evaluate(item, klines)
	if err != nil || sig == nil {
		return nil, err
	}

	return &inspectionResult{Signal: *sig, Klines: klines}, nil
}

func (t *Crocodile) rule(name string) (Rule, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultRuleName
	}

	rule, ok := t.rules[name]
	if !ok {
		return nil, fmt.Errorf("未知规则: %s", name)
	}
	return rule, nil
}

func loadRuleConfig() RuleConfig {
	return RuleConfig{
		Lookback:          goconf.VarIntOrDefault(defaultLookback, "crypto", "crocodile", "lookback"),
		YesterdayMultiple: configFloat(defaultTodayMultiple, "crypto", "crocodile", "yesterdayMultiple"),
		AverageMultiple:   configFloat(defaultAvgMultiple, "crypto", "crocodile", "averageMultiple"),
	}.normalize()
}

func configFloat(defaultValue float64, keys ...string) float64 {
	if value, err := goconf.VarFloat64(keys...); err == nil {
		return value
	}
	if value, err := goconf.VarInt(keys...); err == nil {
		return float64(value)
	}
	raw := goconf.VarStringOrDefault(strconv.FormatFloat(defaultValue, 'f', -1, 64), keys...)
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func (t *Crocodile) GetRuleConfig() RuleConfig {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ruleConfig
}

func (t *Crocodile) SetRuleConfig(cfg RuleConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ruleConfig = cfg.normalize()
	t.refreshRulesLocked()
}

func (t *Crocodile) refreshRules() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.refreshRulesLocked()
}

func (t *Crocodile) refreshRulesLocked() {
	cfg := t.ruleConfig.normalize()
	t.ruleConfig = cfg
	t.rules = map[string]Rule{
		defaultRuleName: volumeSpikeRule{
			lookback:          cfg.Lookback,
			yesterdayMultiple: cfg.YesterdayMultiple,
			averageMultiple:   cfg.AverageMultiple,
		},
	}
}

func (cfg RuleConfig) normalize() RuleConfig {
	if cfg.Lookback <= 0 {
		cfg.Lookback = defaultLookback
	}
	if cfg.YesterdayMultiple <= 0 {
		cfg.YesterdayMultiple = defaultTodayMultiple
	}
	if cfg.AverageMultiple <= 0 {
		cfg.AverageMultiple = defaultAvgMultiple
	}
	return cfg
}

func parseRuleConfig(args string) (RuleConfig, error) {
	parts := strings.Fields(args)
	if len(parts) != 3 {
		return RuleConfig{}, errors.New("参数格式应为: lookback today avg, 例如: 5 3 2")
	}

	lookback, err := strconv.Atoi(parts[0])
	if err != nil {
		return RuleConfig{}, fmt.Errorf("lookback无效: %w", err)
	}
	today, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return RuleConfig{}, fmt.Errorf("today倍数无效: %w", err)
	}
	avg, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return RuleConfig{}, fmt.Errorf("avg倍数无效: %w", err)
	}

	cfg := RuleConfig{Lookback: lookback, YesterdayMultiple: today, AverageMultiple: avg}
	if cfg.Lookback <= 0 || cfg.YesterdayMultiple <= 0 || cfg.AverageMultiple <= 0 {
		return RuleConfig{}, errors.New("参数必须大于0")
	}
	return cfg, nil
}

func (t *Crocodile) loadList() ([]Item, error) {
	list := make([]Item, 0)
	loaded := false

	if t.listURL != "" {
		remote, err := t.loadRemoteList(t.listURL)
		if err == nil {
			list = append(list, remote...)
			loaded = true
		} else {
			log.Printf("crocodile load remote list failed, fallback to local: %v", err)
		}
	}

	paths := []string{t.localPath}
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), "crypto", "crocodile", "list.json"))
	}

	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err == nil {
			local, err := decodeList(b)
			if err != nil {
				return nil, err
			}
			list = append(list, local...)
			loaded = true
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	if !loaded {
		embedded, err := decodeList(embeddedList)
		if err != nil {
			return nil, err
		}
		list = append(list, embedded...)
	}

	return mergeItems(list), nil
}

func (t *Crocodile) loadRemoteList(rawURL string) ([]Item, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("status %s: %s", res.Status, strings.TrimSpace(string(body)))
	}

	return decodeList(body)
}

func decodeList(b []byte) ([]Item, error) {
	if len(strings.TrimSpace(string(b))) == 0 {
		return []Item{}, nil
	}

	list := make([]Item, 0)
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	for i := range list {
		list[i] = normalizeItem(list[i])
	}
	return list, nil
}

func (t *Crocodile) AddItem(item Item) error {
	item = normalizeItem(item)
	if item.ID == "" {
		return errors.New("缺少 CoinGecko ID")
	}
	if item.Name == "" {
		item.Name = item.ID
	}

	local, err := t.loadLocalItems()
	if err != nil {
		return err
	}
	local = mergeItems(append(local, item))

	b, err := json.MarshalIndent(local, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(t.localPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(t.localPath, append(b, '\n'), 0644)
}

func (t *Crocodile) loadLocalItems() ([]Item, error) {
	b, err := os.ReadFile(t.localPath)
	if err == nil {
		return decodeList(b)
	}
	if errors.Is(err, os.ErrNotExist) {
		return []Item{}, nil
	}
	return nil, err
}

func mergeItems(items []Item) []Item {
	merged := make([]Item, 0, len(items))
	index := make(map[string]int, len(items))

	for _, item := range items {
		item = normalizeItem(item)
		if item.ID == "" {
			continue
		}

		key := strings.ToLower(item.ID)
		if i, ok := index[key]; ok {
			merged[i] = item
			continue
		}

		index[key] = len(merged)
		merged = append(merged, item)
	}

	return merged
}

type volumeSpikeRule struct {
	lookback          int
	yesterdayMultiple float64
	averageMultiple   float64
}

func (r volumeSpikeRule) Name() string {
	return defaultRuleName
}

func (r volumeSpikeRule) Evaluate(item Item, klines []crypto.CoingeckoKline) (*Signal, error) {
	if r.lookback <= 0 || r.yesterdayMultiple <= 0 || r.averageMultiple <= 0 {
		return nil, errors.New("规则参数无效")
	}
	if len(klines) < r.lookback+1 {
		return nil, fmt.Errorf("日K数据不足, 需要至少%d根, 当前%d根", r.lookback+1, len(klines))
	}

	klines = klines[len(klines)-r.lookback-1:]
	latest := klines[len(klines)-1]
	yesterday := klines[len(klines)-2]

	sum := 0.0
	for i := 0; i < len(klines)-1; i++ {
		sum += klines[i].Volume
	}
	avg := sum / float64(r.lookback)
	if avg <= 0 || yesterday.Volume <= 0 {
		return nil, nil
	}

	if latest.Volume >= yesterday.Volume*r.yesterdayMultiple && latest.Volume >= avg*r.averageMultiple {
		return &Signal{
			Item:                  item,
			Rule:                  r.Name(),
			RuleConfig:            RuleConfig{Lookback: r.lookback, YesterdayMultiple: r.yesterdayMultiple, AverageMultiple: r.averageMultiple},
			Time:                  latest.Time.Add(-24 * time.Hour),
			Close:                 latest.Close,
			Volume:                latest.Volume,
			PreviousVolume:        yesterday.Volume,
			PreviousAverageVolume: avg,
			YesterdayRatio:        latest.Volume / yesterday.Volume,
			AverageRatio:          latest.Volume / avg,
		}, nil
	}

	return nil, nil
}

func normalizeItem(item Item) Item {
	item.Name = strings.TrimSpace(item.Name)
	item.ID = strings.TrimSpace(item.ID)
	item.Rule = strings.TrimSpace(item.Rule)
	return item
}

func displayName(item Item) string {
	switch {
	case item.Name != "":
		return item.Name
	case item.ID != "":
		return item.ID
	default:
		return "unknown"
	}
}

func signalKey(sig Signal) string {
	return strings.ToLower(sig.Item.ID) + ":" + sig.Rule
}

func (t *Crocodile) markTriggered(key, day string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if lastDay, ok := t.lastTrigger[key]; ok && lastDay == day {
		return false
	}

	t.lastTrigger[key] = day
	return true
}

func (t *Crocodile) formatSignals(signals []Signal) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("crocodile命中 %d 个标的:", len(signals)))

	for _, sig := range signals {
		cfg := sig.RuleConfig.normalize()
		sb.WriteString("\n\n")
		sb.WriteString(displayName(sig.Item))
		sb.WriteString("\nID: ")
		sb.WriteString(sig.Item.ID)
		sb.WriteString("\n规则: ")
		sb.WriteString(sig.Rule)
		sb.WriteString("\n日期: ")
		sb.WriteString(sig.Time.Format("2006-01-02"))
		sb.WriteString(fmt.Sprintf("\n收盘价: %.8f", sig.Close))
		sb.WriteString(fmt.Sprintf("\n今日成交量: %.2f", sig.Volume))
		sb.WriteString(fmt.Sprintf("\n昨日成交量: %.2f", sig.PreviousVolume))
		sb.WriteString(fmt.Sprintf("\n前%d日均量: %.2f", cfg.Lookback, sig.PreviousAverageVolume))
		sb.WriteString(fmt.Sprintf("\n昨日倍数: %.2fx", sig.YesterdayRatio))
		sb.WriteString(fmt.Sprintf("\n均量倍数: %.2fx", sig.AverageRatio))
	}

	return sb.String()
}

func (t *Crocodile) formatList(list []Item) string {
	if len(list) == 0 {
		return "crocodile监控列表为空"
	}

	sort.Slice(list, func(i, j int) bool {
		return strings.ToLower(displayName(list[i])) < strings.ToLower(displayName(list[j]))
	})

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("crocodile当前监控 %d 个标的:", len(list)))
	for i, item := range list {
		sb.WriteString(fmt.Sprintf("\n\n%d. %s\nID: `%s`", i+1, displayName(item), item.ID))
		if item.Rule != "" {
			sb.WriteString("\n规则: ")
			sb.WriteString(item.Rule)
		}
	}

	return sb.String()
}

func (t *Crocodile) formatErrors(errList []string) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("crocodile扫描失败 %d 个标的:", len(errList)))

	limit := len(errList)
	if limit > 3 {
		limit = 3
	}

	for i := 0; i < limit; i++ {
		sb.WriteString("\n")
		sb.WriteString(errList[i])
	}

	if len(errList) > limit {
		sb.WriteString(fmt.Sprintf("\n... 其余 %d 个错误已省略", len(errList)-limit))
	}

	return sb.String()
}

func durationUntilNextRun() time.Duration {
	now := time.Now().UTC()
	// 构造今天 UTC 00:05:00 的时间对象
	next := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, time.UTC)

	// 如果当前时间已经过了今天的 00:05，则将目标时间推到明天
	if now.After(next) {
		next = next.AddDate(0, 0, 1)
	}

	return next.Sub(now)
}
