package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/store"
)

const cgBaseURL = "https://api.coingecko.com/api/v3"

type Coingecko struct {
	mu     sync.Mutex
	keys   []string
	idx    uint64
	list   map[string]float64 // coin id -> holding count
	db     store.Store
	ch     chan<- models.Message
	ctx    context.Context
	cancel context.CancelFunc
	client *http.Client
	log    logger.Log
	cfg    config.Coingecko
}

type coingeckoResp struct {
	ID     string     `json:"id"`
	Symbol string     `json:"symbol"`
	Market marketData `json:"market_data"`
}

type marketData struct {
	CurrentPrice   map[string]float64 `json:"current_price"`
	PriceChange24H float64            `json:"price_change_percentage_24h"`
}

type SearchResp struct {
	Coins []SearchCoin `json:"coins"`
}

type SearchCoin struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	MarketCapRank *int   `json:"market_cap_rank"`
}

type DailyKline struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

func NewCoingecko(db store.Store, cfg config.Coingecko, ch chan<- models.Message, log logger.Log) *Coingecko {
	c := &Coingecko{
		db:     db,
		ch:     ch,
		client: &http.Client{Timeout: 10 * time.Second},
		log:    log,
		cfg:    cfg,
	}
	c.loadKeys()
	return c
}

func (c *Coingecko) loadKeys() {
	if c.cfg.Keys == "" {
		return
	}
	parts := strings.Split(c.cfg.Keys, ",")
	for _, p := range parts {
		if k := strings.TrimSpace(p); k != "" {
			c.keys = append(c.keys, k)
		}
	}
}

func (c *Coingecko) nextKey() string {
	if len(c.keys) == 0 {
		return ""
	}
	i := atomic.AddUint64(&c.idx, 1)
	return c.keys[int(i-1)%len(c.keys)]
}

func (c *Coingecko) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if key := c.nextKey(); key != "" {
		req.Header.Set("x-cg-demo-api-key", key)
	}
	return c.client.Do(req)
}

// fetchPrice 获取单个币种当前价格和持仓总值。
func (c *Coingecko) fetchPrice(id string, count float64) string {
	resp, err := c.get(fmt.Sprintf("%s/coins/%s", cgBaseURL, id))
	if err != nil {
		c.log.Error("coingecko fetchPrice 失败", "id", id, "error", err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.log.Error("coingecko fetchPrice 非200响应", "id", id, "status", resp.StatusCode)
		return ""
	}
	var r coingeckoResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		c.log.Error("coingecko fetchPrice 解析失败", "id", id, "error", err)
		return ""
	}
	price, ok := r.Market.CurrentPrice["usd"]
	if !ok || r.Symbol == "" {
		return ""
	}
	total := count * price
	change := r.Market.PriceChange24H
	return fmt.Sprintf("*%s* `$%.4f` 持仓: `$%.2f` 24h: `%.2f%%`", strings.ToUpper(r.Symbol), price, total, change)
}

// Handle 立即查询所有持仓币种价格并发送。
func (c *Coingecko) Handle(chatID int64) {
	c.syncList()

	c.mu.Lock()
	list := c.list
	c.mu.Unlock()

	if len(list) == 0 {
		c.ch <- models.Message{ChatID: chatID, Text: "持仓列表为空，请在数据库 coingecko/list 中进行配置"}
		return
	}

	lines := []string{"*CoinGecko 持仓报告*"}
	for id, count := range list {
		line := c.fetchPrice(id, count)
		if line != "" {
			lines = append(lines, line)
		}
	}
	c.ch <- models.Message{
		ChatID: chatID,
		Text:   strings.Join(lines, "\n"),
		Kind:   models.KindMarkdown,
	}
}

// Monitor 开启 24h 定时价格播报。
func (c *Coingecko) Monitor(chatID int64) {
	c.mu.Lock()
	if c.cancel != nil {
		c.mu.Unlock()
		c.ch <- models.Message{ChatID: chatID, Text: "CoinGecko 监控已在运行中"}
		return
	}
	ctx, cf := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cf
	c.mu.Unlock()

	c.syncList()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		c.log.Info("CoinGecko 24h 定时监控已启动")
		for {
			select {
			case <-ctx.Done():
				c.log.Info("CoinGecko 定时监控已停止")
				return
			case <-ticker.C:
				c.Handle(chatID)
			}
		}
	}()
}

