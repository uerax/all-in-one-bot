package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/uerax/goconf"
)

const ()

type line struct {
	LowPrice   string
	HighPrice  string
	LowNotify  int64
	HighNotify int64
}

type Probe struct {
	HighLine      map[string]*line
	LowLine       map[string]*line
	C             chan map[string]string
	Kline         chan string
	Meme          chan string
	frequency     int64
	api           *Crypto
	task          map[string]context.CancelFunc
	memeHighTask  map[string]context.CancelFunc
	memeLowTask   map[string]context.CancelFunc
	smartAddr     map[string]context.CancelFunc
	smartBuys     map[string]map[string]struct{}
	smartItv      int
	smartDumpPath string
}

func NewProbe() *Probe {
	p := &Probe{
		HighLine:      make(map[string]*line),
		LowLine:       make(map[string]*line),
		C:             make(chan map[string]string),
		Kline:         make(chan string, 5),
		Meme:          make(chan string, 5),
		frequency:     goconf.VarInt64OrDefault(600, "crypto", "monitor", "frequency"),
		api:           NewCrypto(goconf.VarStringOrDefault("", "crypto", "binance", "apiKey"), goconf.VarStringOrDefault("", "crypto", "binance", "secretKey")),
		task:          make(map[string]context.CancelFunc),
		memeLowTask:   make(map[string]context.CancelFunc),
		memeHighTask:  make(map[string]context.CancelFunc),
		smartAddr:     make(map[string]context.CancelFunc),
		smartBuys:     recoverSmartAddrList(),
		smartItv:      goconf.VarIntOrDefault(30, "crypto", "etherscan", "interval"),
		smartDumpPath: goconf.VarStringOrDefault("/usr/local/share/aio/", "crypto", "etherscan", "path"),
	}

	go p.DumpCron()

	return p
}

func (t *Probe) ListKLineProbe() string {
	b := strings.Builder{}
	b.WriteString("当前正在探测的加密货币有:")
	for k := range t.task {
		b.WriteString("\n`")
		b.WriteString(k)
		b.WriteString("`")
	}
	return b.String()
}

func (t *Probe) AddKLineProbe(crypto string) {
	if _, ok := t.task[crypto]; !ok {
		ctx, cf := context.WithCancel(context.Background())
		t.task[crypto] = cf
		go t.KLineProbe(crypto, ctx)
		t.Kline <- fmt.Sprintf("永续合约: %s 监控已启动", crypto)
	}
}

func (t *Probe) StopKLineProbe(crypto string) {
	if v, ok := t.task[crypto]; ok {
		v()
		delete(t.task, crypto)
		t.Kline <- fmt.Sprintf("永续合约: %s 监控已关闭", crypto)
	}
}

func (t *Probe) KLineProbe(crypto string, ctx context.Context) {
	now := time.Now()
	p := now.Minute() / 15
	instance := 15*(p+1)*60 - now.Minute()*60 - now.Second() - 60
	fmt.Printf("距离15分钟系数还有%d", now.Add(time.Duration(instance*int(time.Second))).Unix())
	time.Sleep(time.Duration(instance) * time.Second)

	do := func() {
		kline := t.api.UFutureKline("15m", 5, crypto)

		if len(kline) == 5 && ((kline[4]+kline[3]+kline[2] == -3 && kline[0]+kline[1] != -2) || (kline[4]+kline[3]+kline[2] == 3 && kline[0]+kline[1] != 2)) {
			t.Kline <- fmt.Sprintf("永续合约: %s 连续三根15m的K线走势一致", crypto)
		}

	}
	do()

	ticker := time.NewTicker(15 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			t.Kline <- fmt.Sprintf("永续合约: %s 监控已关闭", crypto)
			return
		case <-ticker.C:
			do()
		}
	}
}

