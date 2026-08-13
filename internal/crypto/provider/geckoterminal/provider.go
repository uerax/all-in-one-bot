package geckoterminal

import (
	"regexp"
	"strings"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

var (
	evmAddrRegex    = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	solanaAddrRegex = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)
)

type Provider struct {
	client *Client
	log    logger.Log
}

func NewProvider(baseURL string, timeoutSec int, log logger.Log) *Provider {
	return &Provider{
		client: NewClient(baseURL, timeoutSec, log),
		log:    log,
	}
}

func (p *Provider) Name() string {
	return "geckoterminal"
}

func (p *Provider) Type() provider.ProviderType {
	return provider.ProviderTypeDEX
}

// GetPrice fetches DEX market metrics for a token or pool address (e.g., "eth:0x...", "0x...", "solana:address", or ticker).
func (p *Provider) GetPrice(query string) (*provider.MarketData, error) {
	network, addr, isPool := p.parseQuery(query)

	if addr != "" {
		if isPool {
			poolResp, err := p.client.GetPool(network, addr)
			if err == nil && poolResp != nil && poolResp.Data.ID != "" {
				return p.mapPoolDataToMarketData(&poolResp.Data, poolResp.Included), nil
			}
		}

		// Try token pools
		poolsResp, err := p.client.GetTokenPools(network, addr)
		if err == nil && poolsResp != nil && len(poolsResp.Data) > 0 {
			return p.mapPoolDataToMarketData(&poolsResp.Data[0], poolsResp.Included), nil
		}
	}

	// Fallback to pool search
	searchResp, err := p.client.SearchPools(query)
	if err != nil {
		return nil, err
	}
	if searchResp == nil || len(searchResp.Data) == 0 {
		return nil, provider.ErrNotFound
	}

	return p.mapPoolDataToMarketData(&searchResp.Data[0], searchResp.Included), nil
}

// GetMultiPrices retrieves price data for multiple token/pool queries.
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

// SearchDEX performs search explicitly for DEX pools.
func (p *Provider) SearchDEX(query string) ([]provider.SearchResult, error) {
	return p.Search(query)
}

// Search performs DEX pool searching across networks.
func (p *Provider) Search(query string) ([]provider.SearchResult, error) {
	searchResp, err := p.client.SearchPools(query)
	if err != nil {
		return nil, err
	}
	if searchResp == nil || len(searchResp.Data) == 0 {
		return []provider.SearchResult{}, nil
	}

	var results []provider.SearchResult
	includedMap := makeIncludedMap(searchResp.Included)

	for _, pool := range searchResp.Data {
		baseTokenID := pool.Relationships.BaseToken.Data.ID
		tokenInfo := includedMap[baseTokenID]

		symbol := tokenInfo.Symbol
		name := tokenInfo.Name
		if symbol == "" {
			symbol = pool.Attributes.Name
		}
		if name == "" {
			name = pool.Attributes.Name
		}

		results = append(results, provider.SearchResult{
			ID:              pool.ID,
			Symbol:          symbol,
			Name:            name,
			Chain:           pool.Relationships.Network.Data.ID,
			PoolAddress:     pool.Attributes.Address,
			ContractAddress: tokenInfo.Address,
			PriceUSD:        parseFloat(pool.Attributes.BaseTokenPriceUSD),
			Volume24hUSD:    parseFloat(pool.Attributes.VolumeUSD.H24),
			ReserveUSD:      parseFloat(pool.Attributes.ReserveInUSD),
			Source:          p.Name(),
			SourceType:      p.Type(),
		})
	}

	return results, nil
}

