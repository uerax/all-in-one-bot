package crypto

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uerax/goconf"
)

var (
	pollingKey   *PollingKey
	pollingKeyV2 *PollingKeyV2
	once         sync.Once
)

type PollingKey struct {
	Keys []string
	idx  int32
}

func NewPollingKey() *PollingKey {
	once.Do(func() {
		pollingKey = &PollingKey{
			Keys: make([]string, 0),
			idx:  0,
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
	})
	return pollingKey
}

func (t *PollingKey) IsNull() bool {
	return len(t.Keys) == 0
}

func (t *PollingKey) Len() int {
	return len(t.Keys)
}

func (t *PollingKey) AddKeys(keys ...string) {
	t.Keys = append(t.Keys, keys...)
}

func (t *PollingKey) GetKey() string {
	atomic.AddInt32(&t.idx, 1)
	return t.Keys[int(t.idx)%(len(t.Keys)-1)]
}

func (t *PollingKey) CallsPerM() {
	tick := time.NewTicker(time.Minute)
	pre := t.idx
	for range tick.C {
		log.Printf("每分钟调用 %d 次 ApiKey", t.idx-pre)
		pre = t.idx
	}
}

type PollingKeyV2 struct {
	Keys   *LinkedList
	Now    *LinkedList
	mu     sync.Mutex
	Len    int
	Bucket *TokenBucket
	qps    int // 每秒调用最大次数
}

func NewPollingKeyV2() *PollingKeyV2 {
	once.Do(func() {
		pollingKeyV2 = &PollingKeyV2{
			Keys: new(LinkedList),
			Now:  nil,
			mu:   sync.Mutex{},
			qps:  0,
		}
		keys, err := goconf.VarArray("crypto", "etherscan", "keys")
		if err == nil {
			for k := range keys {
				if keys[k] != nil {
					if key, ok := keys[k].(string); ok {
						pollingKeyV2.Add(key)
					}
				}
			}
			pollingKeyV2.Bucket = NewTokenBucket(pollingKeyV2.Len * 5)
			go pollingKeyV2.Qps()
		}
	})
	return pollingKeyV2
}

func (t *PollingKeyV2) Qps() {
	tick := time.NewTicker(time.Second)
	for range tick.C {
		if t.qps > t.Len*4 {
			log.Println("GetKey QPS: ", t.qps)
		}
		t.qps = 0
	}
}

func (t *PollingKeyV2) GetKey() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Now == nil {
		t.Now = t.Keys.Next
	}
	defer func() {
		t.Now = t.Now.Next
	}()
	t.qps++
	t.Bucket.GetToken()
	return t.Now.Key

}

func (t *PollingKeyV2) Add(key string) {
	t.Keys.Add(key)
	t.Len++
}

func (t *PollingKeyV2) IsNull() bool {
	return t.Len == 0
}

type LinkedList struct {
	Key  string
	Next *LinkedList
}

func (t *LinkedList) Add(key string) {
	new := &LinkedList{
		Key: key,
	}
	if t.Next == nil {
		new.Next = new
		t.Next = new
	} else {
		new.Next = t.Next
		head := t.Next
		for head.Next != t.Next {
			head = head.Next
		}
		head.Next = new
	}
}
