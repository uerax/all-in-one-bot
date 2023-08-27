package crypto

import "sync"

type txResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []tx   `json:"result"`
}

type tx struct {
	Value string `json:"value"`
	From  string `json:"from"`
	To    string `json:"To"`
}

type txs struct {
	Buy     float64
	Sell    float64
	Symbol  string
	Profit  float64
	Scam    string
	Pay     float64
	Time    string
	Addr    string
	Ts      int64
	Tx      int
	TotalTx uint32
	WinTx   uint32
	Mu      sync.Mutex
}

type TxProfitRate struct {
	Ts           int64
	Rate         float64
	Earnable     bool
	Quality      bool
	Symbol       string
	Addr         string
	Scam         string
	EarnableScam bool
	QualityScam  bool
	Old			 string
}
