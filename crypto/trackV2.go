package crypto

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/uerax/all-in-one-bot/common"
)

// Buy Return ETH Balance
// Sell Return Balance ETH
func (t *Track) getEthByHtml(hash, symbol string) []float64 {
	res, err := http.Get("https://etherscan.io/tx/" + hash)
	val := make([]float64, 2)
	if err != nil {
		log.Println(err)
		return val
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return val
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return val
	}

	eth, coin := 0.0, 0.0
	pre := ""

	doc.Find(".far.fa-bolt.fa-fw.text-primary.me-1").Parent().Parent().Find(".me-1").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := s.Text()
		if title == "" {
			return
		}
		title = strings.ReplaceAll(title, ",", "")
		//fmt.Println(i, " " + title)

		if strings.EqualFold(title, "Ether") {
			if len(title)-strings.Index(title, ".") > 4 {
				title = title[:strings.Index(title, ".")+4]
			}
			bae, err := strconv.ParseFloat(pre, 64)
			if err == nil {
				eth += bae
			}
		}
		if strings.EqualFold(title, symbol) {
			bae, err := strconv.ParseFloat(pre, 64)
			if err == nil {
				bae = math.Round(bae*1e5) / 1e5
				coin += bae
			}
		}
		pre = title
	})

	if strings.Contains(doc.Find(".far.fa-bolt.fa-fw.text-primary.me-1").Parent().Parent().Find("a").Text(), "USDT") {
		val[0] = 0
		val[1] = 0
		return val
	}

	val[0] = eth
	val[1] = coin
	return val
}