// GetDailyKline fetches daily OHLCV K-line data for a DEX pool.
func (p *Provider) GetDailyKline(query string) ([]provider.DailyKline, error) {
	network, addr, _ := p.parseQuery(query)

	if addr == "" {
		// Search pool first to get network & address
		searchResp, err := p.client.SearchPools(query)
		if err != nil || searchResp == nil || len(searchResp.Data) == 0 {
			return nil, provider.ErrNotFound
		}
		network = searchResp.Data[0].Relationships.Network.Data.ID
		addr = searchResp.Data[0].Attributes.Address
	}

	ohlcvResp, err := p.client.GetPoolOHLCV(network, addr, "day", 100)
	if err != nil {
		return nil, err
	}
	if ohlcvResp == nil || len(ohlcvResp.Data.Attributes.List) == 0 {
		return nil, provider.ErrNotFound
	}

	var klines []provider.DailyKline
	// GeckoTerminal OHLCV format: [timestamp, open, high, low, close, volume]
	for _, row := range ohlcvResp.Data.Attributes.List {
		if len(row) < 6 {
			continue
		}
		ts := time.Unix(int64(row[0]), 0)
		klines = append(klines, provider.DailyKline{
			Timestamp: ts,
			Open:      row[1],
			High:      row[2],
			Low:       row[3],
			Close:     row[4],
			Volume:    row[5],
		})
	}

	return klines, nil
}

// GetTrending fetches top trending DEX pools on a specified network or globally.
func (p *Provider) GetTrending(network string) ([]provider.MarketData, error) {
	trendingResp, err := p.client.GetTrendingPools(network)
	if err != nil {
		return nil, err
	}
	if trendingResp == nil || len(trendingResp.Data) == 0 {
		return []provider.MarketData{}, nil
	}

	var list []provider.MarketData
	for i := range trendingResp.Data {
		list = append(list, *p.mapPoolDataToMarketData(&trendingResp.Data[i], trendingResp.Included))
	}
	return list, nil
}

func (p *Provider) parseQuery(query string) (network string, address string, isPool bool) {
	query = strings.TrimSpace(query)
	parts := strings.SplitN(query, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], false
	}

	if evmAddrRegex.MatchString(query) {
		return "eth", query, false
	}
	if solanaAddrRegex.MatchString(query) && !strings.Contains(query, " ") {
		return "solana", query, false
	}

	return "", "", false
}

func (p *Provider) mapPoolDataToMarketData(pool *PoolData, included []IncludedItem) *provider.MarketData {
	includedMap := makeIncludedMap(included)
	baseTokenID := pool.Relationships.BaseToken.Data.ID
	tokenInfo := includedMap[baseTokenID]

	symbol := tokenInfo.Symbol
	name := tokenInfo.Name
	if symbol == "" {
		symbol = pool.Attributes.Name
	}
	if name == "" {
		name = pool.Attributes.Name
	}

	dexID := pool.Relationships.DEX.Data.ID
	networkID := pool.Relationships.Network.Data.ID

	return &provider.MarketData{
		Symbol:          symbol,
		Name:            name,
		PriceUSD:        parseFloat(pool.Attributes.BaseTokenPriceUSD),
		Change5m:        parseFloat(pool.Attributes.PriceChangePercentage.M5),
		Change1h:        parseFloat(pool.Attributes.PriceChangePercentage.H1),
		Change6h:        parseFloat(pool.Attributes.PriceChangePercentage.H6),
		Change24h:       parseFloat(pool.Attributes.PriceChangePercentage.H24),
		Volume24hUSD:    parseFloat(pool.Attributes.VolumeUSD.H24),
		ReserveUSD:      parseFloat(pool.Attributes.ReserveInUSD),
		FDV:             parseFloat(pool.Attributes.FDVUSD),
		MarketCap:       parseFloat(pool.Attributes.MarketCapUSD),
		Chain:           networkID,
		DEX:             dexID,
		ContractAddress: tokenInfo.Address,
		PoolAddress:     pool.Attributes.Address,
		Source:          p.Name(),
		SourceType:      p.Type(),
		UpdatedAt:       time.Now(),
	}
}

func makeIncludedMap(included []IncludedItem) map[string]IncludedItemAttributes {
	m := make(map[string]IncludedItemAttributes)
	for _, item := range included {
		m[item.ID] = item.Attributes
	}
	return m
}
