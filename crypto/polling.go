package crypto

import (
	"sync"

	"github.com/uerax/goconf"
)

var pollingKey *PollingKey

type PollingKey struct {
	Keys []string
	lock sync.RWMutex
	idx int
}

func NewPollingKey() *PollingKey {
	if pollingKey == nil {
		pollingKey = &PollingKey{
			Keys: make([]string, 0),
			lock: sync.RWMutex{},
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
		
	}
	
	return pollingKey
	
}

func (t *PollingKey) IsNull() bool {
	return len(t.Keys) == 0
}

func (t *PollingKey) AddKeys(keys ...string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Keys = append(t.Keys, keys...)
}

func (t *PollingKey) GetKey() string {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.idx == len(t.Keys) - 1 {
		t.idx = 0
	} else {
		t.idx++
	}
	return t.Keys[t.idx]
}