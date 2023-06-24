package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/goconf"
)

type Track struct {
	C      chan string
	Newest map[string]string
	apiKey string
	Task   map[string]context.CancelFunc
	api    *Crypto
	dumpPath string
}

type txResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []tx   `json:"result"`
}

type tx struct {
	Value string `json:"value"`
	From string `json:"from"`
	To string `json:"To"`
}

type txs struct {
	Buy    float64
	Sell   float64
	Symbol string
	Profit float64
	Scam   string
	Pay    float64
}

func NewTrack() *Track {

	t := &Track{
		C:      make(chan string, 5),
		Newest: recoverTrackingList(),
		apiKey: goconf.VarStringOrDefault("", "crypto", "etherscan", "apiKey"),
		Task:   make(map[string]context.CancelFunc),
		api:    NewCrypto("", ""),
		dumpPath: goconf.VarStringOrDefault("/usr/local/share/aio/", "crypto", "etherscan", "path"),
	}

	go t.DumpCron()
	go t.Recover()

	return t
}

func (t *Track) Recover() {
	for k := range t.Newest {
		ctx, cf := context.WithCancel(context.Background())
		t.Task[k] = cf
		go t.Tracking(k, ctx)
		// 免费api一秒可调用次数有限,分散请求防止达到阈值
		time.Sleep(time.Second)
	}
}

func (t *Track) CronTracking(addr string) {
	if _, ok := t.Task[addr]; !ok {
		ctx, cf := context.WithCancel(context.Background())
		t.Task[addr] = cf
		t.Newest[addr] = ""
		go t.Tracking(addr, ctx)
		t.C <- "*开始追踪* " + addr
	}
}

func (t *Track) StopTracking(addr string) {
	if v, ok := t.Task[addr]; ok {
		v()
		delete(t.Task, addr)
		delete(t.Newest, addr)
		t.C <- "*已停止追踪* " + addr
	}
}