func (t *Track) WalletTrackingV2(addr string) {
	now := time.Now()
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
			log.Println("Panic Addr:", addr)
		}
	}()

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

	if strings.EqualFold(t.Newest[addr].Hash, scan.Result[0].Hash) {
		return
	}

	// 首次加入探测忽略
	if t.Newest[addr].Hash == "" {
		t.Newest[addr].Hash = scan.Result[0].Hash
		late, err := strconv.ParseInt(scan.Result[0].TimeStamp, 10, 64)
		if err == nil {
			t.Newest[addr].Latest = time.Unix(late, 0).Format("2006-01-02 15:04:05")
		}
		return
	}

	t.Newest[addr].Hash = scan.Result[0].Hash
	late, err := strconv.ParseInt(scan.Result[0].TimeStamp, 10, 64)
	if err == nil {
		t.Newest[addr].Latest = time.Unix(late, 0).Format("2006-01-02 15:04:05")
	}

	wg := sync.WaitGroup{}
	sb := strings.Builder{}

	record := scan.Result[0]

	if strings.EqualFold(record.TokenSymbol, "WETH") || isNull(record.From) || isNull(record.To) {
		return
	}

	// 卖单仅提示
	if strings.EqualFold(record.From, addr) {
		sb.WriteString("\n*")
		sb.WriteString(record.TokenName)
		sb.WriteString(": *")
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
		sb.WriteString("----[前往出售](https://app.uniswap.org/#/swap?exactField=input&inputCurrency=")
		sb.WriteString(record.ContractAddress)
		sb.WriteString("&outputCurrency=ETH&chain=ethereum)")
		sb.WriteString("\n")
		sb.WriteString("\n`")
		sb.WriteString(record.ContractAddress)
		sb.WriteString("`")
		t.C <- strings.ToUpper(t.Newest[addr].Remark) + ": `" + addr + "` [Selling](https://etherscan.io/tx/" + record.Hash + ")" + sb.String()
		return
	}

	balance := 0.0
	isHoneypot := ""
	detail := ""
	check := ""
	link := ""
	tax := ""
	count := ""

	getBalance := func() {
		defer wg.Done()
		if strings.EqualFold(record.From, addr) {
			balance += t.getSellEthByHash(record.Hash, addr)
			count = record.Value
			// eth := t.getEthByHtml(record.Hash, false)
			// balance += eth[1]
			// count = fmt.Sprintf("%f", eth[0])
		} else {
			balance += t.getBuyEthByHash(record.Hash)
			count = record.Value
			// eth := t.getEthByHtml(record.Hash, true)
			// balance += eth[0]
			// count = fmt.Sprintf("%f", eth[1])
		}
		log.Println("getBalance耗时: ", time.Since(now))
	}

	// -1s
	getHoneypot := func() {
		defer wg.Done()
		hr := t.api.IsHoneypot(record.ContractAddress)
		if hr == nil {
			return
		}
		if hr.Honeypot.Is {
			isHoneypot += "*[SCAM]*"
		}
		ratio := 0.0

		if hr.SimulationResult.SellTax != 100 && hr.SimulationResult.BuyTax != 100 {
			ratio = 1 / ((1 - hr.SimulationResult.BuyTax/100) * (1 - hr.SimulationResult.SellTax/100))
		}
		tax += fmt.Sprintf("\n*Buy Tax: %.1f%%   |   Sell Tax: %.1f%%   |   Ratio: %.2f*", hr.SimulationResult.BuyTax, hr.SimulationResult.SellTax, ratio)
		log.Println("getHoneypot耗时: ", time.Since(now))
	}

	getDetail := func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic:", r)
				log.Println("Panic Addr:", addr)
			}
		}()

		pair := t.api.MemePrice(record.ContractAddress, "eth")
		if pair != nil {
			detail += fmt.Sprintf("\n\n*Price: $%s (%d)*", pair.PriceUsd, zeroCal(pair.PriceUsd))
			if pair.Lp != nil {
				detail += fmt.Sprintf("   |   *Pool: $%0.5f*", pair.Lp.Usd)
			}
			//detail += fmt.Sprintf("\n*CreationTime: %s*", pair.CreateTime)
			detail += fmt.Sprintf("\n\n*5M:    %0.2f%%    $%0.2f    %d/%d*\n*1H:    %0.2f%%    $%0.2f    %d/%d*\n*6H:    %0.2f%%    $%0.2f    %d/%d*\n*1D:    %0.2f%%    $%0.2f    %d/%d*", pair.PriceChange.M5, pair.Volume.M5, pair.Txns.M5.B, pair.Txns.M5.S, pair.PriceChange.H1, pair.Volume.H1, pair.Txns.H1.B, pair.Txns.H1.S, pair.PriceChange.H6, pair.Volume.H6, pair.Txns.H6.B, pair.Txns.H6.S, pair.PriceChange.H24, pair.Volume.H24, pair.Txns.H24.B, pair.Txns.H24.S)
			log.Println("getDetail耗时: ", time.Since(now))
		}
	}

	getCheck := func() {
		defer wg.Done()
		ck := t.api.MemeCheck(record.ContractAddress, "eth")
		if ck != nil {
			check += fmt.Sprintf("*Locked LP: %0.2f%%*\n*Owner:* `%s`\n[Creator](https://etherscan.io/address/%s) *: Percent: %s*", ck.LpLockedTotal*100.0, ck.OwnerAddress, ck.CreatorAddress, ck.CreatorPercent)
		}
		log.Println("getCheck耗时: ", time.Since(now))
	}

	getLink := func() {
		defer wg.Done()
		links := getLinks(t.getSourceCode(record.ContractAddress))
		if v, ok := links["Website"]; ok {
			link += fmt.Sprintf("[%s](%s)   ", "Website", v)
		}
		if v, ok := links["Twitter"]; ok {
			link += fmt.Sprintf("[%s](%s)   ", "Twitter", v)
		}
		if v, ok := links["Telegram"]; ok {
			link += fmt.Sprintf("[%s](%s)   ", "Telegram", v)
		}
		log.Println("getLink耗时: ", time.Since(now))
	}

	wg.Add(5)

	// 并发减少等待时间
	go getBalance()
	go getHoneypot()
	go getDetail()
	go getCheck()
	go getLink()

	wg.Wait()

	sb.WriteString("\n")
	sb.WriteString(isHoneypot)
	sb.WriteString("*")
	sb.WriteString(record.TokenName)
	sb.WriteString(": *")
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
	sb.WriteString(" ETH   |   ")
	sb.WriteString("Count: ")
	sb.WriteString(count)
	sb.WriteString(" ")
	sb.WriteString(record.TokenSymbol)
	sb.WriteString("*")
	sb.WriteString(detail)
	if link != "" {
		sb.WriteString("\n\n")
		sb.WriteString(link)
	}
	sb.WriteString("\n")
	sb.WriteString(tax)
	sb.WriteString("\n")
	sb.WriteString(check)

	log.Println("查询总耗时: ", time.Since(now))

	t.C <- "*" + strings.ToUpper(t.Newest[addr].Remark) + ":* `" + addr + "` [Buying](https://etherscan.io/tx/" + record.Hash + ")" + sb.String()
}

