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
func (t *Track) getEthByHtml(hash string, buy bool) []float64 {
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
	fromIdx, toIdx := 2, 4
	if buy {
		toIdx++
	}

	from, to := 0.0, 0.0

	doc.Find(".far.fa-bolt.fa-fw.text-primary.me-1").Parent().Parent().Find("span").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := s.Text()
		title = strings.ReplaceAll(title, ",", "")
		//fmt.Println(i, " " + title)
		if i%8 == fromIdx {
			if len(title)-strings.Index(title, ".") > 4 {
				title = title[:strings.Index(title, ".")+4]
			}
			bae, err := strconv.ParseFloat(title, 64)
			if err == nil {
				if !buy {
					bae = math.Round(bae*1e5) / 1e5
				}
				from += bae
			}
		}
		if i%8 == toIdx {
			bae, err := strconv.ParseFloat(title, 64)
			if err == nil {
				if buy {
					bae = math.Round(bae*1e5) / 1e5
				}
				to += bae
			}
		}
	})

	if strings.Contains(doc.Find(".far.fa-bolt.fa-fw.text-primary.me-1").Parent().Parent().Find("a").Text(), "USDT") {
		return val
	}

	val[0] = from
	val[1] = to

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
		sb.WriteString("----[前往购买](https://app.uniswap.org/#/swap?exactField=input&exactAmount=0.02&inputCurrency=ETH&outputCurrency=")
		sb.WriteString(record.ContractAddress)
		sb.WriteString("&chain=ethereum)")
		sb.WriteString("\n\n`")
		sb.WriteString(record.ContractAddress)
		sb.WriteString("`")
		t.C <- strings.ToUpper(t.Newest[addr].Remark) + ":* `" + addr + "` [监控卖出](https://etherscan.io/tx/" + record.Hash + ")" + sb.String()
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

	t.C <- "*" + strings.ToUpper(t.Newest[addr].Remark) + ":* `" + addr + "` [监控买入](https://etherscan.io/tx/" + record.Hash + ")" + sb.String()
}

func (t *Track) SmartAddrFinderV2(token, offset, page string) {
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
							val = t.getEthByHtml(tx.Hash, false)[1]
							// val = t.getSellEthByHash(tx.Hash, address)
						} else {
							val = t.getEthByHtml(tx.Hash, true)[0]
							// val = t.getBuyEthByHash(tx.Hash)
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
					if analyze[address].Time == "" {
						ts, err := strconv.ParseInt(tx.TimeStamp, 10, 64)
						if err == nil {
							analyze[address].Time = time.Unix(ts, 0).Format("01-02 15:04:05")
						}
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
				msg += fmt.Sprintf("\n`%s`*: %s*\n*B:* %0.3f | *S:* %0.3f | *C:* %0.3f | *P:* %0.3f ETH", k, v.Time, v.Buy, v.Sell, v.Pay, v.Profit)
			}
		}
		t.C <- msg
	}
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