func (p *Probe) MemePrice(query string, chain string) {

	pair := new(Pair)
	if chain == "" {
		chain = "eth"
		pair = p.api.MemePrice(query, chain)
		if pair == nil {
			chain = "bsc"
			pair = p.api.MemePrice(query, chain)
			if pair == nil {
				p.Meme <- "没有查询到相关合约"
				return
			}
		}
	} else {
		pair = p.api.MemePrice(query, chain)
		if pair == nil {
			p.Meme <- "查询失败,请检查参数"
			return
		}
	}

	s := fmt.Sprintf("*%s:$%s* \n*Chain:* %s | *Price:* $%s\n\n*5M:*    %0.2f%%    $%0.2f    %d/%d\n*1H:*    %0.2f%%    $%0.2f    %d/%d\n*6H:*    %0.2f%%    $%0.2f    %d/%d\n*1D:*    %0.2f%%    $%0.2f    %d/%d\n\n", pair.BaseToken.Name, pair.BaseToken.Symbol, pair.ChainId, pair.PriceUsd, pair.PriceChange.M5, pair.Volume.M5, pair.Txns.M5.B, pair.Txns.M5.S, pair.PriceChange.H1, pair.Volume.H1, pair.Txns.H1.B, pair.Txns.H1.S, pair.PriceChange.H6, pair.Volume.H6, pair.Txns.H6.B, pair.Txns.H6.S, pair.PriceChange.H24, pair.Volume.H24, pair.Txns.H24.B, pair.Txns.H24.S)

	check := p.api.MemeCheck(query, chain)
	if check != nil {
		if strings.Contains(check.TotalSupply, ".") {
			check.TotalSupply = check.TotalSupply[:strings.Index(check.TotalSupply, ".")]
		}
		if strings.Contains(check.LpTotalSupply, ".") {
			check.LpTotalSupply = check.LpTotalSupply[:strings.Index(check.LpTotalSupply, ".")]
		}
		s += fmt.Sprintf("*Buy Tax:* %s | *Sell Tax:* %s\n*Total Supply:* %s\n*LP Supply:* %s\n*Holder:* %s\n*Owner:* `%s`\n*Creator:* `%s`\nPercent:* %s | Balance:* %s\n", check.BuyTax, check.SellTax, check.TotalSupply, check.LpTotalSupply, check.HolderCount, check.OwnerAddress, check.CreatorAddress, check.CreatorPercent, check.CreatorBalance)
	}

	chainScan := ""
	honeypot := ""
	moonarch := ""
	dextools := ""
	if chain == "bsc" {
		chainScan = "https://bscscan.com/address/"
		honeypot = "https://honeypot.is/?address="
		dextools = "https://www.dextools.io/app/cn/bnb/pair-explorer/"
	} else if chain == "eth" {
		chainScan = "https://etherscan.io/address/"
		honeypot = "https://honeypot.is/ethereum?address="
		moonarch = "eth."
		dextools = "https://www.dextools.io/app/cn/ether/pair-explorer/"
	}

	s += fmt.Sprintf("*Trade:* [dextools](%s%s) | [dexscreener](%s) | [ave.ai](https://ave.ai/token/%s-%s) | [dexview](https://www.dexview.com/%s/%s)\n\n`%s`\n\n*Check:* [ChainScan](%s%s) | [Moonarch](https://%smoonarch.app/token/%s) | [Honeypot](%s%s)", dextools, pair.PairAddress, pair.URL, pair.BaseToken.Addr, chain, chain, pair.BaseToken.Addr, pair.BaseToken.Addr, chainScan, pair.BaseToken.Addr, moonarch, pair.BaseToken.Addr, honeypot, pair.BaseToken.Addr)

	p.Meme <- s
}

func (p *Probe) MemeMonitorList() {
	b := strings.Builder{}
	b.WriteString("当前正在监控的meme币有:")
	for k := range p.memeHighTask {
		b.WriteString("\n`")
		b.WriteString(k)
		b.WriteString("`")
	}
	p.Meme <- b.String()
}

func (p *Probe) CloseMemeMonitor(query string, chain string) {
	if _, ok := p.memeHighTask[query+" "+chain]; ok {
		p.memeHighTask[query+" "+chain]()
		delete(p.memeHighTask, query+" "+chain)
		p.Meme <- query + " " + chain + "\n上涨监控已关闭"
	}
	if _, ok := p.memeLowTask[query+" "+chain]; ok {
		p.memeHighTask[query+" "+chain]()
		delete(p.memeLowTask, query+" "+chain)
		p.Meme <- query + " " + chain + "\n下跌监控已关闭"
	}
}