func (t *Track) SmartAddrFinderV2(token, offset, page string) []string {
	if t.Keys.IsNull() {
		t.C <- "未读取到etherscan的apikey无法启动分析"
		return nil
	}

	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=%s&offset=%s&sort=asc&contractaddress=%s&apikey=%s"
	scan := new(TokenTxResp)
	err := common.HttpGet(fmt.Sprintf(url, page, offset, token, t.Keys.GetKey()), &scan)
	if err != nil {
		log.Println("请求失败: ", err)
		return nil
	}

	if scan.Status != "1" {
		return nil
	}

	recorded := sync.Map{}

	// addr -> txs
	analyze := sync.Map{}

	handle := func(address string, sw *sync.WaitGroup) {
		defer sw.Done()
		address = strings.ToLower(address)
		if isNull(address) {
			return
		}
		if _, ok := recorded.Load(address); !ok {
			recorded.Store(address, struct{}{})
			his := make(map[string]struct{})
			list := t.TransferList(address, token)
			if len(list) == 0 {
				return
			}
			tmp := new(txs)
			for _, tx := range list {
				if _, ok := his[tx.Hash]; !ok {
					his[tx.Hash] = struct{}{}
					if strings.EqualFold(tx.From, address) {
						eth := t.getEthByHtml(tx.Hash, tx.TokenSymbol)
						val := eth[0]
						cnt := eth[1]
						tmp.Profit += val
						tmp.Sell += cnt
						// val := t.getSellEthByHash(tx.Hash, address)
						// tmp.Profit += val
					} else {
						eth := t.getEthByHtml(tx.Hash, tx.TokenSymbol)
						val := eth[0]
						cnt := eth[1]
						tmp.Profit -= val
						tmp.Buy += cnt
						tmp.Pay += val
						// val := t.getBuyEthByHash(tx.Hash)
						// tmp.Profit -= val
						// tmp.Pay += val
					}
				}
			}
			analyze.Store(address, tmp)
		}
	}

	wg := sync.WaitGroup{}
	lens := len(scan.Result)
	wg.Add(lens * 2)
	for i, v := range scan.Result {
		go handle(v.From, &wg)
		go handle(v.To, &wg)
		if i%(3*t.Keys.Len()) == 0 {
			time.Sleep(time.Second)
		}
	}

	start, end := "", ""

	ts, err := strconv.ParseInt(scan.Result[0].TimeStamp, 10, 64)
	if err == nil {
		start = time.Unix(ts, 0).Format("2006-01-02 15:04:05")
	}
	ts, err = strconv.ParseInt(scan.Result[lens-1].TimeStamp, 10, 64)
	if err == nil {
		end = time.Unix(ts, 0).Format("2006-01-02 15:04:05")
	}

	wg.Wait()

	list := make([]string, 0)
	msg := fmt.Sprintf("*合约地址:* `%s`\n*分析完毕: (%s -- %s)*", token, start, end)
	analyze.Range(func(k, value any) bool {
		v := value.(*txs)
		if v.Profit > 0 {
			list = append(list, k.(string))
			if len(msg) > 4000 {
				t.C <- msg
				msg = "---------------切割线---------------"
			}
			msg += fmt.Sprintf("\n`%s`*: %0.3f / %0.3f*", k, v.Pay, v.Profit)
		}
		return true
	})

	t.C <- msg
	return list
}

