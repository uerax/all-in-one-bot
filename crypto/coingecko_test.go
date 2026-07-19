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
	// buildDailyKlines 只保留严格 UTC 00:00 的点; 每个自然日对应一根 K 线
	klines := buildDailyKlines(
		[][]float64{
			{1717200000000, 10},            // 2024-06-01 00:00
			{1717286400000, 20},            // 2024-06-02 00:00
			{1717286400000 + 3600_000, 25}, // 非 00:00, 应被忽略
		},
		[][]float64{
			{1717200000000, 100},
			{1717286400000, 200},
			{1717286400000 + 3600_000, 250},
		},
	)

	if len(klines) != 2 {
		t.Fatalf("len = %d, want 2", len(klines))
	}

	first := klines[0]
	if first.Time.Format("2006-01-02") != "2024-06-01" {
		t.Fatalf("first day = %s, want 2024-06-01", first.Time.Format("2006-01-02"))
	}
	if first.Open != 10 || first.High != 10 || first.Low != 10 || first.Close != 10 || first.Volume != 100 {
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

func TestCoingeckoNextAPIKeyRotates(t *testing.T) {
	c := &Coingecko{keys: []string{"k1", "k2", "k3"}}
	got := []string{c.nextAPIKey(), c.nextAPIKey(), c.nextAPIKey(), c.nextAPIKey()}
	want := []string{"k1", "k2", "k3", "k1"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("nextAPIKey[%d] = %q, want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestCoingeckoNextAPIKeyEmpty(t *testing.T) {
	c := &Coingecko{}
	if got := c.nextAPIKey(); got != "" {
		t.Fatalf("nextAPIKey = %q, want empty", got)
	}
}

func TestCoingeckoApplyAPIKeyHeader(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("x-cg-demo-api-key"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"coins":[]}`))
	}))
	defer srv.Close()

	c := &Coingecko{keys: []string{"alpha", "beta"}}
	if _, err := c.searchFromURL(srv.URL); err != nil {
		t.Fatalf("search 1: %v", err)
	}
	if _, err := c.searchFromURL(srv.URL); err != nil {
		t.Fatalf("search 2: %v", err)
	}
	if len(seen) != 2 || seen[0] != "alpha" || seen[1] != "beta" {
		t.Fatalf("headers = %v, want [alpha beta]", seen)
	}
}

func TestCoingeckoGetDailyKlineSendsKey(t *testing.T) {
	var key string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key = r.Header.Get("x-cg-demo-api-key")
		w.Header().Set("Content-Type", "application/json")
		// one complete UTC midnight sample so buildDailyKlines returns data
		_, _ = w.Write([]byte(`{"prices":[[1717200000000,10]],"total_volumes":[[1717200000000,100]]}`))
	}))
	defer srv.Close()

	c := &Coingecko{keys: []string{"kline-key"}}
	klines, err := c.getDailyKlineFromURL(srv.URL)
	if err != nil {
		t.Fatalf("getDailyKlineFromURL: %v", err)
	}
	if key != "kline-key" {
		t.Fatalf("header key = %q, want kline-key", key)
	}
	if len(klines) != 1 {
		t.Fatalf("len(klines) = %d, want 1", len(klines))
	}
}
