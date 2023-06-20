package crypto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var (
	// apiUrl  = "https://api.binance.com"
	dataUrl      = "https://data.binance.com"
	fapiUrl      = "https://fapi.binance.com"
	memeUrl      = "https://api.dexscreener.com/latest/dex/search/?q="	
	memeCheckUrl = "https://api.gopluslabs.io/api/v1/token_security/%s?contract_addresses=%s"
	honeypotUrl  = "https://api.honeypot.is/v2/IsHoneypot?address="
)

type Crypto struct {
	apiKey    string
	secretKey string
	chainMap  map[string]string
}

func NewCrypto(api, secret string) *Crypto {
	return &Crypto{
		apiKey:    api,
		secretKey: secret,
		chainMap:  map[string]string{"ethereum": "1", "optimism": "10", "cronos": "25", "bsc": "56", "okc": "66", "gnosis": "100", "heco": "128", "polygon": "137", "fantom": "250", "kcc": "321", "zksync": "324", "ethw": "10001", "fon": "201022", "arbitrum": "42161", "avalanche": "43114", "linea": "59140", "harmony": "1666600000", "tron": "tron"},
	}
}

func (t *Crypto) Ping() bool {
	if _, err := http.Get(dataUrl + "/api/v3/ping"); err != nil {
		return false
	}

	return true
}

func (t *Crypto) Price(name ...string) (prices map[string]string) {
	prices = make(map[string]string)
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

	resBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return
	}

	price := []priceResp{}
	if err = json.Unmarshal(resBody, &price); err != nil {
		return
	}

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

	resBody, err := io.ReadAll(resp.Body)
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

	b, _ := io.ReadAll(resp.Body)

	var result [][]interface{}
	json.Unmarshal(b, &result)

	// 打印K线数据
	for _, kline := range result {
		end, _ := strconv.ParseFloat(kline[4].(string), 64)
		start, _ := strconv.ParseFloat(kline[1].(string), 64)
		if end-start >= 0 {
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

func (t *Crypto) MemeCheck(query string, chain string) *MemeChecker {
	c := strings.ToLower(chain)
	if c == "eth" {
		c = "ethereum"
	}

	url := fmt.Sprintf(memeCheckUrl, t.chainMap[c], query)

	r, err := http.Get(url)
	if err != nil {
		fmt.Println("请求失败：", err)
		return nil
	}

	meme := &MemeCheckerResp{}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println("Body读取失败", err)
		return nil
	}

	err = json.Unmarshal(b, &meme)
	if err != nil {
		fmt.Println("json转换失败", err)
		return nil
	}

	for _, v := range meme.MemeCheckers {
		lockedlp := 0.0
		for _, v := range v.LpHolders {
			if v.IsLocked == 1 {
				if f, err := strconv.ParseFloat(v.Percent, 64); err == nil {
					lockedlp += f
				}
			}	
		}
		v.LpLockedTotal = lockedlp
		return v
	}

	return nil
}

func (t *Crypto) HoneypotCheck(addr string) string {
	r, err := http.Get(honeypotUrl + addr)
	if err != nil {
		fmt.Println("honeypotUrl请求失败", err)
		return "Do your own research"
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println("body读取失败", err)
		return "Do your own research"
	}

	res := new(HoneypotResp)
	err = json.Unmarshal(b, &res)
	if err != nil {
		fmt.Println("json序列化失败", err)
		return "Do your own research"
	}

	if res.Honeypot.Is {
		return "HONEYPOT DETECTED!!! RUN THE FUCK AWAY!"
	}

	return "DOES NOT SEEM LIKE A HONEYPOT"
}


func (t *Crypto) WhetherHoneypot(addr string) bool {
	r, err := http.Get(honeypotUrl + addr)
	if err != nil {
		fmt.Println("honeypotUrl请求失败", err)
		return false
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println("body读取失败", err)
		return false
	}

	res := new(HoneypotResp)
	err = json.Unmarshal(b, &res)
	if err != nil {
		fmt.Println("json序列化失败", err)
		return false
	}

	return res.Honeypot.Is
}