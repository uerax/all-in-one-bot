package crypto

type priceResp struct {
	Price  string `json:"price"`
	Symbol string `json:"symbol"`
}

type Meme struct {
	Pairs []*Pair `json:"pairs"`
}

type Pair struct {
	URL         string     `json:"url"`
	ChainId     string     `json:"chainId"`
	PairAddress string     `json:"pairAddress"`
	PriceNative string     `json:"priceNative"`
	PriceUsd    string     `json:"priceUsd"`
	BaseToken   *BaseToken `json:"baseToken"`
}

type BaseToken struct {
	Addr   string `json:"address"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}