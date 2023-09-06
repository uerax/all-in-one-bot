package crypto

import (
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

	"github.com/uerax/goconf"
)

var (
	// apiUrl  = "https://api.binance.com"
	dataUrl          = "https://data.binance.com"
	fapiUrl          = "https://fapi.binance.com"
	memeUrl          = "https://api.dexscreener.com/latest/dex/search/?q="
	memeCheckUrl     = "https://api.gopluslabs.io/api/v1/token_security/%s?contract_addresses=%s"
	honeypotUrl      = "https://api.honeypot.is/v2/IsHoneypot?address="
	dextoolsUrl      = "https://www.dextools.io/shared/data/pair?chain=%s&address=%s"
	uniswapUrl       = "https://etherscan.io/tradingview/uniswap%s/%s/history?fromTs=%d&toTs=%d&resolution=%d&last=%d"
	honeypotPairsUrl = "https://api.honeypot.is/v1/GetPairs?chainID=1&address="
)

var (
	crypto     *Crypto
	onceCrypto sync.Once
)

type Crypto struct {
	apiKey    string
	secretKey string
	chainMap  map[string]string
	pairsMap  map[string]map[string]*Pair
	pairsPath string
}

func NewCrypto() *Crypto {
	onceCrypto.Do(func() {
		crypto = &Crypto{
			apiKey:    goconf.VarStringOrDefault("", "crypto", "binance", "apiKey"),
			secretKey: goconf.VarStringOrDefault("", "crypto", "binance", "secretKey"),
			chainMap:  map[string]string{"ethereum": "1", "optimism": "10", "cronos": "25", "bsc": "56", "okc": "66", "gnosis": "100", "heco": "128", "polygon": "137", "fantom": "250", "kcc": "321", "zksync": "324", "ethw": "10001", "fon": "201022", "arbitrum": "42161", "avalanche": "43114", "linea": "59140", "harmony": "1666600000", "tron": "tron"},
			pairsPath: goconf.VarStringOrDefault("/usr/local/share/aio/", "crypto", "etherscan", "path"),
			pairsMap:  recoverPairsMap(),
		}
		go crypto.CronDumpPairsMap()
	})
	return crypto
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
		log.Println("请求失败：", err)
		return
	}

	resp, err := http.DefaultClient.Do(res)
	if err != nil {
		log.Println("请求失败：", err)
		return
	}

	resBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println("body读取失败：", err)
		return
	}

	price := priceResp{}
	if err = json.Unmarshal(resBody, &price); err != nil {
		log.Println("body转换为结构体失败", err)
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
		log.Println("请求失败：", err)
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

func (t *Crypto) Dexscreener(query string) map[string]*Pair {
	r, err := http.Get(memeUrl + query)
	if err != nil {
		log.Println("请求失败：", err)
		return nil
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败：", err)
		return nil
	}

	meme := new(Meme)
	err = json.Unmarshal(b, &meme)
	if err != nil {
		log.Println("json转换失败: ", err)
		return nil
	}

	m := make(map[string]*Pair)

	for _, v := range meme.Pairs {
		if v.DexId == "uniswap" && v.QuoteToken != nil {
			v.CreateTime = time.Unix(v.CreateAt/1000, 0).Format("2006-01-02 15:04:05")
			for i := range v.Labels {
				if strings.EqualFold(v.QuoteToken.Symbol, "WETH") {
					m[v.Labels[i]] = v
				}
			}
		}
	}

	return m

}

func (t *Crypto) HoneypotPairs(query string) map[string]*Pair {

	meme := make([]*HoneypotPairs, 0)
	req, err := http.NewRequest(http.MethodGet, honeypotPairsUrl+query, nil)
	if err != nil {
		log.Println("请求失败：", err)
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://honeypot.is/")
	req.Header.Set("Origin", "https://honeypot.is")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("请求失败：", err)
		return nil
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败：", err)
		return nil
	}

	err = json.Unmarshal(b, &meme)
	if err != nil {
		log.Println("json转换失败: ", err)
		return nil
	}

	m := make(map[string]*Pair)

	for _, v := range meme {
		if v.Pairs != nil && strings.Contains(v.Pairs.Name, "Uniswap V2") && strings.Contains(v.Pairs.Name, "WETH") {
			pair := new(Pair)
			pair.CreateTime = time.Unix(v.CreatedAtTimestamp, 0).Format("2006-01-02 15:04:05")
			pair.PriceUsd = "0"
			pair.PairAddress = v.Pairs.Address
			pair.Lp = &Liquidity{v.Liquidity}
			m["v2"] = pair

		}
		if v.Pairs != nil && strings.Contains(v.Pairs.Name, "Uniswap V3") && strings.Contains(v.Pairs.Name, "WETH") {
			pair := new(Pair)
			pair.CreateTime = time.Unix(v.CreatedAtTimestamp, 0).Format("2006-01-02 15:04:05")
			pair.PriceUsd = "0"
			pair.PairAddress = v.Pairs.Address
			pair.Lp = &Liquidity{v.Liquidity}
			m["v3"] = pair
		}
	}

	return m

}

func (t *Crypto) MemePrice(query string, chain string) *Pair {
	c := strings.ToLower(chain)
	if c == "eth" {
		c = "ethereum"
	}

	meme := new(Meme)

	r, err := http.Get(memeUrl + query)
	if err != nil {
		log.Println("请求失败：", err)
		return nil
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败：", err)
		return nil
	}

	err = json.Unmarshal(b, &meme)
	if err != nil {
		log.Println("json转换失败: ", err)
		return nil
	}

	for _, v := range meme.Pairs {
		if v.DexId == "uniswap" {
			v.CreateTime = time.Unix(v.CreateAt/1000, 0).Format("2006-01-02 15:04:05")
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
		log.Println("请求失败：", err)
		return nil
	}

	meme := &MemeCheckerResp{}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("Body读取失败", err)
		return nil
	}

	err = json.Unmarshal(b, &meme)
	if err != nil {
		log.Println("json转换失败", err)
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
	req, err := http.NewRequest(http.MethodGet, honeypotUrl+addr, nil)
	if err != nil {
		log.Println("request创建失败", err)
		return "Do your own research"
	}
	req.Header.Set("Referer", "https://honeypot.is/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("honeypotUrl请求失败", err)
		return "Do your own research"
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败", err)
		return "Do your own research"
	}

	res := new(HoneypotResp)
	err = json.Unmarshal(b, &res)
	if err != nil {
		log.Println("json序列化失败", err)
		return "Do your own research"
	}

	if res.Honeypot == nil || res.Honeypot.Is {
		return "HONEYPOT DETECTED!!! RUN THE FUCK AWAY!"
	}

	return "DOES NOT SEEM LIKE A HONEYPOT"
}

func (t *Crypto) WhetherHoneypot(addr string) bool {
	req, err := http.NewRequest(http.MethodGet, honeypotUrl+addr, nil)
	if err != nil {
		log.Println("request创建失败", err)
		return false
	}
	req.Header.Set("Referer", "https://honeypot.is/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("honeypotUrl请求失败", err)
		return false
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败", err)
		return false
	}

	res := new(HoneypotResp)
	err = json.Unmarshal(b, &res)
	if err != nil {
		log.Println("json序列化失败", err)
		return false
	}

	if res.Honeypot == nil {
		return true
	}

	return res.Honeypot.Is
}

func (t *Crypto) DexTools(pair, chain string) *Datum {
	c := strings.ToLower(chain)
	if c == "eth" {
		c = "ether"
	}

	url := fmt.Sprintf(dextoolsUrl, chain, pair)

	r, err := http.Get(url)
	if err != nil {
		log.Println("请求失败：", err)
		return nil
	}

	meme := &DextoolsResp{}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("Body读取失败", err)
		return nil
	}

	err = json.Unmarshal(b, &meme)
	if err != nil {
		log.Println("json转换失败", err)
		return nil
	}

	if len(meme.Data) < 1 {
		return nil
	}

	return &meme.Data[0]
}

func (t *Crypto) IsHoneypot(addr string) *HoneypotResp {
	req, err := http.NewRequest(http.MethodGet, honeypotUrl+addr, nil)
	if err != nil {
		log.Println("request创建失败", err)
		return nil
	}
	req.Header.Set("Referer", "https://honeypot.is/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("honeypotUrl请求失败", err)
		return nil
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败", err)
		return nil
	}

	res := new(HoneypotResp)
	err = json.Unmarshal(b, &res)
	if err != nil {
		log.Println("json序列化失败", err)
		return nil
	}

	if res.Honeypot == nil {
		res.Honeypot = &Honeypot{
			true,
		}
	}

	return res
}

func (t *Crypto) DexKline(pair string, start, end int64, resolution int, last int64, version string) *DexKline {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(uniswapUrl, version, pair, start, end, resolution, last), nil)
	if err != nil {
		log.Println(err)
		return nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	req.Header.Set("Referer", fmt.Sprintf("https://etherscan.io/dex/uniswap%s/%s", version, pair))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("honeypotUrl请求失败", err)
		return nil
	}
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败", err)
		return nil
	}

	res := new(DexKline)
	err = json.Unmarshal(b, &res)
	if err != nil {
		log.Println("json序列化失败", err)
		return nil
	}
	return res
}

