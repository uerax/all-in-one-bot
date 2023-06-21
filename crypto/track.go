package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/uerax/goconf"
)

type Track struct {
	C      chan string
	Newest map[string]string
	apiKey string
	Task   map[string]context.CancelFunc
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
		Task: make(map[string]context.CancelFunc),
	}
}

func (t *Track) CronTracking(addr string) {
	if _, ok := t.Task[addr]; !ok {
		ctx, cf := context.WithCancel(context.Background())
		t.Task[addr] = cf
		go t.Tracking(addr, ctx)
		t.C <- "添加完成"
	}
}

func (t *Track) StopTracking(addr string) {
	if v, ok := t.Task[addr]; ok {
		v()
		delete(t.Task, addr)
		t.C <- "已停止追踪"
	}
}

func (t *Track) Tracking(addr string, ctx context.Context) {
	tick := time.NewTicker(time.Minute)
	for {
		select {
		case <- ctx.Done():
			return
		case <- tick.C:
			go t.WalletTracking(addr)
		}
	}
}

func (t *Track) WalletTracking(addr string) {
	if t.apiKey == "" {
		t.C <- "未读取到etherscan的apikey无法启动监控"
		return
	}
	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=%s&sort=desc&address=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, "10", addr, t.apiKey))
	if err != nil {
		fmt.Println("请求失败")
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
		t.C <- scan.Message
		return
	}

	if len(scan.Result) == 0 || strings.EqualFold(scan.Result[0].Hash, t.Newest[addr]) {
		return
	}

	sb := strings.Builder{}

	for _, record := range scan.Result {
		if record.TokenSymbol != "WETH" {
			if strings.EqualFold(record.From, addr) {
				sb.WriteString("\n*Sell:*")
			} else {
				sb.WriteString("\n*Buy:*")
			}
			sb.WriteString("[")
			sb.WriteString(record.TokenSymbol)
			sb.WriteString("](https://www.dextools.io/app/cn/ether/pair-explorer/")
			sb.WriteString(record.ContractAddress)
			sb.WriteString("):")
			sb.WriteString(t.getEthByHash(record.Hash))
			sb.WriteString(" ETH\n")
			sb.WriteString("`")
			sb.WriteString(record.ContractAddress)
			sb.WriteString("`")
		}
	}

	if sb.Len() > 0 {
		t.C <- "监控地址执行操作:\n" + sb.String()
	}

	t.Newest[addr] = scan.Result[0].Hash

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
