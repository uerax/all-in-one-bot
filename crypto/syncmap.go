package crypto

import "sync"

type SyncMap struct {
	Mu sync.Mutex
	M  map[string]struct{}
}

func (t *SyncMap) Exist(addr string) bool {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if _, ok := t.M[addr]; !ok {
		t.M[addr] = struct{}{}
		return false
	}
	return true
}