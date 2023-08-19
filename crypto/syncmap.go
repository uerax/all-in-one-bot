package crypto

import "sync"

type SyncMap struct {
	Mu sync.Mutex
	M  map[string]any
}

func NewSyncMap() *SyncMap {
	return &SyncMap{
		Mu: sync.Mutex{},
		M:  make(map[string]any),
	}
}

func (t *SyncMap) Swap(addr string, v any) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	t.M[addr] = v
}

func (t *SyncMap) ExistOrStore(addr string, v any) bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if _, ok := t.M[addr]; !ok {
		t.M[addr] = v
		return false
	}
	return true
}

func (t *SyncMap) Store(addr string, v any) {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	t.M[addr] = v
}
