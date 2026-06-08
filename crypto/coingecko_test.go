package crypto

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCoingeckoDailyKlineURLDoesNotUsePaidInterval(t *testing.T) {
	rawURL := coingeckoDailyKlineURL("bitcoin")
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	q := parsed.Query()
	if got := q.Get("vs_currency"); got != "usd" {
		t.Fatalf("vs_currency = %q, want usd", got)
	}
	if got := q.Get("days"); got == "" {
		t.Fatal("days should be set internally")
	}
	if got := q.Get("interval"); got != "" {
		t.Fatalf("interval = %q, want empty", got)
	}
}

func TestBuildDailyKlinesAggregatesPrices(t *testing.T) {
	klines := buildDailyKlines(
		[][]float64{
			{1717200000000, 10},
			{1717203600000, 15},
			{1717207200000, 8},
			{1717210800000, 12},
			{1717286400000, 20},
		},
		[][]float64{
			{1717200000000, 100},
			{1717210800000, 120},
			{1717286400000, 200},
		},
	)

	if len(klines) != 2 {
		t.Fatalf("len = %d, want 2", len(klines))
	}

	first := klines[0]
	if first.Time.Format("2006-01-02") != "2024-06-01" {
		t.Fatalf("first day = %s, want 2024-06-01", first.Time.Format("2006-01-02"))
	}
	if first.Open != 10 || first.High != 15 || first.Low != 8 || first.Close != 12 || first.Volume != 120 {
		t.Fatalf("unexpected first kline: %+v", first)
	}

	second := klines[1]
	if second.Open != 20 || second.High != 20 || second.Low != 20 || second.Close != 20 || second.Volume != 200 {
		t.Fatalf("unexpected second kline: %+v", second)
	}
}

func TestCoingeckoSearchURL(t *testing.T) {
	rawURL := coingeckoSearchURL("wrapped quil")
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	if got := parsed.Query().Get("query"); got != "wrapped quil" {
		t.Fatalf("query = %q, want wrapped quil", got)
	}
}

func TestCoingeckoSearchParsesCoins(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"coins":[{"id":"wrapped-quil","name":"Wrapped QUIL","symbol":"WQUIL","market_cap_rank":1234}]}`))
	}))
	defer srv.Close()

	got, err := NewCoingecko().searchFromURL(srv.URL)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != "wrapped-quil" || got[0].Name != "Wrapped QUIL" || got[0].Symbol != "WQUIL" {
		t.Fatalf("unexpected coin: %+v", got[0])
	}
	if got[0].MarketCapRank == nil || *got[0].MarketCapRank != 1234 {
		t.Fatalf("unexpected rank: %+v", got[0].MarketCapRank)
	}
}

func TestCoingeckoSearchHandlesEmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"coins":[]}`))
	}))
	defer srv.Close()

	got, err := NewCoingecko().searchFromURL(srv.URL)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestCoingeckoSearchReturnsStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	_, err := NewCoingecko().searchFromURL(srv.URL)
	if err == nil {
		t.Fatal("expected error")
	}
}