// Stop 停止 24h 定时监控。
func (c *Coingecko) Stop(chatID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel == nil {
		c.ch <- models.Message{ChatID: chatID, Text: "CoinGecko 监控未运行"}
		return
	}
	c.cancel()
	c.cancel = nil
	c.ch <- models.Message{ChatID: chatID, Text: "CoinGecko 定时监控已关闭"}
}

// Search 搜索币种，返回前 10 条结果。
func (c *Coingecko) Search(chatID int64, query string) {
	resp, err := c.get(fmt.Sprintf("%s/search?query=%s", cgBaseURL, query))
	if err != nil {
		c.ch <- models.Message{ChatID: chatID, Text: "搜索失败: " + err.Error()}
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.ch <- models.Message{ChatID: chatID, Text: fmt.Sprintf("搜索失败: HTTP %d", resp.StatusCode)}
		return
	}
	var sr SearchResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		c.ch <- models.Message{ChatID: chatID, Text: "解析搜索结果失败"}
		return
	}
	if len(sr.Coins) == 0 {
		c.ch <- models.Message{ChatID: chatID, Text: "未找到相关币种"}
		return
	}
	limit := 10
	if len(sr.Coins) < limit {
		limit = len(sr.Coins)
	}
	lines := []string{"*CoinGecko 搜索结果*"}
	for _, coin := range sr.Coins[:limit] {
		rank := "N/A"
		if coin.MarketCapRank != nil {
			rank = fmt.Sprintf("#%d", *coin.MarketCapRank)
		}
		lines = append(lines, fmt.Sprintf("`%s` *%s* (%s) 市值排名: %s",
			coin.ID, coin.Name, strings.ToUpper(coin.Symbol), rank))
	}
	c.ch <- models.Message{
		ChatID: chatID,
		Text:   strings.Join(lines, "\n"),
		Kind:   models.KindMarkdown,
	}
}

// GetDailyKline 获取指定币种近 91 天日线数据。
func (c *Coingecko) GetDailyKline(coin string) ([]DailyKline, error) {
	url := fmt.Sprintf("%s/coins/%s/market_chart?vs_currency=usd&days=91", cgBaseURL, coin)
	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return buildDailyKlines(body)
}

type marketChartResp struct {
	Prices  [][2]float64 `json:"prices"`
	Volumes [][2]float64 `json:"total_volumes"`
}

func buildDailyKlines(body []byte) ([]DailyKline, error) {
	var raw marketChartResp
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	// 按日期聚合（CoinGecko 每天返回一个快照点）
	type dayKey = string
	type priceVol struct {
		price float64
		vol   float64
	}
	dayMap := make(map[dayKey]priceVol, len(raw.Prices))
	dayOrder := make([]string, 0, len(raw.Prices))

	for _, p := range raw.Prices {
		ts := time.UnixMilli(int64(p[0])).UTC()
		key := ts.Format("2006-01-02")
		if _, exists := dayMap[key]; !exists {
			dayOrder = append(dayOrder, key)
		}
		dayMap[key] = priceVol{price: p[1]}
	}
	for _, v := range raw.Volumes {
		ts := time.UnixMilli(int64(v[0])).UTC()
		key := ts.Format("2006-01-02")
		if pv, ok := dayMap[key]; ok {
			pv.vol = v[1]
			dayMap[key] = pv
		}
	}
	klines := make([]DailyKline, 0, len(dayOrder))
	for _, key := range dayOrder {
		pv := dayMap[key]
		t, _ := time.ParseInLocation("2006-01-02", key, time.UTC)
		klines = append(klines, DailyKline{
			Time:   t,
			Open:   pv.price,
			High:   pv.price,
			Low:    pv.price,
			Close:  pv.price,
			Volume: pv.vol,
		})
	}
	return klines, nil
}

// syncList 通过 db Store 加载持仓列表（id -> count）。
func (c *Coingecko) syncList() {
	if c.db == nil {
		return
	}
	var list map[string]float64
	if err := c.db.Load("coingecko", "list", &list); err != nil {
		c.log.Error("coingecko 从 Store 加载持仓列表失败", "error", err)
		return
	}
	c.mu.Lock()
	c.list = list
	c.mu.Unlock()
}
