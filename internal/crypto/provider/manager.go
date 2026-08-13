package provider

import (
	"regexp"
	"strings"
	"sync"
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
}

func NewManager(cgPrice PriceProvider, cgSearch CEXSearchProvider, cgKline KlineProvider, gtPrice PriceProvider, gtSearch DEXSearchProvider, gtKline KlineProvider, gtTrending TrendingProvider) *Manager {
	return &Manager{
		cgPrice:    cgPrice,
		cgSearch:   cgSearch,
		cgKline:    cgKline,
		gtPrice:    gtPrice,
		gtSearch:   gtSearch,
		gtKline:    gtKline,
		gtTrending: gtTrending,
	}
}

// GetPrice fetches market data by query, automatically routing address queries to DEX provider
// and ticker/symbol queries to CEX provider with DEX fallback.
func (m *Manager) GetPrice(query string) (*MarketData, error) {
	q := strings.TrimSpace(query)
	if isContractOrPool(q) {
		if m.gtPrice != nil {
			return m.gtPrice.GetPrice(q)
		}
	}

	// Try CEX provider first for standard symbols
	if m.cgPrice != nil {
		data, err := m.cgPrice.GetPrice(q)
		if err == nil && data != nil {
			return data, nil
		}
	}

	// Fallback to DEX provider
	if m.gtPrice != nil {
		return m.gtPrice.GetPrice(q)
	}

	return nil, ErrNotFound
}

// MultiSearchResult represents combined CEX and DEX search results.
type MultiSearchResult struct {
	CEXResults []SearchResult `json:"cex_results"`
	DEXResults []SearchResult `json:"dex_results"`
}

// Search performs concurrent search across CEX and DEX providers.
func (m *Manager) Search(query string) (*MultiSearchResult, error) {
	res := &MultiSearchResult{
		CEXResults: []SearchResult{},
		DEXResults: []SearchResult{},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	q := strings.TrimSpace(query)

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
	return res, nil
}

// GetDailyKline fetches daily K-line data for Crocodile scanner, prioritizing DEX provider (GeckoTerminal) for accurate volume & price metrics, with CEX fallback.
func (m *Manager) GetDailyKline(query string) ([]DailyKline, error) {
	q := strings.TrimSpace(query)

	// Try DEX provider (GeckoTerminal) first for accurate on-chain DEX volume/OHLCV data
	if m.gtKline != nil {
		klines, err := m.gtKline.GetDailyKline(q)
		if err == nil && len(klines) > 0 {
			return klines, nil
		}
	}

	// Fallback to CEX provider (CoinGecko)
	if m.cgKline != nil {
		return m.cgKline.GetDailyKline(q)
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
