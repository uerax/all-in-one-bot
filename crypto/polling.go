package crypto

import "sync"

type PollingKey struct {
	Keys []string
	lock sync.RWMutex
	idx int
}

func NewPollingKey() *PollingKey {
	return &PollingKey{
		Keys: make([]string, 0),
		lock: sync.RWMutex{},
		idx: 0,
	}
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