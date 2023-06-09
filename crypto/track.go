package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/goconf"
)

type Track struct {
	C            chan string
	Newest       map[string]string
	apiKey       string
	Task         map[string]context.CancelFunc
	api          *Crypto
	dumpPath     string
	Keys         *PollingKey
	trackingLock sync.RWMutex
}

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
	Buy    float64
	Sell   float64
	Symbol string
	Profit float64
	Scam   string
	Pay    float64
	Time   string
}

func NewTrack() *Track {

	t := &Track{
		C:            make(chan string, 5),
		Newest:       recoverTrackingList(),
		apiKey:       goconf.VarStringOrDefault("", "crypto", "etherscan", "apiKey"),
		Task:         make(map[string]context.CancelFunc),
		api:          NewCrypto("", ""),
		dumpPath:     goconf.VarStringOrDefault("/usr/local/share/aio/", "crypto", "etherscan", "path"),
		Keys:         NewPollingKey(),
		trackingLock: sync.RWMutex{},
	}

	go t.DumpCron()
	go t.recover()
	go t.clearInactiveAddr()

	return t
}

func (t *Track) recover() {
	if t.Keys.IsNull() {
		log.Println("未读取到etherscan的apikey无法启动监控")
		return
	}
	for k := range t.Newest {
		ctx, cf := context.WithCancel(context.Background())
		t.Task[k] = cf
		go t.Tracking(k, ctx)
		// 免费api一秒可调用次数有限,分散请求防止达到阈值
		time.Sleep(time.Second)
	}
}

func (t *Track) clearInactiveAddr() {
	c := time.NewTicker(24 * time.Hour)
	defer c.Stop()

	handle := func(addr string, cl context.CancelFunc) {
		if t.Keys.IsNull() {
			return
		}
		addr = strings.ToLower(addr)
		url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=1&sort=desc&address=%s&apikey=%s"
		r, err := http.Get(fmt.Sprintf(url, addr, t.Keys.GetKey()))
		if err != nil {
			log.Println("请求失败")
			return
		}
		defer r.Body.Close()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("读取body失败")
			return
		}
		scan := new(TokenTxResp)
		err = json.Unmarshal(b, &scan)
		if err != nil {
			log.Println("WalletTracking: json转换失败")
			return
		}

		if scan.Status != "1" {
			return
		}

		if len(scan.Result) == 0 {
			return
		}

		ts, err := strconv.ParseInt(scan.Result[0].TimeStamp, 10, 64)
		if err == nil {
			if time.Unix(ts, 0).Add(3 * 24 * time.Hour).Before(time.Now()) {
				cl()
				delete(t.Newest, addr)
				delete(t.Task, addr)
				t.C <- fmt.Sprintf("`%s` 超过3天没有进行交易, 已停止追踪", addr)
			}
		}
	}

	for range c.C {
		for addr, cl := range t.Task {
			go handle(addr, cl)
		}
	}
}

func (t *Track) CronTracking(addr string) {
	if t.Keys.IsNull() {
		log.Println("未读取到etherscan的apikey无法启动监控")
		t.C <- "未读取到etherscan的apikey无法启动监控"
		return
	}
	addr = strings.ToLower(addr)
	if _, ok := t.Task[addr]; !ok {
		ctx, cf := context.WithCancel(context.Background())
		t.Task[addr] = cf
		t.Newest[addr] = ""
		go t.Tracking(addr, ctx)
		log.Println("开始追踪: ", addr)
		t.C <- "*开始追踪* " + addr
	}
}

func (t *Track) StopTracking(addr string) {
	addr = strings.ToLower(addr)
	if v, ok := t.Task[addr]; ok {
		v()
		delete(t.Task, addr)
		delete(t.Newest, addr)
		log.Println("已停止追踪: ", addr)
		t.C <- "*已停止追踪* " + addr
	}
}

