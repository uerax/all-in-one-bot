package crypto

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	apiUrl  = "https://api.binance.com"
	dataUrl = "https://data.binance.com"
	fapiUrl = "https://fapi.binance.com"
	memeUrl = "https://api.dexscreener.com/latest/dex/search/?q="
)

type Crypto struct {
	apiKey    string
	secretKey string
}

func NewCrypto(api, secret string) *Crypto {
	return &Crypto{
		apiKey: api,
		secretKey: secret,
	}
}

func (t *Crypto) Ping() bool {
	if _, err := http.Get(dataUrl + "/api/v3/ping"); err != nil {
		return false
	}

	return true
}

func (t *Crypto) Price(name ...string) (prices map[string]string) {
	if len(name) == 0 {
		return 
	}
	symbols := fmt.Sprintf(`/api/v3/ticker/price?symbols=["%s"]`, strings.Join(name, `","`))
	res, err := http.NewRequest(http.MethodGet, dataUrl+symbols, nil)
	if err != nil {
		return 
	}

	resp, err := http.DefaultClient.Do(res)
	if err != nil {
		return
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	
	price := []priceResp{}
	if err = json.Unmarshal(resBody, &price); err != nil {
		return
	}

	prices = make(map[string]string)
	for _, v := range price {
		prices[v.Symbol] = v.Price
	}

	return 
}

func (t *Crypto) FuturesPrice(name string) (prices string) {
	if len(name) == 0 {
		return 
	}
	symbols := fmt.Sprintf(`/fapi/v1/ticker/price?symbol=%s`, name)
	res, err := http.NewRequest(http.MethodGet, fapiUrl+symbols, nil)
	if err != nil {
		fmt.Println("请求失败：", err)
		return 
	}

	resp, err := http.DefaultClient.Do(res)
	if err != nil {
		fmt.Println("请求失败：", err)
		return
	}

	resBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("body读取失败：", err)
		return
	}
	
	price := priceResp{}
	if err = json.Unmarshal(resBody, &price); err != nil {
		fmt.Println("body转换为结构体失败", err)
		return
	}
	

	return price.Price 
}

// 前3根k线的涨跌结果,1 涨 -1 跌
func (t *Crypto) UFutureKline(interval string, limit int, symbol string) []int {
	url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", symbol, interval, limit)

	res := make([]int, 0, 5)

	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("请求失败：", err)
		return nil
	}
	defer resp.Body.Close()

	// 解析JSON数据
	
	b, _ := ioutil.ReadAll(resp.Body)

	var result [][]interface{}
	json.Unmarshal(b, &result)
	
	// 打印K线数据
	for _, kline := range result {
		end, _ := strconv.ParseFloat(kline[4].(string), 64)
		start, _ := strconv.ParseFloat(kline[1].(string), 64)
		if end - start >= 0 {
			res = append(res, 1)
		} else {
			res = append(res, -1)
		}
		
	}

	return res
}

func (t *Crypto) MemePrice(query string, chain string) *Pair {
	c := strings.ToLower(chain)
	if c == "eth" {
		c = "ethereum"
	}

	meme := new(Meme)

	r, err := http.Get(memeUrl + query)
	if err != nil {
		fmt.Println("请求失败：", err)
		return nil
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println("body读取失败：", err)
		return nil
	}

	err = json.Unmarshal(b, &meme)
	if err != nil {
		fmt.Println("json转换失败: ", err)
		return nil
	}

	for _, v := range meme.Pairs {
		if v.ChainId == c {
			return v
		}
	}

	return nil

}

