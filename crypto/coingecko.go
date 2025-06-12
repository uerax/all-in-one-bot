package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/uerax/goconf"
)

type Coingecko struct {
	api string
	list map[string]float64
	price map[string]MarketData
	C      chan string
	ctx context.Context
	cancel context.CancelFunc
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
	     api: goconf.VarStringOrDefault("", "crypto", "coingecko"),
		 list: getList(),
		 price: make(map[string]MarketData),
		 C: make(chan string, 5),
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
	t.price = make(map[string]MarketData)
	for i := range t.list {
		time.Sleep(2 * time.Second)
		t.Price(i, t.list[i])
	}
}

func (t *Coingecko) Handle() {
	t.SyncPrice()
	if len(t.price) != 0 {
		msg := ""
		total := 0.0
		for k, v := range t.price {
			msg += fmt.Sprintf("\n%s 当前价格为 %fu 持有价值为 %fu", k, v.CurrentPrice.Usd, v.TotalPrice)
			total += v.TotalPrice
		}
		t.C <- fmt.Sprintf("当前总持有价值为 %fu%s", total, msg)
		println(fmt.Sprintf("当前总持有价值为 %fu%s", total, msg))
	}
}

func (t *Coingecko) Stop() {
	t.cancel()
}

func (t *Coingecko) Monitor() {
	if t.cancel != nil {
		t.C <- "已开启监控Bitcointalk新帖, 无需重复开启"
		return
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())
	ticker := time.NewTicker(24 * time.Hour)

	log.Println("开启定时监控持有币的价格")
	t.C <- "开启定时监控持有币的价格"

	for {
		select {
		case <-ticker.C:
			t.Monitor()
		case <-t.ctx.Done():
			log.Println("关闭定时监控Bitcointalk新帖")
			t.C <- "关闭定时监控Bitcointalk新帖"
			t.cancel = nil
			return
		}
	}
}