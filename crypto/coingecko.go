package crypto

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/uerax/goconf"
)

type Coingecko struct {
	api string
	list map[string]float64
	price map[string]MarketData
}

func getList() map[string]float64 {
	
	// 下载JSON文件
	url := "https://raw.githubusercontent.com/uerax/all-in-one-bot/master/crypto/list/list.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	// 读取JSON文件内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("无法读取JSON文件:", err)
		return nil
	}

	// 解析JSON为Map
	filter := make(map[string]float64)
	json.Unmarshal(body, &filter)
	
	return filter
}

func NewCoingecko() *Coingecko {	
	return &Coingecko{
	     goconf.VarStringOrDefault("", "crypto", "coingecko"),
		 getList(),
		 make(map[string]MarketData),
	}
}

type CoingeckoResp struct {
    ID         string     `json:"id"`
    Symbol     string     `json:"symbol"`
    MarketData MarketData `json:"market_data"`
}

type MarketData struct {
    CurrentPrice   CurrentPrice `json:"current_price"`
    TotalPrice     float64      
    PriceChange24H float64      `json:"price_change_24h"`
}

type CurrentPrice struct {
    Usd float64 `json:"usd"`
}

func (t *Coingecko) Price(coin string, count float64) {
	url := "https://api.coingecko.com/api/v3/coins/" + coin
  	method := "GET"

  	client := &http.Client {}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", t.api)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	price := CoingeckoResp{}
	json.Unmarshal(body, &price)
	price.MarketData.TotalPrice = count * price.MarketData.CurrentPrice.Usd
	t.price[price.Symbol] = price.MarketData
}

func (t *Coingecko) SyncPrice() {
	for i := range t.list {
		t.Price(i, t.list[i])
	}
}
