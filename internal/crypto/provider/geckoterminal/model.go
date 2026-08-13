package geckoterminal

import "strconv"

// PoolResponse represents JSON API response for GeckoTerminal pool endpoint.
type PoolResponse struct {
	Data     PoolData       `json:"data"`
	Included []IncludedItem `json:"included,omitempty"`
}

// PoolListResponse represents JSON API response for GeckoTerminal pool search & trending endpoints.
type PoolListResponse struct {
	Data     []PoolData     `json:"data"`
	Included []IncludedItem `json:"included,omitempty"`
}

type PoolData struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    PoolAttributes    `json:"attributes"`
	Relationships PoolRelationships `json:"relationships"`
}

type PoolAttributes struct {
	Address               string            `json:"address"`
	Name                  string            `json:"name"`
	BaseTokenPriceUSD     string            `json:"base_token_price_usd"`
	QuoteTokenPriceUSD    string            `json:"quote_token_price_usd"`
	FDVUSD                string            `json:"fdv_usd"`
	MarketCapUSD          string            `json:"market_cap_usd"`
	ReserveInUSD          string            `json:"reserve_in_usd"`
	VolumeUSD             VolumeUSDMap      `json:"volume_usd"`
	PriceChangePercentage PriceChangePctMap `json:"price_change_percentage"`
	PoolCreatedAt         string            `json:"pool_created_at"`
}

type VolumeUSDMap struct {
	M5  string `json:"m5"`
	H1  string `json:"h1"`
	H6  string `json:"h6"`
	H24 string `json:"h24"`
}

type PriceChangePctMap struct {
	M5  string `json:"m5"`
	H1  string `json:"h1"`
	H6  string `json:"h6"`
	H24 string `json:"h24"`
}

type PoolRelationships struct {
	DEX       RelationshipItem `json:"dex"`
	Network   RelationshipItem `json:"network"`
	BaseToken RelationshipItem `json:"base_token"`
}

type RelationshipItem struct {
	Data RelationshipData `json:"data"`
}

type RelationshipData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type IncludedItem struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes IncludedItemAttributes `json:"attributes"`
}

type IncludedItemAttributes struct {
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
	Address string `json:"address"`
}

// TokenResponse represents JSON API response for GeckoTerminal token endpoint.
type TokenResponse struct {
	Data TokenData `json:"data"`
}

type TokenData struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes TokenAttributes `json:"attributes"`
}

type TokenAttributes struct {
	Address         string `json:"address"`
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	PriceUSD        string `json:"price_usd"`
	FDVUSD          string `json:"fdv_usd"`
	TotalReserveUSD string `json:"total_reserve_in_usd"`
	VolumeUSD       string `json:"volume_usd"`
}

// OHLCVResponse represents K-line response from GeckoTerminal.
type OHLCVResponse struct {
	Data OHLCVData `json:"data"`
}

type OHLCVData struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes OHLCVAttributes `json:"attributes"`
}

type OHLCVAttributes struct {
	List [][]float64 `json:"ohlcv_list"`
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0.0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return v
}
