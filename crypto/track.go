package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/uerax/goconf"
)

type Track struct {
	C      chan string
	Newest map[string]string
	apiKey string
	Task   map[string]context.CancelFunc
	api    *Crypto
}

type txResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []tx   `json:"result"`
}

type tx struct {
	Value string `json:"value"`
}

func NewTrack() *Track {
	return &Track{
		C:      make(chan string, 5),
		Newest: make(map[string]string),
		apiKey: goconf.VarStringOrDefault("", "crypto", "etherscan", "apiKey"),
		Task:   make(map[string]context.CancelFunc),
		api:    NewCrypto("", ""),
	}
}

func (t *Track) CronTracking(addr string) {
	if _, ok := t.Task[addr]; !ok {
		ctx, cf := context.WithCancel(context.Background())
		t.Task[addr] = cf
		go t.Tracking(addr, ctx)
		t.C <- "开始追踪" + addr
	}
}

func (t *Track) StopTracking(addr string) {
	if v, ok := t.Task[addr]; ok {
		v()
		delete(t.Task, addr)
		t.C <- "已停止追踪" + addr
	}
}

func (t *Track) Tracking(addr string, ctx context.Context) {
	tick := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			go t.WalletTracking(addr)
		}
	}
}

func (t *Track) TrackingList(tip bool) string {
	var sb strings.Builder
	for k := range t.Task {
		sb.WriteString("\n`")
		sb.WriteString(k)
		sb.WriteString("`")
	}
	if !tip {
		t.C <- "*当前正在追踪的地址有:*" + sb.String()
	}

	return "*当前正在追踪的地址有:*" + sb.String()

}

func (t *Track) WalletTracking(addr string) {
	if t.apiKey == "" {
		t.C <- "未读取到etherscan的apikey无法启动监控"
		return
	}
	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=%s&sort=desc&address=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, "30", addr, t.apiKey))
	if err != nil {
		fmt.Println("请求失败")
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("读取body失败")
		return
	}
	scan := new(TokenTxResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		fmt.Println("json转换失败")
		return
	}

	if scan.Status != "1" {
		return
	}

	if len(scan.Result) == 0 || strings.EqualFold(scan.Result[0].Hash, t.Newest[addr]) {
		return
	}

	t.Newest[addr] = scan.Result[0].Hash

	sb := strings.Builder{}

	for _, record := range scan.Result {
		if record.TokenSymbol != "WETH" {
			balance := t.getEthByHash(record.Hash)
			if balance == "" {
				continue
			}
			if strings.EqualFold(record.From, addr) {
				sb.WriteString("\n*Sell: *")
			} else {
				sb.WriteString("\n*Buy: *")
			}
			sb.WriteString("[")
			sb.WriteString(record.TokenSymbol)
			sb.WriteString("](https://www.dextools.io/app/cn/ether/pair-explorer/")
			sb.WriteString(record.ContractAddress)
			sb.WriteString("): ")
			sb.WriteString(balance[:10])
			sb.WriteString(" ETH (")
			i, err := strconv.ParseInt(record.TimeStamp, 10, 64)
			if err == nil {
				sb.WriteString(time.Unix(i, 0).Format("01-02 15:04:05"))
			}
			sb.WriteString(")\n`")
			sb.WriteString(record.ContractAddress)
			sb.WriteString("`")
		}
	}

	if sb.Len() > 0 {
		t.C <- "监控地址执行操作:\n" + sb.String()
	}

}

func (t *Track) getEthByHash(hash string) string {
	url := "https://api.etherscan.io/api?module=account&action=txlistinternal&txhash=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, hash, t.apiKey))
	if err != nil {
		fmt.Println("请求失败")
		return ""
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("读取body失败")
		return ""
	}
	scan := new(txResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		fmt.Println("json转换失败")
		return ""
	}

	if scan.Status != "1" || len(scan.Result) == 0 {
		return ""
	}
	val := scan.Result[0].Value
	if len(val) > 18 {
		return val[:len(val)-18] + "." + val[len(val)-18:]
	} else if len(val) == 18 {
		return "0." + val
	} else {
		return "0." + strings.Repeat("0", 18-len(val)) + val
	}

}

func (t *Track) WalletTxAnalyze(addr string, offset string) {
	if t.apiKey == "" {
		t.C <- "未读取到etherscan的apikey无法调用api"
		return
	}
	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=%s&sort=desc&address=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, offset, addr, t.apiKey))
	if err != nil {
		fmt.Println("etherscan请求失败")
		t.C <- "etherscan请求失败"
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("读取body失败")
		t.C <- "读取body失败"
		return
	}
	scan := new(TokenTxResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		fmt.Println("json转换失败")
		t.C <- "json转换失败"
		return
	}

	if scan.Status != "1" {
		t.C <- "响应码异常"
		return
	}

	type txs struct {
		Buy    float64
		Sell   float64
		Symbol string
		Profit float64
		Scam   string
	}
	profit := 0.0
	detail := make(map[string]*txs)

	for _, record := range scan.Result {
		if record.TokenSymbol != "WETH" {
			balance := t.getEthByHash(record.Hash)
			if balance == "" {
				continue
			}
			val, err := strconv.ParseFloat(balance[:10], 64)
			if err == nil {
				if strings.EqualFold(record.From, addr) {
					profit += val
				} else {
					profit -= val
				}
			}

			if record.Decimal != "" {
				dec, err := strconv.Atoi(record.Decimal)
				l := len(record.Value) - dec
				if err == nil {
					tmp := ""
					if l <= 0 {
						tmp = "0." + strings.Repeat("0", -l) + record.Value
					} else {
						tmp = record.Value[:l]
					}
					cnt, err := strconv.ParseFloat(tmp, 64)
					if err == nil {
						if _, ok := detail[record.ContractAddress]; !ok {
							detail[record.ContractAddress] = new(txs)
						}
						if strings.EqualFold(record.From, addr) {
							detail[record.ContractAddress].Sell += cnt
							detail[record.ContractAddress].Profit += val
						} else {
							detail[record.ContractAddress].Buy += cnt
							detail[record.ContractAddress].Profit -= val
						}
						detail[record.ContractAddress].Symbol = record.TokenSymbol
						isHoneypot := t.api.WhetherHoneypot(record.ContractAddress)
						if isHoneypot {
							detail[record.ContractAddress].Scam = "*[SCAM]*"
						}
					}
				}
			}

		}
	}

	msg := fmt.Sprintf("*近%s条交易总利润为: %0.5f eth, 详细交易数如下:*\n", offset, profit)
	for k, v := range detail {
		msg += fmt.Sprintf("%s[%s](https://www.dextools.io/app/cn/ether/pair-explorer/%s)*:* `%s`\n*B:* %0.3f | *S:* %0.3f | *P:* %0.5f eth\n", v.Scam, v.Symbol, k, k, v.Buy, v.Sell, v.Profit)
	}

	t.C <- msg

}
