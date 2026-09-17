package provider

import (
	"errors"
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/cache"
)

type mockPriceProvider struct {
	getPriceFn func(query string) (*MarketData, error)
	callCount  int
}

func (m *mockPriceProvider) Name() string       { return "mock_price" }
func (m *mockPriceProvider) Type() ProviderType { return ProviderTypeCEX }
func (m *mockPriceProvider) GetPrice(query string) (*MarketData, error) {
	m.callCount++
	if m.getPriceFn != nil {
		return m.getPriceFn(query)
	}
	return &MarketData{Symbol: "BTC", PriceUSD: 60000}, nil
}
func (m *mockPriceProvider) GetMultiPrices(queries []string) (map[string]*MarketData, error) {
	return nil, nil
}

type mockSearchProvider struct {
	searchFn  func(query string) ([]SearchResult, error)
	callCount int
}

func (m *mockSearchProvider) Name() string       { return "mock_search" }
func (m *mockSearchProvider) Type() ProviderType { return ProviderTypeCEX }
func (m *mockSearchProvider) SearchCEX(query string) ([]SearchResult, error) {
	m.callCount++
	if m.searchFn != nil {
		return m.searchFn(query)
	}
	return []SearchResult{{ID: "btc", Symbol: "BTC"}}, nil
}
func (m *mockSearchProvider) SearchDEX(query string) ([]SearchResult, error) {
	m.callCount++
	if m.searchFn != nil {
		return m.searchFn(query)
	}
	return []SearchResult{{ID: "btc_dex", Symbol: "BTC"}}, nil
}

type mockKlineProvider struct {
	callCount int
}

func (m *mockKlineProvider) Name() string       { return "mock_kline" }
func (m *mockKlineProvider) Type() ProviderType { return ProviderTypeDEX }
func (m *mockKlineProvider) GetDailyKline(query string) ([]DailyKline, error) {
	m.callCount++
	return []DailyKline{{Close: 100, Volume: 500}}, nil
}

func TestManager_GetPriceCache(t *testing.T) {
	c := cache.NewCache()
	defer c.Close()

	cgMock := &mockPriceProvider{}
	gtMock := &mockPriceProvider{}

	mgr := NewManager(cgMock, nil, nil, gtMock, nil, nil, nil, c)

	// 1. 首次查询，触发底层调用
	data1, err := mgr.GetPrice("btc")
	if err != nil || data1 == nil || data1.Symbol != "BTC" {
		t.Fatalf("first GetPrice failed: %v", err)
	}
	if cgMock.callCount != 1 {
		t.Errorf("expected cgMock callCount=1, got %d", cgMock.callCount)
	}

	// 2. 60秒内第二次查询，必须直接命中缓存
	data2, err := mgr.GetPrice("btc")
	if err != nil || data2 == nil || data2.Symbol != "BTC" {
		t.Fatalf("second GetPrice failed: %v", err)
	}
	if cgMock.callCount != 1 {
		t.Errorf("expected cgMock callCount to stay 1 due to cache hit, got %d", cgMock.callCount)
	}

	// 大小写不敏感测试
	data3, err := mgr.GetPrice("BTC")
	if err != nil || data3 == nil {
		t.Fatalf("case-insensitive GetPrice failed: %v", err)
	}
	if cgMock.callCount != 1 {
		t.Errorf("expected case-insensitive match to hit cache, got callCount=%d", cgMock.callCount)
	}
}

func TestManager_SearchCache(t *testing.T) {
	c := cache.NewCache()
	defer c.Close()

	cgSearch := &mockSearchProvider{}
	gtSearch := &mockSearchProvider{}

	mgr := NewManager(nil, cgSearch, nil, nil, gtSearch, nil, nil, c)

	// 1. 首次搜索
	res1, err := mgr.Search("ethereum")
	if err != nil || res1 == nil {
		t.Fatalf("first Search failed: %v", err)
	}
	if cgSearch.callCount != 1 || gtSearch.callCount != 1 {
		t.Errorf("expected cgSearch=1 and gtSearch=1, got cg=%d gt=%d", cgSearch.callCount, gtSearch.callCount)
	}

	// 2. 15分钟内第二次搜索，直接命中缓存
	res2, err := mgr.Search("ethereum")
	if err != nil || res2 == nil {
		t.Fatalf("second Search failed: %v", err)
	}
	if cgSearch.callCount != 1 || gtSearch.callCount != 1 {
		t.Errorf("expected search to hit cache, got cg=%d gt=%d", cgSearch.callCount, gtSearch.callCount)
	}
}

func TestManager_CoinGeckoRateLimitNoFallback(t *testing.T) {
	c := cache.NewCache()
	defer c.Close()

	cgMock := &mockPriceProvider{
		getPriceFn: func(query string) (*MarketData, error) {
			return nil, errors.New("coingecko: http 429")
		},
	}
	gtMock := &mockPriceProvider{
		getPriceFn: func(query string) (*MarketData, error) {
			t.Error("gtMock should NOT be called when CEX failed with 429 rate limit")
			return &MarketData{Symbol: "UNRELATED"}, nil
		},
	}

	mgr := NewManager(cgMock, nil, nil, gtMock, nil, nil, nil, c)

	_, err := mgr.GetPrice("qubic-network")
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited when CEX 429, got %v", err)
	}
	if gtMock.callCount != 0 {
		t.Errorf("expected gtMock callCount=0, got %d", gtMock.callCount)
	}
}

func TestManager_GetDailyKlineBypassCache(t *testing.T) {
	c := cache.NewCache()
	defer c.Close()

	klineMock := &mockKlineProvider{}
	mgr := NewManager(nil, nil, nil, nil, nil, klineMock, nil, c)

	// 连续调用两次，kline 必须保持直通不走长效模糊缓存
	_, err := mgr.GetDailyKline("base:0x123")
	if err != nil {
		t.Fatalf("first GetDailyKline failed: %v", err)
	}
	_, err = mgr.GetDailyKline("base:0x123")
	if err != nil {
		t.Fatalf("second GetDailyKline failed: %v", err)
	}

	if klineMock.callCount != 2 {
		t.Errorf("expected GetDailyKline to bypass cache and call provider each time, got %d", klineMock.callCount)
	}
}
