package coin

import (
	"errors"
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/mocks"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
)

type mockStore struct {
	data map[string]map[string]any
}

func (m *mockStore) Set(database string, key string) (map[string]struct{}, error) {
	return nil, nil
}

func (m *mockStore) Save(database string, key string, value any) error {
	return nil
}

func (m *mockStore) Load(database string, key string, target any) error {
	dbData, ok := m.data[database]
	if !ok {
		return errors.New("not found")
	}
	val, ok := dbData[key]
	if !ok {
		return errors.New("not found")
	}
	if ptr, ok := target.(*map[string]float64); ok {
		if mapVal, ok := val.(map[string]float64); ok {
			*ptr = mapVal
			return nil
		}
	}
	return errors.New("type mismatch")
}

func TestSyncList_PrimaryPath(t *testing.T) {
	st := &mockStore{
		data: map[string]map[string]any{
			"coin": {
				"list": map[string]float64{"BTC": 1.5, "ETH": 10.0},
			},
		},
	}
	ch := make(chan models.Message, 10)
	svc := NewService(st, nil, ch, &mocks.MockLogger{})

	svc.syncList()

	svc.mu.Lock()
	defer svc.mu.Unlock()
	if len(svc.list) != 2 || svc.list["BTC"] != 1.5 || svc.list["ETH"] != 10.0 {
		t.Fatalf("expected primary path coin/list to load, got: %v", svc.list)
	}
}

func TestSyncList_LegacyFallback(t *testing.T) {
	st := &mockStore{
		data: map[string]map[string]any{
			"coingecko": {
				"list": map[string]float64{"SOL": 50.0},
			},
		},
	}
	ch := make(chan models.Message, 10)
	svc := NewService(st, nil, ch, &mocks.MockLogger{})

	svc.syncList()

	svc.mu.Lock()
	defer svc.mu.Unlock()
	if len(svc.list) != 1 || svc.list["SOL"] != 50.0 {
		t.Fatalf("expected legacy fallback coingecko/list to load, got: %v", svc.list)
	}
}
