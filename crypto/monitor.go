package crypto

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"tg-aio-bot/common"
	"time"

	"github.com/uerax/goconf"
)

type user struct {
	Id       int64
	HighLine map[string]string //crypto -> priceline
	LowLine  map[string]string //crypto -> priceline
}

type CryptoMonitor struct {
	UTC             sync.Map // id -> user  map[id]*user
	CTU             sync.Map // crypto -> user map[id]
	ctx             context.Context
	cancel          context.CancelFunc
	api             *Crypto
	notifyHigher    map[int64]map[string]string // id -> crypto -> price
	notifyHigherLog map[int64]map[string]int64
	notifyLower     map[int64]map[string]string
	notifyLowerLog  map[int64]map[string]int64
	C               chan map[int64]map[string]string // id -> crypto -> price
	unit            string
}

func NewCryptoMonitor() *CryptoMonitor {
	parent, done := context.WithCancel(context.Background())
	return &CryptoMonitor{
		UTC:             sync.Map{},
		CTU:             sync.Map{},
		ctx:             parent,
		cancel:          done,
		notifyHigher:    make(map[int64]map[string]string),
		notifyHigherLog: make(map[int64]map[string]int64),
		notifyLower:     make(map[int64]map[string]string),
		notifyLowerLog:  make(map[int64]map[string]int64),
		api:             NewCrypto(goconf.VarStringOrDefault("", "crypto", "binance", "apiKey"), goconf.VarStringOrDefault("", "crypto", "binance", "secretKey")),
		C:               make(chan map[int64]map[string]string, 1),
		unit:            goconf.VarStringOrDefault("USDT", "crypto", "unit"),
	}
}

func (t *CryptoMonitor) Context() {
	t.cancel()
	parent, done := context.WithCancel(context.Background())
	t.ctx = parent
	t.cancel = done
	go t.Start()
}

func (t *CryptoMonitor) AddHighMonitor(id int64, crypto, price string) {
	crypto = strings.ToUpper(crypto)
	usr, _ := t.UTC.LoadOrStore(id, &user{
		Id:       id,
		HighLine: make(map[string]string),
		LowLine:  make(map[string]string),
	})
	usr.(*user).HighLine[crypto+t.unit] = price
	ctp, _ := t.CTU.LoadOrStore(crypto+t.unit, make(map[int64]*user))
	ctp.(map[int64]*user)[id] = usr.(*user)
	t.Context()

}

func (t *CryptoMonitor) AddLowMonitor(id int64, crypto, price string) {
	crypto = strings.ToUpper(crypto)
	usr, _ := t.UTC.LoadOrStore(id, &user{
		Id:       id,
		HighLine: make(map[string]string),
		LowLine:  make(map[string]string),
	})
	usr.(*user).LowLine[crypto+t.unit] = price
	ctp, _ := t.CTU.LoadOrStore(crypto+t.unit, make(map[int64]*user))
	ctp.(map[int64]*user)[id] = usr.(*user)

	t.Context()

}

func (t *CryptoMonitor) Start() {
	interval := goconf.VarIntOrDefault(10, "crypto", "monitor", "interval")
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
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
			t.clearNotify()
		case <-t.ctx.Done():
			// 当收到ctx的完成信号时，停止探测
			fmt.Println("[定时探测结束]")
			return
		}
	}
}

func (t *CryptoMonitor) clearNotify() {
	t.notifyHigher = make(map[int64]map[string]string)
	t.notifyLower = make(map[int64]map[string]string)
}

func (t *CryptoMonitor) probe(cryptos []string) {
	frequency := goconf.VarInt64OrDefault(600, "crypto", "monitor", "frequency")
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
		for k1, v1 := range val.(map[int64]*user) {
			// k1 crypto v1 priceline

			threshold, err := strconv.ParseFloat(v1.HighLine[k], 64)
			if err != nil {
				continue
			}

			if curPrice >= threshold {
				if _, ok := t.notifyHigher[k1]; !ok {
					t.notifyHigher[k1] = make(map[string]string)
				}
				if _, ok := t.notifyHigherLog[k1]; !ok {
					t.notifyHigherLog[k1] = make(map[string]int64)
				}
				if last, ok := t.notifyHigherLog[k1][k]; ok {
					if now-last >= frequency {
						t.notifyHigher[k1][k] = v
						t.notifyHigherLog[k1][k] = now
					}
				} else {
					t.notifyHigher[k1][k] = v
					t.notifyHigherLog[k1][k] = now
				}

			}
		}
		for k1, v1 := range val.(map[int64]*user) {
			// k1 crypto v1 priceline

			threshold, err := strconv.ParseFloat(v1.LowLine[k], 64)
			if err != nil {
				continue
			}

			if curPrice <= threshold {
				if _, ok := t.notifyLower[k1]; !ok {
					t.notifyLower[k1] = make(map[string]string)
				}
				if _, ok := t.notifyLowerLog[k1]; !ok {
					t.notifyLowerLog[k1] = make(map[string]int64)
				}
				if last, ok := t.notifyLowerLog[k1][k]; ok {
					if now-last >= frequency {
						t.notifyLower[k1][k] = v
						t.notifyLowerLog[k1][k] = now
					}
				} else {
					t.notifyLower[k1][k] = v
					t.notifyLowerLog[k1][k] = now
				}

			}
		}

	}
	if len(t.notifyHigher) != 0 && t.notifyHigher != nil {
		t.C <- t.notifyHigher
	}

	if len(t.notifyLower) != 0 && t.notifyLower != nil {
		t.C <- t.notifyLower
	}

}

func (t *CryptoMonitor) GetPrice(id int64, crypto ...string) map[string]string {
	for k := range crypto {
		crypto[k] = strings.ToUpper(crypto[k]) + t.unit
	}
	if crypto == nil {
		if value, ok := t.UTC.Load(id); ok {
			for k := range value.(*user).LowLine {
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

func (t *CryptoMonitor) DeleteMonitor(id int64, crypto ...string) {
	for k := range crypto {
		crypto[k] = strings.ToUpper(crypto[k]) + t.unit
	}
	if crypto != nil {
		value, ok := t.UTC.Load(id)
		if ok {
			for k := range crypto {
				delete(value.(*user).LowLine, crypto[k])
				delete(value.(*user).HighLine, crypto[k])
				//delete(t.notifyHigherLog[id], crypto[k]) 如果删除后需要刷新通知状态,会出现并发操作的情况需要改造成sync.Map.
				//delete(t.notifyLowerLog[id], crypto[k])
			}
		}
	}

}

func (t *CryptoMonitor) GetUFuturePrice(id int64, crypto string) string {
	return t.api.FuturesPrice(crypto)
}
