package provider

import (
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/cache"
)

var (
	evmRegex    = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	solanaRegex = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)
)

type Manager struct {
	cgPrice    PriceProvider
	cgSearch   CEXSearchProvider
	cgKline    KlineProvider
	gtPrice    PriceProvider
	gtSearch   DEXSearchProvider
	gtKline    KlineProvider
	gtTrending TrendingProvider
	cache      cache.Cache
}

func NewManager(cgPrice PriceProvider, cgSearch CEXSearchProvider, cgKline KlineProvider, gtPrice PriceProvider, gtSearch DEXSearchProvider, gtKline KlineProvider, gtTrending TrendingProvider, cacher cache.Cache) *Manager {
	return &Manager{
		cgPrice:    cgPrice,
		cgSearch:   cgSearch,
		cgKline:    cgKline,
		gtPrice:    gtPrice,
		gtSearch:   gtSearch,
		gtKline:    gtKline,
		gtTrending: gtTrending,
		cache:      cacher,
	}
}

// GetPrice fetches market data by query, automatically routing address queries to DEX provider
// and ticker/symbol queries to CEX provider with DEX fallback, backed by a 60s cache.
func (m *Manager) GetPrice(query string) (*MarketData, error) {
	q := strings.TrimSpace(query)
	cacheKey := "price:" + strings.ToLower(q)

	if m.cache != nil {
		if val, ok := m.cache.Get(cacheKey); ok {
			if md, ok := val.(*MarketData); ok {
				return md, nil
			}
		}
	}

	if isContractOrPool(q) {
		if m.gtPrice != nil {
			data, err := m.gtPrice.GetPrice(q)
			if err == nil && data != nil {
				if m.cache != nil {
					m.cache.Set(cacheKey, data, 60*time.Second)
				}
				return data, nil
			}
			return nil, err
		}
	}

	// Try CEX provider first for standard symbols
	if m.cgPrice != nil {
		data, err := m.cgPrice.GetPrice(q)
		if err == nil && data != nil {
			if m.cache != nil {
				m.cache.Set(cacheKey, data, 60*time.Second)
			}
			return data, nil
		}
		// 防呆防护：若 CoinGecko 因 429 限流失败，禁止盲目 fallback 到 DEX 搜索生成脏数据，直接上报限流
		if err != nil && (errors.Is(err, ErrRateLimited) || strings.Contains(strings.ToLower(err.Error()), "429")) {
			return nil, ErrRateLimited
		}
	}

	// Fallback to DEX provider (only when CEX returned not found)
	if m.gtPrice != nil {
		data, err := m.gtPrice.GetPrice(q)
		if err == nil && data != nil {
			if m.cache != nil {
				m.cache.Set(cacheKey, data, 60*time.Second)
			}
			return data, nil
		}
		return nil, err
	}

	return nil, ErrNotFound
}

// MultiSearchResult represents combined CEX and DEX search results.
type MultiSearchResult struct {
	CEXResults []SearchResult `json:"cex_results"`
	DEXResults []SearchResult `json:"dex_results"`
}

// Search performs concurrent search across CEX and DEX providers, backed by a 15m cache.
func (m *Manager) Search(query string) (*MultiSearchResult, error) {
	q := strings.TrimSpace(query)
	cacheKey := "search:" + strings.ToLower(q)

	if m.cache != nil {
		if val, ok := m.cache.Get(cacheKey); ok {
			if res, ok := val.(*MultiSearchResult); ok {
				return res, nil
			}
		}
	}

	res := &MultiSearchResult{
		CEXResults: []SearchResult{},
		DEXResults: []SearchResult{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	if m.cgSearch != nil && !isContractOrPool(q) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cRes, err := m.cgSearch.SearchCEX(q)
			if err == nil {
				mu.Lock()
				if len(cRes) > 5 {
					res.CEXResults = cRes[:5]
				} else {
					res.CEXResults = cRes
				}
				mu.Unlock()
			}
		}()
	}

	if m.gtSearch != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dRes, err := m.gtSearch.SearchDEX(q)
			if err == nil {
				mu.Lock()
				if len(dRes) > 5 {
					res.DEXResults = dRes[:5]
				} else {
					res.DEXResults = dRes
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if m.cache != nil && (len(res.CEXResults) > 0 || len(res.DEXResults) > 0) {
		m.cache.Set(cacheKey, res, 15*time.Minute)
	}

	return res, nil
}

// GetDailyKline fetches daily K-line data for Crocodile scanner strictly from DEX provider (GeckoTerminal).
func (m *Manager) GetDailyKline(query string) ([]DailyKline, error) {
	q := strings.TrimSpace(query)

	if m.gtKline != nil {
		return m.gtKline.GetDailyKline(q)
	}

	return nil, ErrNotFound
}

// GetTrending fetches top DEX trending pools on a network.
func (m *Manager) GetTrending(network string) ([]MarketData, error) {
	if m.gtTrending != nil {
		return m.gtTrending.GetTrending(network)
	}
	return nil, ErrNotFound
}

func isContractOrPool(query string) bool {
	if strings.Contains(query, ":") {
		return true
	}
	if evmRegex.MatchString(query) {
		return true
	}
	if solanaRegex.MatchString(query) && !strings.Contains(query, " ") {
		return true
	}
	return false
}
