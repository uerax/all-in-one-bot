package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/goconf"
)

type Coingecko struct {
	api    string
	list   map[string]float64
	price  map[string]MarketData
	ch     chan<- common.AioEvent
	ctx    context.Context
	cancel context.CancelFunc
}

func getList() map[string]float64 {

	// 下载JSON文件
	url := "https://raw.githubusercontent.com/uerax/all-in-one-bot/v2/crypto/list/list.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	// 读取JSON文件内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("无法读取JSON文件:", err)
		return nil
	}

	// 解析JSON为Map
	filter := make(map[string]float64)
	json.Unmarshal(body, &filter)

	return filter
}

func NewCoingecko(ch ...chan<- common.AioEvent) *Coingecko {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &Coingecko{
		api:   goconf.VarStringOrDefault("", "crypto", "coingecko"),
		list:  getList(),
		price: make(map[string]MarketData),
		ch:    eventCh,
	}
}

type CoingeckoResp struct {
	ID         string     `json:"id"`
	Symbol     string     `json:"symbol"`
	MarketData MarketData `json:"market_data"`
}

type MarketData struct {
	CurrentPrice   CurrentPrice `json:"current_price"`
	TotalPrice     float64
	PriceChange24H float64 `json:"price_change_24h"`
}

type CurrentPrice struct {
	Usd float64 `json:"usd"`
}

type CoingeckoKline struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type CoingeckoSearchCoin struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	MarketCapRank *int   `json:"market_cap_rank"`
}

const coingeckoDailyKlineDays = 30

type coingeckoMarketChartResp struct {
	Prices       [][]float64 `json:"prices"`
	TotalVolumes [][]float64 `json:"total_volumes"`
}

type coingeckoSearchResp struct {
	Coins []CoingeckoSearchCoin `json:"coins"`
}

func (t *Coingecko) Price(coin string, count float64) {
	url := "https://api.coingecko.com/api/v3/coins/" + coin
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", t.api)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	price := CoingeckoResp{}
	json.Unmarshal(body, &price)
	price.MarketData.TotalPrice = count * price.MarketData.CurrentPrice.Usd
	t.price[price.Symbol] = price.MarketData
}

func coingeckoSearchURL(query string) string {
	u := "https://api.coingecko.com/api/v3/search"
	q := url.Values{}
	q.Set("query", query)
	return u + "?" + q.Encode()
}

func (t *Coingecko) Search(query string) ([]CoingeckoSearchCoin, error) {
	return t.searchFromURL(coingeckoSearchURL(query))
}

func (t *Coingecko) searchFromURL(rawURL string) ([]CoingeckoSearchCoin, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	if t.api != "" {
		req.Header.Add("x-cg-demo-api-key", t.api)
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
		log.Printf("coingecko search status %s: %s", res.Status, strings.TrimSpace(string(body)))
		return nil, fmt.Errorf("coingecko status %s", res.Status)
	}

	var result coingeckoSearchResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Coins, nil
}

func coingeckoDailyKlineURL(coin string) string {
	u := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/market_chart", url.PathEscape(coin))
	q := url.Values{}
	q.Set("vs_currency", "usd")
	q.Set("days", strconv.Itoa(coingeckoDailyKlineDays))
	return u + "?" + q.Encode()
}

func (t *Coingecko) GetDailyKline(coin string) ([]CoingeckoKline, error) {
	u := coingeckoDailyKlineURL(coin)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	if t.api != "" {
		req.Header.Add("x-cg-demo-api-key", t.api)
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
		log.Printf("coingecko kline status %s: %s", res.Status, strings.TrimSpace(string(body)))
		return nil, fmt.Errorf("coingecko status %s", res.Status)
	}

	var raw coingeckoMarketChartResp
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	klines := buildDailyKlines(raw.Prices, raw.TotalVolumes)
	if len(klines) == 0 {
		return nil, fmt.Errorf("coingecko daily kline is empty")
	}

	return klines, nil
}

func buildDailyKlines(prices [][]float64, volumes [][]float64) []CoingeckoKline {
	if len(prices) == 0 {
		return nil
	}

	volumeByDay := make(map[string]float64, len(volumes))
	for _, volume := range volumes {
		if len(volume) < 2 {
			continue
		}

		ts := time.UnixMilli(int64(volume[0])).UTC()
		day := ts.Format("2006-01-02")
		volumeByDay[day] = volume[1]
	}

	dayMap := make(map[string]CoingeckoKline, len(prices))
	order := make([]string, 0, len(prices))
	seen := make(map[string]struct{}, len(prices))

	for _, price := range prices {
		if len(price) < 2 {
			continue
		}

		ts := time.UnixMilli(int64(price[0])).UTC()
		day := ts.Format("2006-01-02")
		closePrice := price[1]
		kline, ok := dayMap[day]
		if !ok {
			kline = CoingeckoKline{
				Time:   time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC),
				Open:   closePrice,
				High:   closePrice,
				Low:    closePrice,
				Close:  closePrice,
				Volume: volumeByDay[day],
			}
		} else {
			if closePrice > kline.High {
				kline.High = closePrice
			}
			if closePrice < kline.Low {
				kline.Low = closePrice
			}
			kline.Close = closePrice
			kline.Volume = volumeByDay[day]
		}

		dayMap[day] = kline
		if _, ok := seen[day]; !ok {
			seen[day] = struct{}{}
			order = append(order, day)
		}
	}

	sort.Strings(order)
	klines := make([]CoingeckoKline, 0, len(order))
	for _, day := range order {
		klines = append(klines, dayMap[day])
	}

	return klines
}

func (t *Coingecko) SyncPrice() {
	t.SyncList()
	t.price = make(map[string]MarketData)
	for i := range t.list {
		t.Price(i, t.list[i])
	}
}

func (t *Coingecko) SyncList() {
	t.list = getList()
}

func (t *Coingecko) Handle() {
	t.SyncPrice()
	if len(t.price) != 0 {
		msg := ""
		total := 0.0
		for k, v := range t.price {
			msg += fmt.Sprintf("\n*%s* 当前价格为 *%su* 持有价值为 *%su*", k, strconv.FormatFloat(v.CurrentPrice.Usd, 'f', -1, 64), strconv.FormatFloat(v.TotalPrice, 'f', -1, 64))
			total += v.TotalPrice
		}
		common.Send(t.ch, common.Markdown(fmt.Sprintf("当前总持有价值为 *%su*%s", strconv.FormatFloat(total, 'f', -1, 64), msg), true))
	}
}

func (t *Coingecko) Stop() {
	t.cancel()
}

func (t *Coingecko) Monitor() {
	if t.cancel != nil {
		common.Send(t.ch, common.Markdown("已开启监控持有币的价格, 无需重复开启", true))
		return
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())
	ticker := time.NewTicker(24 * time.Hour)

	log.Println("开启定时监控持有币的价格")
	common.Send(t.ch, common.Markdown("开启定时监控持有币的价格", true))

	for {
		select {
		case <-ticker.C:
			t.Handle()
		case <-t.ctx.Done():
			log.Println("关闭定时监控Bitcointalk新帖")
			common.Send(t.ch, common.Markdown("关闭定时监控持有币的价格", true))
			t.cancel = nil
			return
		}
	}
}
