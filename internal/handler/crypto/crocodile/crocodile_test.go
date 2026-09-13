package crocodile

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
)

func TestEvaluate_ExcludesUnclosedToday(t *testing.T) {
	nowUTC := time.Now().UTC()
	todayStr := nowUTC.Format("2006-01-02")
	todayTime, _ := time.Parse("2006-01-02", todayStr)

	// Create 7 daily klines: 5 historical days, yesterday (target), today (unclosed 5-min candle)
	klines := []provider.DailyKline{
		{Timestamp: todayTime.AddDate(0, 0, -6), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -5), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -4), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -3), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -2), Volume: 100, Close: 1.0}, // Prev Day (Volume = 100)
		{Timestamp: todayTime.AddDate(0, 0, -1), Volume: 400, Close: 1.5}, // Yesterday Target (Volume = 400, 4x prev, 4x avg)
		{Timestamp: todayTime, Volume: 5, Close: 1.5},                     // Today Unclosed (Volume = 5 only!)
	}

	rc := RuleConfig{
		Lookback:          5,
		YesterdayMultiple: 3.0,
		AverageMultiple:   2.0,
	}

	item := Item{ID: "solana", Name: "Solana"}

	sig, err := evaluate(item, klines, rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sig == nil {
		t.Fatal("expected volume spike signal, got nil")
	}

	// Verify that the evaluated target candle is indeed Yesterday (Volume 400), NOT Today (Volume 5)
	if sig.Volume != 400 {
		t.Errorf("sig.Volume = %f, want 400 (Yesterday's volume)", sig.Volume)
	}

	if sig.YesterdayRatio != 4.0 {
		t.Errorf("sig.YesterdayRatio = %f, want 4.0", sig.YesterdayRatio)
	}
}

type mockStore struct {
	data map[string]any
}

func (m *mockStore) Set(database string, key string) (map[string]struct{}, error) {
	return nil, nil
}

func (m *mockStore) Load(database string, key string, target any) error {
	val, ok := m.data[database+":"+key]
	if !ok {
		return nil
	}
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func (m *mockStore) Save(database string, key string, value any) error {
	m.data[database+":"+key] = value
	return nil
}

func TestCrocodile_AddAndDeleteItem(t *testing.T) {
	ms := &mockStore{data: make(map[string]any)}
	msgCh := make(chan models.Message, 10)
	c := &Crocodile{
		db:          ms,
		klineCache:  make(map[string]klineCacheItem),
		lastTrigger: make(map[string]string),
		ch:          msgCh,
	}

	// 1. Add items
	if err := c.addItem(Item{ID: "bitcoin", Name: "BTC"}); err != nil {
		t.Fatalf("addItem failed: %v", err)
	}
	if err := c.addItem(Item{ID: "ethereum", Name: "ETH"}); err != nil {
		t.Fatalf("addItem failed: %v", err)
	}

	list, err := c.loadList()
	if err != nil || len(list) != 2 {
		t.Fatalf("expected 2 items, got %d (err: %v)", len(list), err)
	}

	// 2. Delete non-existent item
	deleted, err := c.deleteItem("doge")
	if err != nil {
		t.Fatalf("deleteItem error: %v", err)
	}
	if deleted {
		t.Fatalf("expected deleted=false for non-existent item")
	}

	// 3. Delete case-insensitively
	deleted, err = c.deleteItem("BitCoin")
	if err != nil {
		t.Fatalf("deleteItem error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted=true for bitcoin")
	}

	list, _ = c.loadList()
	if len(list) != 1 || list[0].ID != "ethereum" {
		t.Fatalf("expected 1 item 'ethereum', got: %+v", list)
	}

	// 4. Delete via DeleteMonitor with cache & trigger cleanup
	c.klineCache["ethereum"] = klineCacheItem{}
	c.lastTrigger["ethereum:volume_spike"] = "2026-09-14"

	c.DeleteMonitor(100, "ethereum")

	select {
	case msg := <-msgCh:
		if !strings.Contains(msg.Text, "已删除监控: `ethereum`") {
			t.Errorf("unexpected message: %s", msg.Text)
		}
	default:
		t.Fatal("expected message in msgCh")
	}

	if _, ok := c.klineCache["ethereum"]; ok {
		t.Error("expected klineCache to be cleared")
	}
	if _, ok := c.lastTrigger["ethereum:volume_spike"]; ok {
		t.Error("expected lastTrigger to be cleared")
	}

	list, _ = c.loadList()
	if len(list) != 0 {
		t.Fatalf("expected empty list, got: %+v", list)
	}
}
