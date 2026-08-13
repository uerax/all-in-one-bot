package coingecko

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

const cgBaseURL = "https://api.coingecko.com/api/v3"

type Provider struct {
	keys   []string
	idx    uint64
	client *http.Client
	log    logger.Log
	cfg    config.Coingecko
}

func NewProvider(cfg config.Coingecko, log logger.Log) *Provider {
	p := &Provider{
		client: &http.Client{Timeout: 10 * time.Second},
		log:    log,
		cfg:    cfg,
	}
	p.loadKeys()
	return p
}

func (p *Provider) Name() string {
	return "coingecko"
}

func (p *Provider) Type() provider.ProviderType {
	return provider.ProviderTypeCEX
}

func (p *Provider) loadKeys() {
	if p.cfg.Keys == "" {
		return
	}
	parts := strings.Split(p.cfg.Keys, ",")
	for _, part := range parts {
		if k := strings.TrimSpace(part); k != "" {
			p.keys = append(p.keys, k)
		}
	}
}

func (p *Provider) nextKey() string {
	if len(p.keys) == 0 {
		return ""
	}
	i := atomic.AddUint64(&p.idx, 1)
	return p.keys[int(i-1)%len(p.keys)]
}

func (p *Provider) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if key := p.nextKey(); key != "" {
		req.Header.Set("x-cg-demo-api-key", key)
	}
	return p.client.Do(req)
}

type coingeckoCoinResp struct {
	ID        string     `json:"id"`
	Symbol    string     `json:"symbol"`
	Name      string     `json:"name"`
	MarketCap *int       `json:"market_cap_rank"`
	Market    marketData `json:"market_data"`
}

type marketData struct {
	CurrentPrice   map[string]float64 `json:"current_price"`
	PriceChange24H float64            `json:"price_change_percentage_24h"`
	TotalVolume    map[string]float64 `json:"total_volume"`
	MarketCap      map[string]float64 `json:"market_cap"`
}

// GetPrice fetches single coin market data by CoinGecko ID or ticker.
func (p *Provider) GetPrice(query string) (*provider.MarketData, error) {
	id := strings.ToLower(strings.TrimSpace(query))

	resp, err := p.get(fmt.Sprintf("%s/coins/%s", cgBaseURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Fallback: search coin ID by query first
		searchResults, sErr := p.Search(id)
		if sErr == nil && len(searchResults) > 0 {
			return p.GetPrice(searchResults[0].ID)
		}
		return nil, provider.ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko: http %d", resp.StatusCode)
	}

	var r coingeckoCoinResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("coingecko: decode error: %w", err)
	}

	price := r.Market.CurrentPrice["usd"]
	vol := r.Market.TotalVolume["usd"]
	mcap := r.Market.MarketCap["usd"]
	rank := 0
	if r.MarketCap != nil {
		rank = *r.MarketCap
	}

	return &provider.MarketData{
		Symbol:        strings.ToUpper(r.Symbol),
		Name:          r.Name,
		PriceUSD:      price,
		Change24h:     r.Market.PriceChange24H,
		Volume24hUSD:  vol,
		MarketCap:     mcap,
		MarketCapRank: rank,
		Source:        p.Name(),
		SourceType:    p.Type(),
		UpdatedAt:     time.Now(),
	}, nil
}

// GetMultiPrices retrieves price data for multiple CoinGecko coin IDs.
func (p *Provider) GetMultiPrices(queries []string) (map[string]*provider.MarketData, error) {
	res := make(map[string]*provider.MarketData)
	for _, q := range queries {
		data, err := p.GetPrice(q)
		if err == nil && data != nil {
			res[q] = data
		}
	}
	return res, nil
}

type searchResp struct {
	Coins []searchCoin `json:"coins"`
}

type searchCoin struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	MarketCapRank *int   `json:"market_cap_rank"`
}

// SearchCEX performs CEX asset search.
func (p *Provider) SearchCEX(query string) ([]provider.SearchResult, error) {
	return p.Search(query)
}

// Search queries CoinGecko search endpoint.
func (p *Provider) Search(query string) ([]provider.SearchResult, error) {
	resp, err := p.get(fmt.Sprintf("%s/search?query=%s", cgBaseURL, query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko search: http %d", resp.StatusCode)
	}

	var sr searchResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("coingecko search: decode error: %w", err)
	}

	var results []provider.SearchResult
	for _, coin := range sr.Coins {
		rank := 0
		if coin.MarketCapRank != nil {
			rank = *coin.MarketCapRank
		}
		results = append(results, provider.SearchResult{
			ID:            coin.ID,
			Symbol:        strings.ToUpper(coin.Symbol),
			Name:          coin.Name,
			MarketCapRank: rank,
			Source:        p.Name(),
			SourceType:    p.Type(),
		})
	}
	return results, nil
}

// GetDailyKline fetches 91-day historical daily K-lines for CoinGecko assets.
func (p *Provider) GetDailyKline(coin string) ([]provider.DailyKline, error) {
	url := fmt.Sprintf("%s/coins/%s/market_chart?vs_currency=usd&days=91", cgBaseURL, coin)
	resp, err := p.get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko kline: http %d", resp.StatusCode)
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

func buildDailyKlines(body []byte) ([]provider.DailyKline, error) {
	var raw marketChartResp
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
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
	klines := make([]provider.DailyKline, 0, len(dayOrder))
	for _, key := range dayOrder {
		pv := dayMap[key]
		t, _ := time.ParseInLocation("2006-01-02", key, time.UTC)
		klines = append(klines, provider.DailyKline{
			Timestamp: t,
			Open:      pv.price,
			High:      pv.price,
			Low:       pv.price,
			Close:     pv.price,
			Volume:    pv.vol,
		})
	}
	return klines, nil
}
