package crocodile

import (
	"testing"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
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
