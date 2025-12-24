package bitcointalk

import "github.com/uerax/all-in-one-bot/lite/internal/store"

type MockStore struct{
	store.Store
	SetFunc func(db string, k string) (map[string]struct{}, error)
}

func (m *MockStore) Set(db string, k string) (map[string]struct{}, error) {
    if m.SetFunc != nil {
        return m.SetFunc(db, k)
    }
    return nil, nil
}