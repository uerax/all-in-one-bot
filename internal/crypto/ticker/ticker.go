package ticker

type AssetQuery struct {
    Symbol   string // 币名，如 BTC
    ID       string // 平台 ID，如 Coingecko 的 "bitcoin"
    Address  string // 链上地址
    Chain    string // 链名，如 "eth", "bsc"
}

type Ticker interface {
	Price(coin AssetQuery) (float64, error)
}