func (p *Probe) MemeGrowthMonitor(query string, chain string, price string) {
	test := p.api.MemePrice(query, chain)
	if test == nil {
		p.Meme <- "token查询失败,请检查token是否有误"
		return
	}
	t := time.NewTicker(time.Minute)
	ctx, cf := context.WithCancel(context.Background())
	p.memeHighTask[query+" "+chain] = cf
	p.Meme <- "开始上涨监控: \n" + query + " " + chain
	for {
		select {
		case <-t.C:
			line, err := strconv.ParseFloat(price, 64)
			if err != nil {
				p.Meme <- "输入的价格有误,无法识别"
				fmt.Println("价格转换异常：", err)
				delete(p.memeHighTask, query+" "+chain)
				return
			}
			pair := p.api.MemePrice(query, chain)
			if pair == nil {
				continue
			}
			now, err := strconv.ParseFloat(pair.PriceUsd, 64)
			if err != nil {
				p.Meme <- "价格转换异常,请检查日志"
				fmt.Println("价格转换异常：", err)
				delete(p.memeHighTask, query+" "+chain)
				return
			}
			if line <= now {
				chainScan := ""
				honeypot := ""
				moonarch := ""
				dextools := ""
				if chain == "bsc" {
					chainScan = "https://bscscan.com/address/"
					honeypot = "https://honeypot.is/?address="
					dextools = "https://www.dextools.io/app/cn/bnb/pair-explorer/"
				} else if chain == "eth" {
					chainScan = "https://etherscan.io/address/"
					honeypot = "https://honeypot.is/ethereum?address="
					moonarch = "eth."
					dextools = "https://www.dextools.io/app/cn/ether/pair-explorer/"
				}
				s := fmt.Sprintf("*价格已上涨到监控位置: %s*\n\n*%s:$%s* \n*Chain:* %s | *Price:* $%s\n\n*5M:*  %0.2f%%  *:*  $%0.2f\n*1H:*  %0.2f%%  *:*  $%0.2f\n*6H:*  %0.2f%%  *:*  $%0.2f\n*1D:*  %0.2f%%  *:*  $%0.2f\n\n*Trade:* [dextools](%s%s) | [dexscreener](%s) | [ave.ai](https://ave.ai/token/%s-%s) | [dexview](https://www.dexview.com/%s/%s)\n\n`%s`\n\n*Check:* [ChainScan](%s%s) | [Moonarch](https://%smoonarch.app/token/%s) | [Honeypot](%s%s)", price, pair.BaseToken.Name, pair.BaseToken.Symbol, pair.ChainId, pair.PriceUsd, pair.PriceChange.M5, pair.Volume.M5, pair.PriceChange.H1, pair.Volume.H1, pair.PriceChange.H6, pair.Volume.H6, pair.PriceChange.H24, pair.Volume.H24, dextools, pair.PairAddress, pair.URL, pair.BaseToken.Addr, chain, chain, pair.BaseToken.Addr, pair.BaseToken.Addr, chainScan, pair.BaseToken.Addr, moonarch, pair.BaseToken.Addr, honeypot, pair.BaseToken.Addr)
				p.Meme <- s
				delete(p.memeHighTask, query+" "+chain)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *Probe) MemeDeclineMonitor(query string, chain string, price string) {
	test := p.api.MemePrice(query, chain)
	if test == nil {
		p.Meme <- "token查询失败,请检查token是否有误"
		return
	}
	t := time.NewTicker(time.Minute)
	ctx, cf := context.WithCancel(context.Background())
	p.memeLowTask[query+" "+chain] = cf
	p.Meme <- "开始下跌监控: \n" + query + " " + chain
	for {
		select {
		case <-t.C:
			line, err := strconv.ParseFloat(price, 64)
			if err != nil {
				p.Meme <- "输入的价格有误,无法识别"
				fmt.Println("价格转换异常：", err)
				delete(p.memeLowTask, query+" "+chain)
				return
			}
			pair := p.api.MemePrice(query, chain)
			if pair == nil {
				continue
			}
			now, err := strconv.ParseFloat(pair.PriceUsd, 64)
			if err != nil {
				p.Meme <- "价格转换异常,请检查日志"
				fmt.Println("价格转换异常：", err)
				delete(p.memeLowTask, query+" "+chain)
				return
			}
			if line >= now {
				chainScan := ""
				honeypot := ""
				moonarch := ""
				dextools := ""
				if chain == "bsc" {
					chainScan = "https://bscscan.com/address/"
					honeypot = "https://honeypot.is/?address="
					dextools = "https://www.dextools.io/app/cn/bnb/pair-explorer/"
				} else if chain == "eth" {
					chainScan = "https://etherscan.io/address/"
					honeypot = "https://honeypot.is/ethereum?address="
					moonarch = "eth."
					dextools = "https://www.dextools.io/app/cn/ether/pair-explorer/"
				}
				s := fmt.Sprintf("*价格已下跌到监控位置: %s*\n\n*%s:$%s* \n*Chain:* %s | *Price:* $%s\n\n*5M:*  %0.2f%%  *:*  $%0.2f\n*1H:*  %0.2f%%  *:*  $%0.2f\n*6H:*  %0.2f%%  *:*  $%0.2f\n*1D:*  %0.2f%%  *:*  $%0.2f\n\n*Trade:* [dextools](%s%s) | [dexscreener](%s) | [ave.ai](https://ave.ai/token/%s-%s) | [dexview](https://www.dexview.com/%s/%s)\n\n`%s`\n\n*Check:* [ChainScan](%s%s) | [Moonarch](https://%smoonarch.app/token/%s) | [Honeypot](%s%s)", price, pair.BaseToken.Name, pair.BaseToken.Symbol, pair.ChainId, pair.PriceUsd, pair.PriceChange.M5, pair.Volume.M5, pair.PriceChange.H1, pair.Volume.H1, pair.PriceChange.H6, pair.Volume.H6, pair.PriceChange.H24, pair.Volume.H24, dextools, pair.PairAddress, pair.URL, pair.BaseToken.Addr, chain, chain, pair.BaseToken.Addr, pair.BaseToken.Addr, chainScan, pair.BaseToken.Addr, moonarch, pair.BaseToken.Addr, honeypot, pair.BaseToken.Addr)
				p.Meme <- s
				delete(p.memeLowTask, query+" "+chain)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (t *Probe) AddSmartAddr(addr string) {
	if _, ok := t.smartAddr[addr]; !ok {
		ctx, cf := context.WithCancel(context.Background())
		t.smartAddr[addr] = cf
		go t.SmartAddrProbe(ctx, addr)
		t.Meme <- fmt.Sprintf("已开启 %s 地址的监控", addr)
	}
}

func (t *Probe) DeleteSmartAddr(addr string) {
	if cf, ok := t.smartAddr[addr]; ok {
		cf()
		delete(t.smartAddr, addr)
		t.Meme <- fmt.Sprintf("已关闭 %s 地址的监控", addr)
	}
}

func (t *Probe) SmartAddr(addr string, offset string) {
	apiKey := goconf.VarStringOrDefault("", "crypto", "etherscan", "apiKey")
	if apiKey == "" {
		t.Meme <- "未读取到etherscan的apikey无法启动监控"
		return
	}
	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=%s&sort=desc&address=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, offset, addr, apiKey))
	if err != nil {
		fmt.Println("请求失败")
		t.Meme <- "etherscan请求失败"
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("读取body失败")
		t.Meme <- "读取body失败"
		return
	}
	scan := new(TokenTxResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		fmt.Println("json转换失败")
		t.Meme <- "json转换失败"
		return
	}

	if scan.Status != "1" {
		t.Meme <- "返回码不为1,检查地址是否正确"
		return
	}

	msg := strings.Builder{}
	msg.WriteString("探测到新买入地址有:")
	e := make(map[string]struct{})
	for _, v := range scan.Result {
		if v.TokenSymbol != "WETH" {
			if _, ok := e[v.ContractAddress]; !ok {
				e[v.ContractAddress] = struct{}{}
				msg.WriteString("\n[")
				msg.WriteString(v.TokenSymbol)
				msg.WriteString("](https://www.dextools.io/app/cn/ether/pair-explorer/")
				msg.WriteString(v.ContractAddress)
				msg.WriteString("): `")
				msg.WriteString(v.ContractAddress)
				msg.WriteString("`")
			}
		}
	}
	t.Meme <- msg.String()
}

func (t *Probe) SetSmartAddrProbeItv(itv string) {
	i, err := strconv.Atoi(itv)
	if err != nil {
		t.Meme <- "请输入[1,60]分钟"
		return
	}
	if i > 60 || i < 1 {
		t.Meme <- "间隔必须在1-60之间"
		return
	}

	t.smartItv = i
	for addr, cf := range t.smartAddr {
		cf()
		ctx, c := context.WithCancel(context.Background())
		t.smartAddr[addr] = c
		go t.SmartAddrProbe(ctx, addr)
	}
	t.Meme <- "修改完成"
}

func (t *Probe) SmartAddrProbe(ctx context.Context, addr string) {
	apiKey := goconf.VarStringOrDefault("", "crypto", "etherscan", "apiKey")
	if apiKey == "" {
		t.Meme <- "未读取到etherscan的apikey无法启动监控"
		delete(t.smartAddr, addr)
		return
	}

	if t.smartItv > 60 || t.smartItv < 1 {
		t.Meme <- "间隔必须在1-60之间"
		return
	}

	now := time.Now()
	m := time.Duration((60-now.Minute())%t.smartItv) * time.Minute
	if m == 0 {
		m = time.Duration(t.smartItv) * time.Minute
	}
	time.Sleep(m - time.Second)

	monitor := func() {
		url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=50&sort=desc&address=%s&apikey=%s"
		r, err := http.Get(fmt.Sprintf(url, addr, apiKey))
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
		cnt := 0
		msg := strings.Builder{}
		for _, v := range scan.Result {
			if v.TokenSymbol != "WETH" {
				if _, ok := t.smartBuys[addr][v.ContractAddress]; !ok {
					cnt++
					t.smartBuys[addr][v.ContractAddress] = struct{}{}
					msg.WriteString("\n[")
					msg.WriteString(v.TokenSymbol)
					msg.WriteString("](https://www.dextools.io/app/cn/ether/pair-explorer/")
					msg.WriteString(v.ContractAddress)
					msg.WriteString("): `")
					msg.WriteString(v.ContractAddress)
					msg.WriteString("`")
				}
			}
		}

		if msg.Len() != 0 {
			t.Meme <- "探测到新买入地址有:" + msg.String()
		}
	}

	if _, ok := t.smartBuys[addr]; !ok {
		t.smartBuys[addr] = make(map[string]struct{})
	}

	select {
	case <-ctx.Done():
		return
	default:
		monitor()
	}

	tk := time.NewTicker(time.Minute * time.Duration(t.smartItv))
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			monitor()
		}

	}
}

func (t *Probe) DumpCron() {
	h := time.NewTicker(time.Hour)
	for range h.C {
		t.DumpSmartAddrList()
	}
}

func (t *Probe) DumpSmartAddrList() {
	if len(t.smartBuys) == 0 {
		return
	}
	b, err := json.Marshal(t.smartBuys)
	if err != nil {
		fmt.Println("序列化失败:", err)
		t.Meme <- "序列化失败"
		return
	}

	if _, err := os.Stat(t.smartDumpPath); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.smartDumpPath, os.ModePerm) // 创建目录
		if err != nil {
			fmt.Println("创建本地文件失败")
			t.Meme <- "创建本地文件失败"
			return
		}
	}

	err = ioutil.WriteFile(t.smartDumpPath+"smart_dump.json", b, 0644)
	if err != nil {
		fmt.Println("dump文件创建/写入失败")
		t.Meme <- "dump文件创建/写入失败"
		return
	}

	t.Meme <- "已探测地址dump完成"

}


func recoverSmartAddrList() map[string]map[string]struct{} {
	dump := make(map[string]map[string]struct{})
	b, err := ioutil.ReadFile(goconf.VarStringOrDefault("/usr/local/share/aio/", "crypto", "etherscan", "path") + "smart_dump.json")
	if err != nil {
		fmt.Println("dump文件读取失败:", err)
		return dump
	}

	err = json.Unmarshal(b, &dump)
	if err != nil {
		fmt.Println("SmartAddrList数据读取失败:", err)
		return dump
	}

	return dump

}