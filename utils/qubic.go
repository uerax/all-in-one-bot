package utils

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Qubic struct {  
	AverageScore     float64          `json:"averageScore"`    
	EstimatedIts     int64            `json:"estimatedIts"`    
	SolutionsPerHour int64            `json:"solutionsPerHourCalculated"`
}

var (
	orgeApi = "https://tradeogre.com/api/v1"
	mexcApi = "https://www.mexc.com/open/api/v2"
	defaultToken = ""
	defaultIt = 1000
)

func init() {
	go QubicTokenAutoRefresh()
}

func (t *Utils) QubicProfit(token string) {
	if len(token) < 50 {
		i, err := strconv.Atoi(token)
		if err == nil {
			defaultIt = i
		}
		token = ""
	}
	it := defaultIt
	wait := sync.WaitGroup{}
	wait.Add(1)
	price := 0.0
	go func() {
		price = qubicPrice()
		wait.Done()
	}()
	qb, err := QubicInfo(token)
	if err != nil {
		t.ErrC <- err.Error()
		return
	}
	ep1, ep2 := 1035502957.0, 281213017.0

	now := time.Now()

	// totalScore := int(qb.AverageScore) * 676
	dayOfWeek := int(now.Weekday())
	earningPerHour := 0.0
	totalHours := 7 * 24
	if dayOfWeek < 3 {
		// 星期三晚上20点刷新，所以加4
		totalHours = 4 + (24 * (4 + dayOfWeek - 1)) + now.Hour()
	} else if dayOfWeek > 3 {
		totalHours = 4 + (24 * (dayOfWeek - 3 - 1)) + now.Hour()
	} else {
		if now.Hour() > 20 {
			totalHours = now.Hour() - 20
		} else {
			totalHours = 6 * 24 + now.Hour()
		}
	}
	earningPerHour = qb.AverageScore / float64(totalHours)
	// hoursUntilSunday := (7 * 24) - (dayOfWeek * 24 + now.Hour())
	totalEarning := float64(earningPerHour * (7 * 24))
	earn1, earn2 := ep1 / (totalEarning * 1.06), ep2 / (totalEarning * 1.06)

	sol := float64(it) * float64(qb.SolutionsPerHour) / float64(qb.EstimatedIts)

	msg := fmt.Sprintf("当前全网算力: *%d it/s*\n平均出块速度: *%d / h*\n当前平均分: *%.f*\n\n本周预计平均分: *%.f*\n\n%d算力预计1小时出块: *%.3f*\n%d算力预计1天出块: *%.3f*\n%d算力预计7天出块: *%.3f*\n\n%d算力当前预计出块: *%.3f*\n\n单个块预计总收益: *%.f qubic*\nEp1单块预计收益: *%.f qubic*\nEp2单块预计收益: *%.f qubic*\n\n纪元预计总收益: *%d qubic*\nEp1预计总收益: *%d qubic*\nEp2预计总收益: *%d qubic*", qb.EstimatedIts, qb.SolutionsPerHour, qb.AverageScore, totalEarning, it, sol, it, sol*24, it, sol*24*7, it, float64(totalHours)*sol, earn1 + earn2, earn1, earn2, int(sol * 24 * 7 * (earn1 + earn2)), int(sol * 24 * 7 * earn1), int(sol * 24 * 7 * earn2))

	wait.Wait()
	priceMsg := fmt.Sprintf("\n\n当前Qubic价格: *%.12f U*\n单个块预计收益: *%.3f U*\n纪元预计总收益: *%.3f U*\n\nEp1预计收益: *%.3f U*\nEp2预计收益: *%.3f U*", price, (earn1 + earn2)*price, (earn1 + earn2)*price*sol*24*7, earn1*price*sol*24*7, earn2*price*sol*24*7)

	t.MsgC <- msg + priceMsg
	
}

