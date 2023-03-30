package crypto

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"tg-aio-bot/common"
	"time"

	"github.com/uerax/goconf"
)

type user struct {
	Id int64
	HighLine map[string]string //crypto -> priceline
	lowLine map[string]string
}

type CryptoMonitor struct {
	UTC sync.Map	// id -> user
	CTU sync.Map	// crypto -> user
	ctx	context.Context
	cancel context.CancelFunc
	api *Crypto
	notifyHigher map[int64]map[string]string	// id -> crypto -> price
	notifyHigherLog map[int64]map[string]int64
	notifyLower map[int64]map[string]string
	notifyLowerLog map[int64]map[string]int64
	C chan map[int64]map[string]string // id -> crypto -> price
}

func NewCryptoMonitor() *CryptoMonitor {
	parent, done := context.WithCancel(context.Background())
	return &CryptoMonitor{
		UTC: sync.Map{},
		CTU: sync.Map{},
		ctx: parent,
		cancel: done,
		notifyHigher: make(map[int64]map[string]string),
		notifyHigherLog: make(map[int64]map[string]int64),
		notifyLower: make(map[int64]map[string]string),
		notifyLowerLog: make(map[int64]map[string]int64),
		api: NewCrypto(goconf.VarStringOrDefault("", "crypto", "binance", "apiKey"),goconf.VarStringOrDefault("", "crypto", "binance", "secretKey")),
		C: make(chan map[int64]map[string]string, 1),
	}
}

func (t *CryptoMonitor) Context() {
	t.cancel()
	t.notifyHigher = make(map[int64]map[string]string)
	t.notifyLower = make(map[int64]map[string]string)
	parent, done := context.WithCancel(context.Background())
	t.ctx = parent
	t.cancel = done
	go t.Start()
}

func (t *CryptoMonitor) AddHighMonitor(id int64, crypto, price string) {
	usr, _ := t.UTC.LoadOrStore(id, &user{
		Id: id,
		HighLine: make(map[string]string),
		lowLine: make(map[string]string),
	})
	usr.(*user).HighLine[crypto+"USDT"] = price
	t.CTU.LoadOrStore(crypto+"USDT", usr)

	t.Context()
	
}

func (t *CryptoMonitor) AddLowMonitor(id int64, crypto, price string) {
	usr, _ := t.UTC.LoadOrStore(id, &user{
		Id: id,
		HighLine: make(map[string]string),
		lowLine: make(map[string]string),
	})
	usr.(*user).lowLine[crypto+"USDT"] = price
	t.CTU.LoadOrStore(crypto+"USDT", usr)

	t.Context()
	
}

func (t *CryptoMonitor) Start() {
	ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
	fmt.Println("[开始执行定时探测]")
    for {
        select {
        case <-ticker.C:
			keys := make([]string, 0, 10)
			t.CTU.Range(func(key, value any) bool {
				keys = append(keys, key.(string))
				return true
			})
			t.probe(keys)

        case <-t.ctx.Done():
            // 当收到ctx的完成信号时，停止探测
			fmt.Println("[定时探测结束]")
            return
        }
    }
}

func (t *CryptoMonitor) probe(cryptos []string) {
	now := time.Now().Unix()
	m := t.api.Price(cryptos...)
	for k, v := range m {
		// k crypto v price
		val, ok := t.CTU.Load(k)
		if !ok {
			continue
		}
		curPrice, err := strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}
		for _, v1 := range val.(*user).HighLine {
			// k1 crypto v1 priceline

			threshold, err := strconv.ParseFloat(v1, 64)
			if err != nil {
				continue
			}

			if curPrice >= threshold {
				if _, ok := t.notifyHigher[val.(*user).Id]; !ok {
					t.notifyHigher[val.(*user).Id] = make(map[string]string)
				}
				if _, ok := t.notifyHigherLog[val.(*user).Id]; !ok {
					t.notifyHigherLog[val.(*user).Id] = make(map[string]int64)
				}
				if last, ok := t.notifyHigherLog[val.(*user).Id][k]; ok {
					if now - last >= 3600 {
						t.notifyHigher[val.(*user).Id][k] = v
						t.notifyHigherLog[val.(*user).Id][k] = now
					}
				} else {
					t.notifyHigher[val.(*user).Id][k] = v
					t.notifyHigherLog[val.(*user).Id][k] = now
				}
				
			}
		}
		for _, v1 := range val.(*user).lowLine {
			// k1 crypto v1 priceline

			threshold, err := strconv.ParseFloat(v1, 64)
			if err != nil {
				continue
			}

			if curPrice <= threshold {
				if _, ok := t.notifyLower[val.(*user).Id]; !ok {
					t.notifyLower[val.(*user).Id] = make(map[string]string)
				}
				if _, ok := t.notifyLowerLog[val.(*user).Id]; !ok {
					t.notifyLowerLog[val.(*user).Id] = make(map[string]int64)
				}
				if last, ok := t.notifyLowerLog[val.(*user).Id][k]; ok {
					if now - last >= 3600 {
						t.notifyLower[val.(*user).Id][k] = v
						t.notifyLowerLog[val.(*user).Id][k] = now
					}
				} else {
					t.notifyLower[val.(*user).Id][k] = v
					t.notifyLowerLog[val.(*user).Id][k] = now
				}
				
			}
		}
		
	}
	if len(t.notifyHigher) != 0 && t.notifyHigher != nil {
		t.C <- t.notifyHigher
		t.notifyHigher = make(map[int64]map[string]string)
	
	}
	
	if len(t.notifyLower) != 0 && t.notifyLower != nil {
		t.C <- t.notifyLower
		t.notifyLower = make(map[int64]map[string]string)
	}

}

func (t *CryptoMonitor) GetPrice(id int64, crypto ...string) map[string]string {
	for k := range crypto {
		crypto[k] = crypto[k] + "USDT"
	}
	if len(crypto) == 0 || crypto == nil {
		if value, ok := t.UTC.Load(id); ok {
			for k := range value.(*user).lowLine {
				crypto = append(crypto, k)
			}
			for k := range value.(*user).HighLine {
				if !common.InSlice(crypto, k) {
					crypto = append(crypto, k)
				}
			}
		}
	}
	return t.api.Price(crypto...)
}
