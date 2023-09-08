package crypto

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	pairsHandle *PairsHandle
	oncePairsApi sync.Once
)

type PairsHandle struct {
	Apis *PairsApi
	mu   sync.Mutex
}

func (t *PairsHandle) Pairs(query string) (map[string]*PairInfo, error) {
	m, err := t.Apis.Api.Call(query)
	t.mu.Lock()
	if err != nil {
		t.Apis = t.Apis.Next
		log.Println("Pairs Apis Next")
		t.mu.Unlock()
		m, err = t.Apis.Api.Call(query)
		return m, err
	}
	t.mu.Unlock()
	return m, err
}

type PairsApi struct {
	Api  PairsInterface
	Next *PairsApi
}

func InitPairsApi() *PairsApi {
	dexscreenerApi := &PairsApi{
		Api: new(Dexscreener),
	}
	honeypotPairApi := &PairsApi{
		Api: new(HoneypotPair),
	}
	honeypotPairApi.Next = dexscreenerApi
	dexscreenerApi.Next = honeypotPairApi
	return honeypotPairApi
}

func NewPairsHandle() *PairsHandle {
	oncePairsApi.Do(func() {
		pairsHandle = &PairsHandle{
			Apis: InitPairsApi(),
		}
	})
	return pairsHandle
}

type PairsInterface interface {
	Call(query string) (map[string]*PairInfo, error)
}

type Dexscreener struct{}

func (t *Dexscreener) Call(query string) (map[string]*PairInfo, error) {
	r, err := http.Get(memeUrl + query)
	if err != nil {
		log.Println("请求失败：", err)
		return nil, apiErr
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败：", err)
		return nil, apiErr
	}

	meme := new(Meme)
	err = json.Unmarshal(b, &meme)
	if err != nil {
		log.Println("json转换失败: ", err)
		return nil, apiErr
	}

	m := make(map[string]*PairInfo)

	for _, v := range meme.Pairs {
		if v.DexId == "uniswap" && v.QuoteToken != nil {
			v.CreateTime = time.Unix(v.CreateAt/1000, 0).Format("2006-01-02 15:04:05")
			for i := range v.Labels {
				if strings.EqualFold(v.QuoteToken.Symbol, "WETH") {
					pair := new(PairInfo)
					pair.CreateTime = v.CreateTime
					pair.PairAddress = v.PairAddress
					// if v.Lp != nil {
					// 	pair.Lp = v.Lp.Usd
					// }
					m[v.Labels[i]] = pair
				}
			}
		}
	}

	return m, nil
}

type HoneypotPair struct{}

func (t *HoneypotPair) Call(query string) (map[string]*PairInfo, error) {
	meme := make([]*HoneypotPairs, 0)
	req, err := http.NewRequest(http.MethodGet, honeypotPairsUrl+query, nil)
	if err != nil {
		log.Println("请求失败：", err)
		return nil, apiErr
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://honeypot.is/")
	req.Header.Set("Origin", "https://honeypot.is")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("请求失败：", err)
		return nil, apiErr
	}

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("body读取失败：", err)
		return nil, apiErr
	}

	err = json.Unmarshal(b, &meme)
	if err != nil {
		log.Println("json转换失败: ", err)
		return nil, apiErr
	}

	m := make(map[string]*PairInfo)

	for _, v := range meme {
		if v.Pairs != nil && strings.Contains(v.Pairs.Name, "Uniswap V2") && strings.Contains(v.Pairs.Name, "WETH") {
			pair := new(PairInfo)
			pair.CreateTime = time.Unix(v.CreatedAtTimestamp, 0).Format("2006-01-02 15:04:05")
			pair.PairAddress = v.Pairs.Address
			// pair.Lp = v.Liquidity
			m["v2"] = pair

		}
		if v.Pairs != nil && strings.Contains(v.Pairs.Name, "Uniswap V3") && strings.Contains(v.Pairs.Name, "WETH") {
			pair := new(PairInfo)
			pair.CreateTime = time.Unix(v.CreatedAtTimestamp, 0).Format("2006-01-02 15:04:05")
			pair.PairAddress = v.Pairs.Address
			// pair.Lp = v.Liquidity
			m["v3"] = pair
		}
	}

	return m, nil

}