func QubicInfo(token string) (*Qubic, error) {
	
	url := "https://api.qubic.li/Score/Get"

	req, _ := http.NewRequest("GET", url, nil)

	if len(token) == 0 {
		token = defaultToken
	} else {
		defaultToken = token
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Host", "api.qubic.li")
	req.Header.Add("Origin", "https://app.qubic.li")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://app.qubic.li/")
	req.Header.Add("Sec-Fetch-Dest", "empty")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	reader := res.Body
	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(res.Body)
        if err != nil {
            return nil, err
        }
        defer reader.Close()
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	qb := Qubic{}
	err = json.Unmarshal(body, &qb)

	return &qb, err
}

func qubicPrice() float64 {

	url := "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest?id=29169"

	req, _ := http.NewRequest("GET", url, nil)

	//req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Thunder Client (https://www.thunderclient.com)")
	req.Header.Add("X-CMC_PRO_API_KEY", "2fd0cde2-ea61-4c5c-96df-ee34f6d6e256")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0
	}

	type Usd struct {
		Price float64 `json:"price"`
	}

	type Quote struct {
		Usd Usd `json:"USD"`
	}

	type Qb struct {
		Quote Quote `json:"quote"`
	}

	type Data struct {
		Qb Qb `json:"29169"`
	}

	type QbResp struct {
		Data Data `json:"data"`
	}

	qb := QbResp{}

	json.Unmarshal(body, &qb)
	
	return qb.Data.Qb.Quote.Usd.Price

}

func QubicTokenAutoRefresh() {
	refresh := func()  {
		type Token struct {
			Token        string `json:"token"`
			Success      bool   `json:"success"`
		}
	
		url := "https://api.qubic.li/Auth/Login"
	
		payload := strings.NewReader("{\n  \"password\": \"guest13@Qubic.li\",\n  \"userName\": \"guest@qubic.li\",\n  \"twoFactorCode\": \"\"\n}")
	
		req, _ := http.NewRequest("POST", url, payload)
	
		req.Header.Add("Content-Type", "application/json")
	
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return 
		}
		qb := Token{}
		err = json.Unmarshal(body, &qb)
		if err != nil || !qb.Success {
			return 
		}
	
		defaultToken = qb.Token
	}
	tk := time.NewTicker(12 * time.Hour)
	refresh()
	for range tk.C {
		go refresh()
	}
}

func (t *Utils) QubicToken() {
	type Token struct {
		Token        string `json:"token"`
		Success      bool   `json:"success"`
	}

	url := "https://api.qubic.li/Auth/Login"

	payload := strings.NewReader("{\n  \"password\": \"guest13@Qubic.li\",\n  \"userName\": \"guest@qubic.li\",\n  \"twoFactorCode\": \"\"\n}")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.ErrC <- err.Error()
		return 
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.ErrC <- err.Error()
		return 
	}
	qb := Token{}
	err = json.Unmarshal(body, &qb)
	if err != nil || !qb.Success {
		t.ErrC <- err.Error()
		return 
	}

	defaultToken = qb.Token
	t.MsgC <- "Token 刷新成功"
}


type Orge struct {
	Success      bool   `json:"success"`
	Initialprice string `json:"initialprice"`
	Price        string `json:"price"`
	High         string `json:"high"`
	Low          string `json:"low"`
	Volume       string `json:"volume"`
	Bid          string `json:"bid"`
	Ask          string `json:"ask"`
}

type MexcResp struct {
	Code int64   `json:"code"`
	Data []Mexc `json:"data"`
}

