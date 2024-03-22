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
	SolutionsPerHour int64            `json:"solutionsPerHour"`
}

var (
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

	msg := fmt.Sprintf("当前全网算力: *%d it/s*\n当前出块速度: *%d / h*\n当前平均分: *%.f*\n\n本周预计平均分: *%.f*\n\n%d算力预计1小时出块: *%.3f*\n%d算力预计1天出块: *%.3f*\n%d算力预计7天出块: *%.3f*\n\n%d算力当前预计出块: *%.3f*\n\n单个块预计总收益: *%.f qubic*\nEp1单块预计收益: *%.f qubic*\nEp2单块预计收益: *%.f qubic*\n\n纪元预计总收益: *%d qubic*\nEp1预计总收益: *%d qubic*\nEp2预计总收益: *%d qubic*", qb.EstimatedIts, qb.SolutionsPerHour, qb.AverageScore, totalEarning, it, sol, it, sol*24, it, sol*24*7, it, float64(totalHours)*sol, earn1 + earn2, earn1, earn2, int(sol * 24 * 7 * (earn1 + earn2)), int(sol * 24 * 7 * earn1), int(sol * 24 * 7 * earn2))

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