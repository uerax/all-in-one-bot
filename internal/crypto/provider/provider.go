package provider

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("provider: asset or pool not found")
	ErrRateLimited   = errors.New("provider: rate limit exceeded")
	ErrInvalidParams = errors.New("provider: invalid parameters")
)

type ProviderType string

const (
	ProviderTypeCEX ProviderType = "cex"
	ProviderTypeDEX ProviderType = "dex"
)

// MarketData represents standardized cryptocurrency price and market metrics.
type MarketData struct {
	Symbol          string       `json:"symbol"`
	Name            string       `json:"name"`
	PriceUSD        float64      `json:"price_usd"`
	Change5m        float64      `json:"change_5m,omitempty"`
	Change1h        float64      `json:"change_1h,omitempty"`
	Change6h        float64      `json:"change_6h,omitempty"`
	Change24h       float64      `json:"change_24h"`
	Volume24hUSD    float64      `json:"volume_24h_usd"`
	ReserveUSD      float64      `json:"reserve_usd,omitempty"` // DEX Liquidity
	FDV             float64      `json:"fdv,omitempty"`
	MarketCap       float64      `json:"market_cap,omitempty"`
	MarketCapRank   int          `json:"market_cap_rank,omitempty"`
	Chain           string       `json:"chain,omitempty"` // Network (e.g. eth, solana, base)
	DEX             string       `json:"dex,omitempty"`   // DEX platform name (e.g. Uniswap v3, Raydium)
	ContractAddress string       `json:"contract_address,omitempty"`
	PoolAddress     string       `json:"pool_address,omitempty"`
	Source          string       `json:"source"`        // Provider name (e.g. "coingecko", "geckoterminal")
	SourceType      ProviderType `json:"source_type"`   // ProviderTypeCEX or ProviderTypeDEX
	UpdatedAt       time.Time    `json:"updated_at"`
}

// SearchResult represents unified search query results.
type SearchResult struct {
	ID              string       `json:"id"`
	Symbol          string       `json:"symbol"`
	Name            string       `json:"name"`
	Chain           string       `json:"chain,omitempty"`
	PoolAddress     string       `json:"pool_address,omitempty"`
	ContractAddress string       `json:"contract_address,omitempty"`
	PriceUSD        float64      `json:"price_usd,omitempty"`
	Volume24hUSD    float64      `json:"volume_24h_usd,omitempty"`
	ReserveUSD      float64      `json:"reserve_usd,omitempty"`
	MarketCapRank   int          `json:"market_cap_rank,omitempty"`
	Source          string       `json:"source"`
	SourceType      ProviderType `json:"source_type"`
}

// DailyKline represents a 1-day candlestick record for volume scanning and analysis.
type DailyKline struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
}

// Provider identifies the underlying crypto data provider.
type Provider interface {
	Name() string
	Type() ProviderType
}

// PriceProvider defines single and multi-asset price & market data retrieval.
type PriceProvider interface {
	Provider
	GetPrice(query string) (*MarketData, error)
	GetMultiPrices(queries []string) (map[string]*MarketData, error)
}

// CEXSearchProvider defines search interface for CEX / listed market assets.
type CEXSearchProvider interface {
	Provider
	SearchCEX(query string) ([]SearchResult, error)
}

// DEXSearchProvider defines search interface for DEX pools and on-chain contract addresses.
type DEXSearchProvider interface {
	Provider
	SearchDEX(query string) ([]SearchResult, error)
}

// SearchProvider represents a provider capable of general search.
type SearchProvider interface {
	Provider
	Search(query string) ([]SearchResult, error)
}

// KlineProvider defines daily K-line fetching for historical volume & price monitoring.
type KlineProvider interface {
	Provider
	GetDailyKline(query string) ([]DailyKline, error)
}

// TrendingProvider defines DEX / CEX trending pool / asset retrieval.
type TrendingProvider interface {
	Provider
	GetTrending(network string) ([]MarketData, error)
}