func (t *txs) Add(val float64) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	t.Profit += val
}

func (t *txs) Sub(val float64) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	t.Profit -= val
	t.Pay += val
}

func (t *Track) WalletTxAnalyzeV2(addr string, offset string) {
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

	recorded := sync.Map{}
	analyze := sync.Map{}
	profit := new(txs)
	handle := func(token string, wg *sync.WaitGroup) {
		defer wg.Done()
		if strings.EqualFold(token, "0x29480f9385de5f1e7084c2c09167a155d1285ccc") {
			// USDT
			return
		}
		if strings.EqualFold(token, "0xdac17f958d2ee523a2206206994597c13d831ec7") {
			// USDT
			return
		}
		if strings.EqualFold(token, "0x3579781bcfefc075d2cb08b815716dc0529f3c7d") {
			// ETH
			return
		}
		if _, ok := recorded.Load(token); !ok {
			recorded.Store(token, struct{}{})
			list := t.TransferList(addr, token)
			if len(list) == 0 {
				return
			}
			tmp := new(txs)
			his := make(map[string]struct{})
			for _, tx := range list {
				if _, ok := his[strings.ToLower(tx.Hash)]; ok {
					continue
				}
				his[strings.ToLower(tx.Hash)] = struct{}{}
				if strings.EqualFold(tx.From, addr) {
					// eth := t.getEthByHtml(tx.Hash, tx.TokenSymbol)
					// tmp.Profit += eth[0]
					// tmp.Sell += eth[1]
					// profit.Add(eth[0])
					val := t.getSellEthByHash(tx.Hash, addr)
					tmp.Profit += val
					profit.Add(val)
				} else {
					// eth := t.getEthByHtml(tx.Hash, tx.TokenSymbol)
					// tmp.Profit -= eth[0]
					// tmp.Buy += eth[1]
					// tmp.Pay += eth[0]
					// profit.Sub(eth[0])
					val := t.getBuyEthByHash(tx.Hash)
					tmp.Profit -= val
					tmp.Pay += val
					profit.Sub(val)
				}
				if tmp.Time == "" {
					ts, err := strconv.ParseInt(tx.TimeStamp, 10, 64)
					if err == nil {
						tmp.Time = time.Unix(ts, 0).Format("2006-01-02 15:04:05")
					}
				}
				tmp.Symbol = tx.TokenSymbol
			}
			analyze.Store(token, tmp)
		}

	}

	wg := sync.WaitGroup{}
	wg.Add(len(scan.Result))
	for i, record := range scan.Result {
		go handle(record.ContractAddress, &wg)
		if i % (2 * t.Keys.Len()) == 0 {
			time.Sleep(time.Second)
		}
	}

	wg.Wait()

	msg := fmt.Sprintf("[Wallet](https://etherscan.io/address/%s#tokentxns)*支出: %0.5f | 净收入: %0.5f*\n", addr, profit.Pay, profit.Profit)
	analyze.Range(func(k, value any) bool {
		v := value.(*txs)
		if len(msg) > 4000 {
			t.C <- msg
			msg = "---------------切割线---------------\n"
		}
		msg += fmt.Sprintf("[%s](https://www.dextools.io/app/cn/ether/pair-explorer/%s)*:* `%s`\n*T: %s | C: %0.3f | P: %0.3f *\n", v.Symbol, k, k, v.Time, v.Pay, v.Profit)
		return true
	})

	t.C <- msg

}
