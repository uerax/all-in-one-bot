package crypto

import (
	"encoding/json"
	"io"
	"net/http"
)

type TokenTxs struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Txs     []Txs  `json:"result"`
}

type Txs struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
}

var chains = map[string]string{
	"eth":     "1",
	"bsc":     "56",
	"polygon": "137",
	"base":    "8453",
}

type Chainscan struct {
	C    chan string
	Url  string
	Keys *PollingKeyV2
}

func NewScan() *Chainscan {
	return &Chainscan{
		C:    make(chan string, 10),
		Url:  "https://api.etherscan.io/v2/api",
		Keys: NewPollingKeyV2(),
	}
}

func (t *Chainscan) Profit(cttAddr, tokenAddr, chain string) {
	t.Tokentx(cttAddr, tokenAddr, chain)

}

func (t *Chainscan) Tokentx(cttAddr, tokenAddr, chain string) []Txs {
	url := t.Url + "?chainid=" + chains[chain] +
		"&module=account" +
		"&action=tokentx" +
		"&contractaddress=" + cttAddr +
		"&address=" + tokenAddr +
		"&apikey=" + t.Keys.GetKey()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.C <- err.Error()
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.C <- "Error: " + resp.Status
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.C <- err.Error()
		return nil
	}

	txs := &TokenTxs{}
	if err := json.Unmarshal(body, txs); err != nil {
		t.C <- "Error parsing response: " + err.Error()
		return nil
	}

	return txs.Txs
}

type TransRec struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Receipt Receipt `json:"result"`
}

type Receipt struct {
	BlockHash         string      `json:"blockHash"`
	BlockNumber       string      `json:"blockNumber"`
	ContractAddress   interface{} `json:"contractAddress"`
	CumulativeGasUsed string      `json:"cumulativeGasUsed"`
	EffectiveGasPrice string      `json:"effectiveGasPrice"`
	From              string      `json:"from"`
	GasUsed           string      `json:"gasUsed"`
	Logs              []ReceiptLog       `json:"logs"`
	LogsBloom         string      `json:"logsBloom"`
	Status            string      `json:"status"`
	To                string      `json:"to"`
	TransactionHash   string      `json:"transactionHash"`
	TransactionIndex  string      `json:"transactionIndex"`
	Type              string      `json:"type"`
}

type ReceiptLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}



func (t *Chainscan) GetTransactionReceipt(txHash, chain string) []ReceiptLog {
	url := t.Url + "?chainid=" + chains[chain] +
		"&module=proxy" +
		"&action=eth_getTransactionReceipt" +
		"&txhash=" + txHash +
		"&apikey=" + t.Keys.GetKey()

	req, _ := http.NewRequest(http.MethodGet, url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.C <- err.Error()
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.C <- "Error: " + resp.Status
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.C <- err.Error()
		return nil
	}

	tr := &TransRec{}
	if err := json.Unmarshal(body, tr); err != nil {
		t.C <- "Error parsing response: " + err.Error()
		return nil
	}

	return tr.Receipt.Logs
}