func (t *Track) Tracking(addr string, ctx context.Context) {
	tick := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			go t.WalletTrackingV2(addr)
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
	if t.Keys.IsNull() {
		t.C <- "未读取到etherscan的apikey无法启动监控"
		return
	}
	addr = strings.ToLower(addr)
	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=%s&sort=desc&address=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, "3", addr, t.Keys.GetKey()))
	if err != nil {
		log.Println("请求失败")
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("读取body失败")
		return
	}
	scan := new(TokenTxResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		log.Println("WalletTracking: json转换失败")
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
		t.Newest[addr] = strings.ToLower(scan.Result[0].Hash)
		return
	}

	// 防止网络波动导致的newest错误
	t.trackingLock.Lock()
	defer t.trackingLock.Unlock()

	newest := ""

	sb := strings.Builder{}
	his := make(map[string]struct{})
	for _, record := range scan.Result {

		if strings.EqualFold(record.Hash, t.Newest[addr]) {
			break
		}

		if !strings.EqualFold(record.TokenSymbol, "WETH") || !isNull(record.From) || !isNull(record.To) {

			// 过滤掉重复记录
			if _, ok := his[strings.ToLower(record.Hash)]; ok {
				continue
			}
			balance := 0.0

			his[strings.ToLower(record.Hash)] = struct{}{}
			if strings.EqualFold(record.From, addr) {
				balance = t.getSellEthByHash(record.Hash, addr)
			} else {
				balance = t.getBuyEthByHash(record.Hash)
			}

			if balance == 0.0 {
				continue
			}

			if newest == "" {
				newest = strings.ToLower(record.Hash)
			}

			sb.WriteString("\n")
			isHoneypot := t.api.WhetherHoneypot(record.ContractAddress)
			if isHoneypot {
				sb.WriteString("*[SCAM]*")
			}
			if strings.EqualFold(record.From, addr) {
				sb.WriteString("*Sell: *")
			} else {
				sb.WriteString("*Buy: *")
			}
			sb.WriteString("[")
			sb.WriteString(record.TokenSymbol)
			sb.WriteString("](https://www.dextools.io/app/cn/ether/pair-explorer/")
			sb.WriteString(record.ContractAddress)
			sb.WriteString(") ")
			i, err := strconv.ParseInt(record.TimeStamp, 10, 64)
			if err == nil {
				sb.WriteString("*(")
				sb.WriteString(time.Unix(i, 0).Format("2006-01-02 15:04:05"))
				sb.WriteString(")*")
			}
			sb.WriteString("----[前往购买](https://app.uniswap.org/#/swap?outputCurrency=")
			sb.WriteString(record.ContractAddress)
			sb.WriteString("&chain=ethereum)")
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("%f", balance))
			sb.WriteString(" ETH / ")
			sb.WriteString(record.Value)
			sb.WriteString(" ")
			sb.WriteString(record.TokenSymbol)
			sb.WriteString("\n`")
			sb.WriteString(record.ContractAddress)
			sb.WriteString("`")
		}
	}

	if newest != "" {
		t.Newest[addr] = strings.ToLower(newest)
	}

	if sb.Len() > 0 {
		t.C <- "`" + addr + "` *执行操作:* " + sb.String()
	}

}

func (t *Track) AnalyzeAddrTokenProfit(addr, token string) {
	if t.Keys.IsNull() {
		t.C <- "未读取到etherscan的apikey无法启动监控"
		return
	}
	transferListUrl := "https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=%s&address=%s&apikey=%s"
	tx := new(TokenTxResp)
	err := common.HttpGet(fmt.Sprintf(transferListUrl, token, addr, t.Keys.GetKey()), &tx)
	if err != nil {
		log.Println("请求失败: ", err)
		return
	}

	if tx.Status != "1" {
		return
	}

	his := make(map[string]struct{})
	analyze := new(txs)

	for _, record := range tx.Result {

		if _, ok := his[strings.ToLower(record.Hash)]; ok {
			continue
		}

		analyze.Symbol = record.TokenSymbol

		volume := 0.0
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
					volume += cnt
				}
			}
		}
		val := 0.0
		// Sell
		if strings.EqualFold(record.From, addr) {
			val = t.getSellEthByHash(record.Hash, addr)
			analyze.Sell += volume
			analyze.Profit += val
			// Buy
		} else {
			val = t.getBuyEthByHash(record.Hash)
			analyze.Buy += volume
			analyze.Profit -= val
			analyze.Pay += val
		}

		his[strings.ToLower(record.Hash)] = struct{}{}

	}

	t.C <- fmt.Sprintf("[%s](https://www.dextools.io/app/cn/ether/pair-explorer/%s)*总利润为: %0.5f eth: *\n*B:* %0.2f | *S:* %0.2f | *C:* %0.3f eth\n", analyze.Symbol, token, analyze.Profit, analyze.Buy, analyze.Sell, analyze.Pay)
}