type Mexc struct {
	Symbol     string `json:"symbol"`
	Volume     string `json:"volume"`
	Amount     string `json:"amount"`
	High       string `json:"high"`
	Low        string `json:"low"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Open       string `json:"open"`
	Last       string `json:"last"`
	Time       int64  `json:"time"`
	ChangeRate string `json:"change_rate"`
}

func MexcPrice(coin string) float64 {
	url := mexcApi + "/market/ticker?symbol=" + coin + "_usdt"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0
	}

	price := MexcResp{}

	err = json.Unmarshal(body, &price)
	if err != nil || price.Code != 200 || len(price.Data) == 0 {
		return 0
	}
	p, _ := strconv.ParseFloat(price.Data[0].Last, 64)
	return p
	
}

func OrgePrice(coin string) float64 {
	url := orgeApi + "/ticker/" + coin + "-usdt"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0
	}

	price := Orge{}

	err = json.Unmarshal(body, &price)
	if err != nil {
		return 0
	}

	p, _ := strconv.ParseFloat(price.Price, 64)

	return p
}

func (t *Utils) Compare(usdt string) {
	u := 500.0
	if usdt != "" {
		if t, err := strconv.ParseFloat(usdt, 64); err == nil {
			u = t
		}
	}
	bn := t.bn.Price(`LTCUSDT","XRPUSDT","BTCUSDT`)
	bn_ltc, _ := strconv.ParseFloat(bn["LTCUSDT"], 64)
	bn_xrp, _ := strconv.ParseFloat(bn["XRPUSDT"], 64)
	bn_btc, _ := strconv.ParseFloat(bn["BTCUSDT"], 64)
	
	ltc, xrp, pyi, xmr, btc := 0.0, 0.0, 0.0, 0.0, 0.0
	
	mexc_ltc, mexc_pyi, mexc_xmr := 0.0, 0.0, 0.0
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func ()  {
		ltc, xrp, pyi, xmr, btc = OrgePrice("ltc"),OrgePrice("xrp"),OrgePrice("pyi"),OrgePrice("xmr"),OrgePrice("btc")
		wg.Done()
	}()
	go func ()  {
		mexc_ltc, mexc_pyi, mexc_xmr = MexcPrice("ltc"), MexcPrice("pyi"), MexcPrice("xmr")
		wg.Done()
	}()
	wg.Wait()

	bn_ltc_loss := u - (u / ltc * bn_ltc)
	bn_xrp_loss := u - (u / xrp * bn_xrp)
	bn_btc_loss := u - (u / btc * bn_btc)

	mexc_ltc_loss := u - (u / ltc * mexc_ltc)
	mexc_pyi_loss := u - (u / pyi * mexc_pyi)
	mexc_xmr_loss := u - (u / xmr * mexc_xmr)

	msg := fmt.Sprintf("%.1f U 从Orge提现手续费分析:\n\nLTC 提现 Binance 损耗: %.3f U + %.3f U\n总计损耗为: %.3f U\nLTC 提现 Mexc 损耗: %.4f U + 1 U + %.3f U\n总计损耗为: %.3f U\nOgre 价格: %.5f\nBinance 价格: %.5f\nMexc 价格: %.5f\n\nXRP 提现 Binance 损耗: %.3f U + %.3f U\n总计损耗为: %.3f U\nOgre 价格: %.5f\nBinance 价格: %.5f\n\nPYI 提现 Mexc 损耗: %.2f U + 1 U + %.2f U\n总计损耗为: %.2f U\nOgre 价格: %.5f\nMexc 价格: %.5f\n\nXMR 提现 Mexc 损耗: %.3f U + 1 U + %.3f U\n总计损耗为: %.4f U\nOgre 价格: %.5f\nMexc 价格: %.5f\n\nBTC 提现 Binance 损耗: %.4f U + %.4f U\n总计损耗为: %.4f U\nOgre 价格: %.2f\nBinance 价格: %.2f", u, bn_ltc_loss, 0.001*ltc, bn_ltc_loss+(0.001*ltc), mexc_ltc_loss, 0.001*ltc, mexc_ltc_loss+(0.001*ltc) + 1, ltc, bn_ltc, mexc_ltc, bn_xrp_loss, 0.001*xrp,bn_xrp_loss+(0.001*xrp), xrp, bn_xrp, mexc_pyi_loss, pyi * 5, mexc_pyi_loss + (pyi * 5) + 1, pyi, mexc_pyi, mexc_xmr_loss, 0.00021917*xmr, mexc_xmr_loss+(0.00021917*xmr)+1, xmr, mexc_xmr,bn_btc_loss, 0.00011225*btc, bn_btc_loss + (0.00011225*btc), btc, bn_btc)

	t.MsgC <- msg

}