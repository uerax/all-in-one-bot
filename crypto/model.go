package crypto

type priceResp struct {
	Price  string `json:"price"`
	Symbol string `json:"symbol"`
}

type Meme struct {
	Pairs []*Pair `json:"pairs"`
}

type Pair struct {
	URL         string       `json:"url"`
	ChainId     string       `json:"chainId"`
	PairAddress string       `json:"pairAddress"`
	PriceNative string       `json:"priceNative"`
	PriceUsd    string       `json:"priceUsd"`
	BaseToken   *BaseToken   `json:"baseToken"`
	PriceChange *PriceChange `json:"priceChange"`
}

type BaseToken struct {
	Addr   string `json:"address"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type PriceChange struct {
	M5  float64 `json:"m5"`
	H1  float64 `json:"h1"`
	H6  float64 `json:"h6"`
	H24 float64 `json:"h24"`
}