func (t *Track) Tracking(addr string, ctx context.Context) {
	tick := time.NewTicker(time.Second * 10)
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

	// 首次不做探测
	if t.Newest[addr] == "" {
		t.Newest[addr] = scan.Result[0].Hash
		return
	}

	sb := strings.Builder{}
	his := make(map[string]struct{})
	for _, record := range scan.Result {
		if record.Hash == t.Newest[addr] {
			break
		}
		if record.TokenSymbol != "WETH" || !isNull(record.From) || !isNull(record.To) {
			balance := 0.0
			if _, ok := his[record.Hash]; !ok {
				his[record.Hash] = struct{}{}
				balance = t.getEthByHash(record.Hash)
				if balance == 0 {
					continue
				}
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
			sb.WriteString(fmt.Sprintf("%f", balance))
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

	t.Newest[addr] = scan.Result[0].Hash

	if sb.Len() > 0 {
		t.C <- "`" + addr + "`*执行操作:*" + sb.String()
	}

}

func (t *Track) getEthByHash(hash string) float64 {
	url := "https://api.etherscan.io/api?module=account&action=txlistinternal&txhash=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, hash, t.apiKey))
	if err != nil {
		fmt.Println("请求失败")
		return 0
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("读取body失败")
		return 0
	}
	scan := new(txResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		fmt.Println("json转换失败")
		return 0
	}

	if scan.Status != "1" || len(scan.Result) == 0 {
		return 0
	}

	cnt := 0.0
	for _, v := range scan.Result {
		tmp := ""
		// 内部交易可能会出现多轮交易,该步骤用于过滤 0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2为WETH
		if !strings.EqualFold("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", v.From) && !strings.EqualFold("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", v.To) {
			continue
		}
		if len(v.Value) > 18 {
			tmp = v.Value[:len(v.Value)-18] + "." + v.Value[len(v.Value)-18:]
		} else if len(v.Value) == 18 {
			tmp = "0." + v.Value
		} else {
			tmp = "0." + strings.Repeat("0", 18-len(v.Value)) + v.Value
		}
		f, err := strconv.ParseFloat(tmp, 64)
		if err == nil {
			cnt += f
		}
	}
	return cnt

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

	
	profit := 0.0
	detail := make(map[string]*txs)
	his := make(map[string]struct{})

	for _, record := range scan.Result {
		if record.TokenSymbol != "WETH" || !isNull(record.From) || !isNull(record.To) {
			val := 0.0
			if _, ok := his[record.Hash]; !ok {
				his[record.Hash] = struct{}{}
				val = t.getEthByHash(record.Hash)
				if val == 0 {
					continue
				}
			}
			
			if strings.EqualFold(record.From, addr) {
				profit += val
			} else {
				profit -= val
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
							detail[record.ContractAddress].Pay += val
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

	msg := fmt.Sprintf("[Wallet](https://etherscan.io/address/%s#tokentxns)*近%s条交易总利润为: %0.5f eth: *\n", addr, offset, profit)
	for k, v := range detail {
		msg += fmt.Sprintf("%s[%s](https://www.dextools.io/app/cn/ether/pair-explorer/%s)*:* `%s`\n*B:* %0.2f | *S:* %0.2f | *C:* %0.5f eth | *P:* %0.5f eth\n", v.Scam, v.Symbol, k, k, v.Buy, v.Sell, v.Pay, v.Profit)
	}

	t.C <- msg

}

func (t *Track) DumpTrackingList(tip bool) {
	if len(t.Newest) == 0 {
		if tip {
			t.C <- "列表为空,不执行dump"
		}
		return
	}
	b, err := json.Marshal(t.Newest)
	if err != nil {
		fmt.Println("序列化失败:", err)
		if tip {
			t.C <- "dump失败: list序列化报错"
		}
		return
	}

	if _, err := os.Stat(t.dumpPath); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.dumpPath, os.ModePerm) // 创建目录
		if err != nil {
			fmt.Println("创建本地文件夹失败")
			if tip {
				t.C <- "dump失败: 创建本地文件夹失败"
			}
			return
		}
	}
	err = os.WriteFile(t.dumpPath+"tracking.json", b, 0644)
	if err != nil {
		fmt.Println("dump文件创建/写入失败")
		if tip {
			t.C <- "dump失败: dump文件创建/写入失败"
		}
		return
	}

	if tip {
		t.C <- "dump完成"
	}

}

func recoverTrackingList() map[string]string {
	dump := make(map[string]string)
	b, err := os.ReadFile(goconf.VarStringOrDefault("/usr/local/share/aio/", "crypto", "etherscan", "path") + "tracking.json")
	if err != nil {
		fmt.Println("dump文件读取失败:", err)
		return dump
	}

	err = json.Unmarshal(b, &dump)
	if err != nil {
		fmt.Println("TrackingList数据读取失败:", err)
		return dump
	}

	return dump

}

func (t *Track) DumpCron() {
	h := time.NewTicker(time.Hour)
	for range h.C {
		t.DumpTrackingList(false)
	}
}

// 1. 拉取某个 token 最初的交易列表
// 2. 遍历列表, 查询所有 from 和 to 的地址在该 token 的交易记录
// 3. 过滤掉买卖数超过 6 的地址, 此类一般为夹子机器人
// 4. 对交易记录遍历查询内部交易, 对地址的买卖数收益记录
func (t *Track) SmartAddrFinder(token, offset, page string) {
	if t.apiKey == "" {
		t.C <- "未读取到etherscan的apikey无法启动分析"
		return
	}
	// getContractCreationUrl := "https://api.etherscan.io/api?module=contract&action=getcontractcreation&contractaddresses=%s&apikey=%s"
	// creator := new(ContractCreationResp)
	// err := common.HttpGet(fmt.Sprintf(getContractCreationUrl, token, apiKey), &creator)

	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=%s&offset=%s&sort=asc&contractaddress=%s&apikey=%s"
	scan := new(TokenTxResp)
	err := common.HttpGet(fmt.Sprintf(url, page, offset, token, t.apiKey), &scan)
	if err != nil {
		fmt.Println("请求失败: ", err)
		return
	}

	if scan.Status != "1" {
		return
	}

	recorded := map[string]struct{}{
		"0x0000000000000000000000000000000000000000": {},
		"0x000000000000000000000000000000000000dead": {},
		// uniswap router
		"0x3fC91A3afd70395Cd496C647d5a6CC9D4B2b7FAD": {},
		"0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45": {},
	}

	analyze := make(map[string]*txs)

	handle := func (address string)  {
		if _, ok := recorded[address]; !ok {
			recorded[address] = struct{}{}
			list := t.TransferList(address, token)
			if len(list) != 0 {
				analyze[address] = new(txs)
				for _, tx := range list {
					val := t.getEthByHash(tx.Hash)
					cnt := 0.0
					dec, err := strconv.Atoi(tx.Decimal)
					if err == nil {
						l := len(tx.Value) - dec
						tmp := ""
						if l <= 0 {
							tmp = "0." + strings.Repeat("0", -l) + tx.Value
						} else {
							tmp = tx.Value[:l]
						}
						cnt, _ = strconv.ParseFloat(tmp, 64)
					}
					
					if address == tx.From {
						// sell
						analyze[address].Profit += val
						analyze[address].Sell += cnt
					} else {
						// buy
						analyze[address].Profit -= val
						analyze[address].Buy += cnt
						analyze[address].Pay += val
					}
				}
			}
		}
	}

	for _, v := range scan.Result {
		handle(v.From)
		handle(v.To)
	}

	if len(analyze) > 0 {
		msg := fmt.Sprintf("*合约地址:* `%s`\n *------------分析完毕:------------*", token)
		for k, v := range analyze {
			if len(msg) > 3500 {
				msg += "\n内容过长进行裁剪"
				t.C <- msg
				msg = "裁剪后的下部分:"
			}
			msg += fmt.Sprintf("\n`%s`\n*B:* %0.3f | *S:* %0.3f | *C:* %0.5f | *P:* %0.5f ETH", k, v.Buy, v.Sell, v.Pay, v.Profit)
		}
		t.C <- msg
	}
	
}

func (t *Track) TransferList(addr, token string) []TokenTx {
	transferListUrl := "https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=%s&address=%s&apikey=%s"
	tx := new(TokenTxResp)
	err := common.HttpGet(fmt.Sprintf(transferListUrl, token, addr, t.apiKey), &tx)

	if err != nil {
		fmt.Println("请求失败: ", err)
		return nil
	}
	if len(tx.Result) > 6 {
		return nil
	}

	return tx.Result
}


func isNull(addr string) bool {
	return strings.EqualFold("0x0000000000000000000000000000000000000000", addr) || strings.EqualFold("0x000000000000000000000000000000000000dead", addr)
}