func (t *Track) getBuyEthByHash(hash string) float64 {
	url := "https://api.etherscan.io/api?module=account&action=txlistinternal&txhash=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, hash, t.Keys.GetKey()))
	if err != nil {
		log.Println("请求失败")
		return 0.0
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("读取body失败")
		return 0.0
	}
	scan := new(txResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		log.Println("json转换失败")
		return 0.0
	}

	if scan.Status != "1" || len(scan.Result) == 0 {
		fmt.Printf(url, hash, t.Keys.GetKey())
		log.Printf("\n%+v\n", scan)
		return 0.0
	}

	cnt := 0.0
	router := ""
	for _, v := range scan.Result {
		if strings.EqualFold("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", v.To) {
			router = v.From
			break
		}
	}

	for _, v := range scan.Result {
		tmp := ""
		// 暂时认定买入操作只会有一轮内部交易
		if !strings.EqualFold(router, v.From) {
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

func (t *Track) getSellEthByHash(hash, addr string) float64 {
	url := "https://api.etherscan.io/api?module=account&action=txlistinternal&txhash=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, hash, t.Keys.GetKey()))
	if err != nil {
		log.Println("请求失败")
		return 0.0
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("读取body失败")
		return 0.0
	}
	scan := new(txResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		log.Println("json转换失败")
		return 0.0
	}

	if scan.Status != "1" || len(scan.Result) == 0 {
		fmt.Printf(url, hash, t.Keys.GetKey())
		log.Printf("\n%+v\n", scan)
		return 0.0
	}

	cnt := 0.0
	for _, v := range scan.Result {
		if strings.EqualFold(v.To, addr) {
			tmp := ""
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
	}
	return cnt
}

func (t *Track) WalletTxAnalyze(addr string, offset string) {
	if t.Keys.IsNull() {
		t.C <- "未读取到etherscan的apikey无法调用api"
		return
	}
	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=%s&sort=desc&address=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, offset, addr, t.Keys.GetKey()))
	if err != nil {
		log.Println("etherscan请求失败")
		t.C <- "etherscan请求失败"
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("读取body失败")
		t.C <- "读取body失败"
		return
	}
	scan := new(TokenTxResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		log.Println("json转换失败")
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
	// 最近一次交易时间
	recent := ""
	for _, record := range scan.Result {
		if !strings.EqualFold(record.TokenSymbol, "WETH") || !isNull(record.From) || !isNull(record.To) {
			if _, ok := his[strings.ToLower(record.Hash)]; ok {
				continue
			}

			val := 0.0

			his[strings.ToLower(record.Hash)] = struct{}{}
			if strings.EqualFold(record.From, addr) {
				val = t.getSellEthByHash(record.Hash, addr)
			} else {
				val = t.getBuyEthByHash(record.Hash)
			}
			if val == 0.0 {
				continue
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
						ts, err := strconv.ParseInt(record.TimeStamp, 10, 64)
						if err == nil {
							if recent == "" {
								recent = record.TimeStamp
							}
							detail[record.ContractAddress].Time = time.Unix(ts, 0).Format("2006-01-02 15:04:05")
						}
						detail[record.ContractAddress].Symbol = record.TokenSymbol
						isHoneypot := t.api.WhetherHoneypot(record.ContractAddress)
						if isHoneypot {
							detail[record.ContractAddress].Scam = "SCAM"
						}
					}
				}
			}
		}
	}

	warn := ""

	if recent != "" {
		ts, err := strconv.ParseInt(recent, 10, 64)
		if err == nil {
			if time.Unix(ts, 0).Add(10 * 24 * time.Hour).Before(time.Now()) {
				warn = "*该地址已超过十天未交易,最后一次:" + time.Unix(ts, 0).Format("2006-01-02 15:04:05") + "*\n"
			}
		}
	}

	netMargin := profit

	for _, v := range detail {
		if v.Buy == 0.0 || v.Sell == 0.0 {
			netMargin -= v.Profit
		}
	}

	msg := fmt.Sprintf("%s[Wallet](https://etherscan.io/address/%s#tokentxns)*总利润: %0.5f | 净利润: %0.5f :*\n", warn, addr, profit, netMargin)
	for k, v := range detail {
		if len(msg) > 3500 {
			msg += "*------内容过长进行裁剪------*"
			t.C <- msg
			time.Sleep(time.Microsecond * 100)
			msg = "*------裁剪后的另外部分------\n*"
		}
		unsold := ""
		if v.Sell == 0.0 {
			unsold = "UNSOLE"
		}
		msg += fmt.Sprintf("[%s](https://www.dextools.io/app/cn/ether/pair-explorer/%s)*:* `%s`\n*%s* | *Detect:%s* | *Status:%s*\n*B:* %0.2f | *S:* %0.2f | *C:* %0.3f | *P:* %0.3f eth\n", v.Symbol, k, k, v.Time, v.Scam, unsold, v.Buy, v.Sell, v.Pay, v.Profit)
	}

	t.C <- msg

}

func (t *Track) DumpTrackingList(tip bool) {
	b, err := json.Marshal(t.Newest)
	if err != nil {
		log.Println("序列化失败:", err)
		if tip {
			t.C <- "dump失败: list序列化报错"
		}
		return
	}

	if _, err := os.Stat(t.dumpPath); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.dumpPath, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地文件夹失败")
			if tip {
				t.C <- "dump失败: 创建本地文件夹失败"
			}
			return
		}
	}
	err = os.WriteFile(t.dumpPath+"tracking.json", b, 0644)
	if err != nil {
		log.Println("dump文件创建/写入失败")
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
		return dump
	}

	json.Unmarshal(b, &dump)

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
	if t.Keys.IsNull() {
		t.C <- "未读取到etherscan的apikey无法启动分析"
		return
	}
	// getContractCreationUrl := "https://api.etherscan.io/api?module=contract&action=getcontractcreation&contractaddresses=%s&apikey=%s"
	// creator := new(ContractCreationResp)
	// err := common.HttpGet(fmt.Sprintf(getContractCreationUrl, token, apiKey), &creator)

	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=%s&offset=%s&sort=asc&contractaddress=%s&apikey=%s"
	scan := new(TokenTxResp)
	err := common.HttpGet(fmt.Sprintf(url, page, offset, token, t.Keys.GetKey()), &scan)
	if err != nil {
		log.Println("请求失败: ", err)
		return
	}

	if scan.Status != "1" {
		return
	}

	recorded := map[string]struct{}{
		"0x0000000000000000000000000000000000000000": {},
		"0x000000000000000000000000000000000000dead": {},
		// uniswap router
		"0x3fc91a3afd70395cd496c647d5a6cc9d4b2b7fad": {},
		"0x68b3465833fb72a70ecdf485e0e4c7bd8665fc45": {},
	}

	analyze := make(map[string]*txs)

	handle := func(address string) {
		address = strings.ToLower(address)
		if _, ok := recorded[address]; !ok {
			recorded[address] = struct{}{}
			his := make(map[string]struct{})
			list := t.TransferList(address, token)
			if len(list) != 0 {
				analyze[address] = new(txs)
				for _, tx := range list {
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

					val := 0.0
					if _, ok := his[tx.Hash]; !ok {
						his[tx.Hash] = struct{}{}
						if strings.EqualFold(tx.From, address) {
							val = t.getSellEthByHash(tx.Hash, address)
						} else {
							val = t.getBuyEthByHash(tx.Hash)
						}
					}

					if strings.EqualFold(address, tx.From) {
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
				msg += "\n*------内容过长进行裁剪------*"
				t.C <- msg
				time.Sleep(time.Microsecond * 100)
				msg = "*------裁剪后的另外部分------*"
			}
			if !(v.Pay == 0.0 || v.Profit < 0.0) {
				msg += fmt.Sprintf("\n`%s`\n*B:* %0.3f | *S:* %0.3f | *C:* %0.3f | *P:* %0.3f ETH", k, v.Buy, v.Sell, v.Pay, v.Profit)
			}
		}
		t.C <- msg
	}

}

func (t *Track) TransferList(addr, token string) []TokenTx {
	transferListUrl := "https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=%s&address=%s&apikey=%s"
	tx := new(TokenTxResp)
	err := common.HttpGet(fmt.Sprintf(transferListUrl, token, addr, t.Keys.GetKey()), &tx)

	if err != nil {
		log.Println("请求失败: ", err)
		return nil
	}
	if len(tx.Result) > 6 {
		return nil
	}

	return tx.Result
}

func (t *Track) WalletTrackingV2(addr string) {
	now := time.Now()

	addr = strings.ToLower(addr)
	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=1&sort=desc&address=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, addr, t.Keys.GetKey()))
	if err != nil {
		log.Println("请求失败")
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("读取body失败")
		return
	}
	scan := new(TokenTxResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		log.Println("WalletTracking: json转换失败")
		return
	}

	if scan.Status != "1" || len(scan.Result) == 0 {
		return
	}

	if strings.EqualFold(t.Newest[addr], scan.Result[0].Hash) {
		return
	}

	// 首次加入探测忽略
	if t.Newest[addr] == "" {
		t.Newest[addr] = scan.Result[0].Hash
		return
	}

	t.Newest[addr] = scan.Result[0].Hash

	wg := sync.WaitGroup{}
	sb := strings.Builder{}

	record := scan.Result[0]

	if strings.EqualFold(record.TokenSymbol, "WETH") || isNull(record.From) || isNull(record.To) {
		return
	}
	balance := 0.0
	isHoneypot := ""
	detail := ""
	check := ""

	getBalance := func() {
		defer wg.Done()
		if strings.EqualFold(record.From, addr) {
			balance += t.getSellEthByHash(record.Hash, addr)
		} else {
			balance += t.getBuyEthByHash(record.Hash)
		}
		log.Println("getBalance耗时: ", time.Since(now))
	}

	// -1s
	// getHoneypot := func() {
	// 	defer wg.Done()
	// 	if t.api.WhetherHoneypot(record.ContractAddress) {
	// 		isHoneypot += "*[SCAM]*"
	// 	}
	// 	log.Println("getHoneypot耗时: ", time.Since(now))
	// }

	getDetail := func() {
		defer wg.Done()
		pair := t.api.MemePrice(record.ContractAddress, "eth")
		if pair != nil {
			detail += fmt.Sprintf("   |   *Price: $%s (%d)*\n\n*Pool: $%0.f  |  CreationTime: %s*\n\n*5M:    %0.2f%%    $%0.2f    %d/%d*\n*1H:    %0.2f%%    $%0.2f    %d/%d*\n*6H:    %0.2f%%    $%0.2f    %d/%d*\n*1D:    %0.2f%%    $%0.2f    %d/%d*\n", pair.PriceUsd, zeroCal(pair.PriceUsd), pair.Lp.Usd, pair.CreateTime, pair.PriceChange.M5, pair.Volume.M5, pair.Txns.M5.B, pair.Txns.M5.S, pair.PriceChange.H1, pair.Volume.H1, pair.Txns.H1.B, pair.Txns.H1.S, pair.PriceChange.H6, pair.Volume.H6, pair.Txns.H6.B, pair.Txns.H6.S, pair.PriceChange.H24, pair.Volume.H24, pair.Txns.H24.B, pair.Txns.H24.S)
			log.Println("getDetail耗时: ", time.Since(now))
		}
	}

	getCheck := func() {
		defer wg.Done()
		ck := t.api.MemeCheck(record.ContractAddress, "eth")
		if ck != nil {
			check += fmt.Sprintf("*Buy Tax: %s   |   Sell Tax: %s   |   Locked LP: %0.2f%%*\n*Owner:* `%s`\n*Creator Percent: %s   |   Balance: %s*", ck.BuyTax, ck.SellTax, ck.LpLockedTotal*100.0, ck.OwnerAddress, ck.CreatorPercent, ck.CreatorBalance)
		}
		log.Println("getCheck耗时: ", time.Since(now))
	}

	wg.Add(3)

	// 并发减少等待时间
	go getBalance()
	//go getHoneypot()
	go getDetail()
	go getCheck()

	wg.Wait()

	sb.WriteString("\n")
	sb.WriteString(isHoneypot)
	if strings.EqualFold(record.From, addr) {
		sb.WriteString("*SELL: *")
	} else {
		sb.WriteString("*BUY: *")
	}
	sb.WriteString("[")
	sb.WriteString(record.TokenSymbol)
	sb.WriteString("](https://www.dextools.io/app/cn/ether/pair-explorer/")
	sb.WriteString(record.ContractAddress)
	sb.WriteString(") ")
	ts, err := strconv.ParseInt(record.TimeStamp, 10, 64)
	if err == nil {
		sb.WriteString("*(")
		sb.WriteString(time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
		sb.WriteString(")*")
	}
	sb.WriteString("----[前往购买](https://app.uniswap.org/#/swap?exactField=input&exactAmount=0.02&inputCurrency=ETH&outputCurrency=")
	sb.WriteString(record.ContractAddress)
	sb.WriteString("&chain=ethereum)")
	sb.WriteString("\n\n`")
	sb.WriteString(record.ContractAddress)
	sb.WriteString("`")
	sb.WriteString("\n\n*Cost: ")
	sb.WriteString(fmt.Sprintf("%f", balance))
	sb.WriteString(" ETH*")
	sb.WriteString(detail)
	sb.WriteString("\n")
	sb.WriteString(check)

	log.Println("查询总耗时: ", time.Since(now))

	t.C <- "`" + addr + "` *监控* [TX](https://etherscan.io/tx/" + record.Hash + ")" + sb.String()
}