func (t *Crypto) Pairs(token string) map[string]*Pair {
	token = strings.ToLower(token)
	if v, ok := t.pairsMap[token]; ok {
		return v
	}
	//p := t.Dexscreener(token)
	p := t.HoneypotPairs(token)
	t.pairsMap[token] = p
	return p
}

func (t *Crypto) CronDumpPairsMap() {
	tick := time.NewTicker(time.Hour)
	for range tick.C {
		t.DumpPairsMap()
	}
}

func (t *Crypto) DumpPairsMap() {
	if len(t.pairsMap) == 0 {
		return
	}
	b, err := json.Marshal(t.pairsMap)
	if err != nil {
		log.Println("序列化失败:", err)
		return
	}

	if _, err := os.Stat(t.pairsPath); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.pairsPath, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地文件夹失败")
			return
		}
	}
	err = os.WriteFile(t.pairsPath+"pairs_dump.json", b, 0644)
	if err != nil {
		log.Println("dump文件创建/写入失败")
		return
	}
}

func recoverPairsMap() map[string]map[string]*Pair {
	dump := make(map[string]map[string]*Pair)
	b, err := os.ReadFile(goconf.VarStringOrDefault("/usr/local/share/aio/", "crypto", "etherscan", "path") + "pairs_dump.json")
	if err != nil {
		return dump
	}

	err = json.Unmarshal(b, &dump)
	if err != nil {
		return dump
	}

	return dump
}
