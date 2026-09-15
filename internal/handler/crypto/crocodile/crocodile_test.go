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

	// Create 9 daily klines: 7 historical days, yesterday (target), today (unclosed 5-min candle)
	klines := []provider.DailyKline{
		{Timestamp: todayTime.AddDate(0, 0, -8), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -7), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -6), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -5), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -4), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -3), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -2), Volume: 100, Close: 1.0}, // Prev Day (Volume = 100)
		{Timestamp: todayTime.AddDate(0, 0, -1), Volume: 400, Close: 1.5}, // Yesterday Target (Volume = 400, 4x prev, 4x avg)
		{Timestamp: todayTime, Volume: 5, Close: 1.5},                     // Today Unclosed (Volume = 5 only!)
	}

	rc := RuleConfig{
		Lookback:          7,
		YesterdayMultiple: 2.0,
		AverageMultiple:   3.0,
	}

	item := Item{Network: "solana", Name: "SOL"}

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

	if sig.AverageRatio != 4.0 {
		t.Errorf("sig.AverageRatio = %f, want 4.0", sig.AverageRatio)
	}

	if sig.PriceChangePct < 49.9 || sig.PriceChangePct > 50.1 {
		t.Errorf("sig.PriceChangePct = %f, want 50.0%%", sig.PriceChangePct)
	}
}

func TestEvaluate_DescendingOrUnsortedKlines(t *testing.T) {
	nowUTC := time.Now().UTC()
	todayStr := nowUTC.Format("2006-01-02")
	todayTime, _ := time.Parse("2006-01-02", todayStr)

	// GeckoTerminal 等 API 默认返回倒序 K 线（最新在前，最旧在后）
	descendingKlines := []provider.DailyKline{
		{Timestamp: todayTime, Volume: 10, Close: 1.5},                    // Today (unclosed)
		{Timestamp: todayTime.AddDate(0, 0, -1), Volume: 500, Close: 1.2}, // Yesterday (target, 5x spike)
		{Timestamp: todayTime.AddDate(0, 0, -2), Volume: 100, Close: 1.0}, // Prev Day
		{Timestamp: todayTime.AddDate(0, 0, -3), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -4), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -5), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -6), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -7), Volume: 100, Close: 1.0},
		{Timestamp: todayTime.AddDate(0, 0, -8), Volume: 100, Close: 1.0},
	}

	rc := RuleConfig{
		Lookback:          7,
		YesterdayMultiple: 2.0,
		AverageMultiple:   3.0,
	}

	item := Item{Network: "base", Name: "b3"}

	sig, err := evaluate(item, descendingKlines, rc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sig == nil || !sig.Triggered {
		t.Fatal("expected volume spike signal to be triggered from descending klines, got nil or untriggered")
	}

	if sig.Volume != 500 {
		t.Errorf("sig.Volume = %f, want 500 (Yesterday's volume)", sig.Volume)
	}

	if sig.YesterdayRatio != 5.0 {
		t.Errorf("sig.YesterdayRatio = %f, want 5.0", sig.YesterdayRatio)
	}

	if sig.AverageRatio != 5.0 {
		t.Errorf("sig.AverageRatio = %f, want 5.0", sig.AverageRatio)
	}

	if sig.PriceChangePct < 19.9 || sig.PriceChangePct > 20.1 {
		t.Errorf("sig.PriceChangePct = %f, want 20.0%%", sig.PriceChangePct)
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		price    float64
		expected string
	}{
		{123.456, "$123.46"},
		{1.0, "$1.00"},
		{0.0512, "$0.0512"},
		{0.005126, "$0.005126"},
		{0.000185, "$0.000185"},
		{0.00000399, "$0.00000399"},
		{0.0, "$0.00"},
	}

	for _, tt := range tests {
		got := formatPrice(tt.price)
		if got != tt.expected {
			t.Errorf("formatPrice(%f) = %q, want %q", tt.price, got, tt.expected)
		}
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

	// 1. Add items with network and name
	if err := c.addItem(Item{Network: "base", Name: "auki"}); err != nil {
		t.Fatalf("addItem failed: %v", err)
	}
	if err := c.addItem(Item{Network: "eth", Name: "wquil"}); err != nil {
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

	// 3. Delete case-insensitively by name
	deleted, deletedLabel, err := c.deleteItem("Auki")
	if err != nil {
		t.Fatalf("deleteItem error: %v", err)
	}
	if !deleted || deletedLabel != "[BASE] auki" {
		t.Fatalf("expected deleted=true and label='[BASE] auki', got deleted=%v, label=%s", deleted, deletedLabel)
	}

	list, _ = c.loadList()
	if len(list) != 1 || list[0].Name != "wquil" {
		t.Fatalf("expected 1 item 'wquil', got: %+v", list)
	}

	// 4. Delete via DeleteMonitor with cache & trigger cleanup
	c.klineCache["eth:wquil"] = klineCacheItem{}
	c.lastTrigger["eth:wquil:volume_spike"] = "2026-09-14"

	c.DeleteMonitor(100, "wquil")

	select {
	case msg := <-msgCh:
		if !strings.Contains(msg.Text, "已删除 DEX 监控: `[ETH] wquil`") {
			t.Errorf("unexpected message: %s", msg.Text)
		}
	default:
		t.Fatal("expected message in msgCh")
	}

	if _, ok := c.klineCache["eth:wquil"]; ok {
		t.Error("expected klineCache to be cleared")
	}
	if _, ok := c.lastTrigger["eth:wquil:volume_spike"]; ok {
		t.Error("expected lastTrigger to be cleared")
	}

	list, _ = c.loadList()
	if len(list) != 0 {
		t.Fatalf("expected empty list, got: %+v", list)
	}
}

func TestCrocodile_StartupSyncAndMemoryOnly(t *testing.T) {
	initialItems := []Item{
		{Network: "eth", Name: "wquil"},
		{Network: "base", Name: "DRB"},
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
		if !strings.Contains(msg.Text, "wquil") || !strings.Contains(msg.Text, "DRB") {
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
		if !strings.Contains(msg.Text, "已删除 DEX 监控: `[ETH] wquil`") {
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
		if strings.Contains(msg.Text, "wquil") {
			t.Fatalf("deleted coin should not be present in list: %s", msg.Text)
		}
		if !strings.Contains(msg.Text, "DRB") {
			t.Fatalf("remaining coin should be present: %s", msg.Text)
		}
	default:
		t.Fatal("expected message in msgCh")
	}
	if ms.loadCount != 1 {
		t.Fatalf("expected store Load count to remain 1, got %d", ms.loadCount)
	}
}
