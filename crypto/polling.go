package crypto

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/uerax/goconf"
)

var pollingKey *PollingKey

type PollingKey struct {
	Keys []string
	idx  int32
}

func NewPollingKey() *PollingKey {
	if pollingKey == nil {
		pollingKey = &PollingKey{
			Keys: make([]string, 0),
			idx: 0,
		}
		keys, err := goconf.VarArray("crypto", "etherscan", "keys")
		if err == nil {
			for k := range keys {
				if keys[k] != nil {
					pollingKey.AddKeys(keys[k].(string))
				}	
			}
		}
		go pollingKey.CallsPerM()
	}
	
	return pollingKey
}

func (t *PollingKey) IsNull() bool {
	return len(t.Keys) == 0
}

func (t *PollingKey) AddKeys(keys ...string) {
	t.Keys = append(t.Keys, keys...)
}

func (t *PollingKey) GetKey() string  {
	atomic.AddInt32(&t.idx, 1)
	return t.Keys[int(t.idx) % (len(t.Keys) - 1)]
}

func (t *PollingKey) CallsPerM() {
	tick := time.NewTicker(time.Minute)
	pre := t.idx
	for range tick.C {
		log.Printf("每分钟调用 %d 次 ApiKey", t.idx - pre)
		pre = t.idx
	}
}