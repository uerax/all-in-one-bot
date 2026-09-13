package crocodile

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
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
	data      map[string]any
	loadCount int
	saveCount int
}

func (m *mockStore) Set(database string, key string) (map[string]struct{}, error) {
	return nil, nil
}

func (m *mockStore) Load(database string, key string, target any) error {
	m.loadCount++
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
	m.saveCount++
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
	deleted, _, err := c.deleteItem("doge")
	if err != nil {
		t.Fatalf("deleteItem error: %v", err)
	}
	if deleted {
		t.Fatalf("expected deleted=false for non-existent item")
	}

	// 3. Delete case-insensitively
	deleted, deletedID, err := c.deleteItem("BitCoin")
	if err != nil {
		t.Fatalf("deleteItem error: %v", err)
	}
	if !deleted || deletedID != "bitcoin" {
		t.Fatalf("expected deleted=true and deletedID='bitcoin', got deleted=%v, id=%s", deleted, deletedID)
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

	// 5. Add and delete by Name (case-insensitive)
	if err := c.addItem(Item{ID: "debtreliefbot", Name: "DRB"}); err != nil {
		t.Fatalf("addItem failed: %v", err)
	}
	deleted, deletedID, err = c.deleteItem("drb")
	if err != nil || !deleted || deletedID != "debtreliefbot" {
		t.Fatalf("expected deleted by name 'drb' -> 'debtreliefbot', got deleted=%v, id=%s, err=%v", deleted, deletedID, err)
	}
	list, _ = c.loadList()
	if len(list) != 0 {
		t.Fatalf("expected empty list after name delete, got: %+v", list)
	}
}

func TestCrocodile_StartupSyncAndMemoryOnly(t *testing.T) {
	initialItems := []Item{
		{ID: "wrapped-quil", Name: "wquil"},
		{ID: "debtreliefbot", Name: "DRB"},
	}

	ms := &mockStore{
		data: map[string]any{
			"crocodile:list": initialItems,
		},
	}
	msgCh := make(chan models.Message, 10)

	// 1. NewCrocodile runs startup init: loads once from store and persists
	c := NewCrocodile(ms, nil, msgCh, config.Crocodile{}, nil)

	if ms.loadCount != 1 {
		t.Fatalf("expected store Load to be called exactly 1 time on startup, got %d", ms.loadCount)
	}

	items := c.getItems()
	if len(items) != 2 {
		t.Fatalf("expected 2 items in memory, got %d", len(items))
	}

	// 2. ListMonitor must read from memory without calling store Load
	c.ListMonitor(100)
	select {
	case msg := <-msgCh:
		if !strings.Contains(msg.Text, "wrapped-quil") || !strings.Contains(msg.Text, "debtreliefbot") {
			t.Fatalf("unexpected ListMonitor message: %s", msg.Text)
		}
	default:
		t.Fatal("expected message in msgCh")
	}
	if ms.loadCount != 1 {
		t.Fatalf("expected store Load NOT to be called by ListMonitor, got %d", ms.loadCount)
	}

	// 3. DeleteMonitor deletes from memory and calls Save, but does NOT call Load
	c.DeleteMonitor(100, "wquil")
	select {
	case msg := <-msgCh:
		if !strings.Contains(msg.Text, "已删除监控: `wrapped-quil`") {
			t.Fatalf("unexpected DeleteMonitor message: %s", msg.Text)
		}
	default:
		t.Fatal("expected message in msgCh")
	}
	if ms.loadCount != 1 {
		t.Fatalf("expected store Load NOT to be called by DeleteMonitor, got %d", ms.loadCount)
	}

	// 4. ListMonitor reflects deletion immediately from memory
	c.ListMonitor(100)
	select {
	case msg := <-msgCh:
		if strings.Contains(msg.Text, "wrapped-quil") {
			t.Fatalf("deleted coin should not be present in list: %s", msg.Text)
		}
		if !strings.Contains(msg.Text, "debtreliefbot") {
			t.Fatalf("remaining coin should be present: %s", msg.Text)
		}
	default:
		t.Fatal("expected message in msgCh")
	}
	if ms.loadCount != 1 {
		t.Fatalf("expected store Load count to remain 1, got %d", ms.loadCount)
	}
}