func isNull(addr string) bool {
	return strings.EqualFold("0x0000000000000000000000000000000000000000", addr) || strings.EqualFold("0x000000000000000000000000000000000000dead", addr)
}

func zeroCal(str string) (zero int) {
	if !strings.Contains(str, ".") {
		return zero
	}
	for i := strings.Index(str, ".") + 1; i < len(str)-1; i++ {
		if str[i] != '0' {
			break
		}
		zero++
	}
	return zero
}

func (t *Track) WalletLastTransaction() {
	if t.Keys.IsNull() {
		t.C <- "未读取到etherscan的apikey无法启动监控"
		return
	}

	sb := strings.Builder{}
	sb.WriteString("*当前监控的钱包最后交易时间:*")
	getErc20 := func(addr string) {
		url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=1&sort=desc&address=%s&apikey=%s"
		scan := new(TokenTxResp)
		err := common.HttpGet(fmt.Sprintf(url, addr, t.Keys.GetKey()), &scan)
		if err != nil {
			log.Println("请求失败: ", err)
			return
		}

		if scan.Status != "1" || len(scan.Result) == 0 {
			log.Println("响应码为1或者结果长度为0")
			return
		}

		ts, err := strconv.ParseInt(scan.Result[0].TimeStamp, 10, 64)
		if err == nil {
			sb.WriteString(fmt.Sprintf("\n`%s`*: %s*", addr, time.Unix(ts, 0).Format("2006-01-02 15:04:05")))
		}

	}

	for addr := range t.Task {
		getErc20(addr)
	}

	t.C <- sb.String()

}
