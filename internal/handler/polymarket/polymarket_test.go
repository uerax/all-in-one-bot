package polymarket

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/mocks"
)

func TestServiceListMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "2" {
			t.Fatalf("unexpected limit: %s", got)
		}
		if got := r.URL.Query().Get("active"); got != "true" {
			t.Fatalf("unexpected active: %s", got)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id":"1","question":"Will BTC hit 200k?","slug":"btc-200k","active":true,"closed":false},
			{"id":"2","question":"Will ETH hit 10k?","slug":"eth-10k","active":true,"closed":false}
		]`))
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL: server.URL,
		ClobBaseURL:  server.URL,
		Timeout:      1,
		DefaultLimit: 10,
	}, &mocks.MockLogger{})

	active := true
	markets, err := svc.ListMarkets(ListMarketsOptions{Limit: 2, Active: &active, Query: "btc"})
	if err != nil {
		t.Fatalf("ListMarkets() error = %v", err)
	}
	if len(markets) != 1 {
		t.Fatalf("ListMarkets() len = %d, want 1", len(markets))
	}
	if markets[0].ID != "1" {
		t.Fatalf("ListMarkets() first id = %s, want 1", markets[0].ID)
	}
}

func TestServiceGetMarketByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/42" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"42","question":"Test market","slug":"test-market"}`))
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL: server.URL,
		ClobBaseURL:  server.URL,
		Timeout:      1,
		DefaultLimit: 10,
	}, &mocks.MockLogger{})

	market, err := svc.GetMarketByID("42")
	if err != nil {
		t.Fatalf("GetMarketByID() error = %v", err)
	}
	if market.Slug != "test-market" {
		t.Fatalf("GetMarketByID() slug = %s, want test-market", market.Slug)
	}
}

func TestServiceGetMarketBySlug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/slug/test-market" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"42","question":"Test market","slug":"test-market"}`))
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL: server.URL,
		ClobBaseURL:  server.URL,
		Timeout:      1,
		DefaultLimit: 10,
	}, &mocks.MockLogger{})

	market, err := svc.GetMarketBySlug("test-market")
	if err != nil {
		t.Fatalf("GetMarketBySlug() error = %v", err)
	}
	if market.ID != "42" {
		t.Fatalf("GetMarketBySlug() id = %s, want 42", market.ID)
	}
}

func TestServiceGetTokenPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/price" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("token_id"); got != "token-1" {
			t.Fatalf("unexpected token id: %s", got)
		}
		if got := r.URL.Query().Get("side"); got != PriceSideBuy {
			t.Fatalf("unexpected side: %s", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"price":"0.42"}`))
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL: server.URL,
		ClobBaseURL:  server.URL,
		Timeout:      1,
		DefaultLimit: 10,
	}, &mocks.MockLogger{})

	price, err := svc.GetTokenPrice("token-1", PriceSideBuy)
	if err != nil {
		t.Fatalf("GetTokenPrice() error = %v", err)
	}
	if price.Price != "0.42" {
		t.Fatalf("GetTokenPrice() price = %s, want 0.42", price.Price)
	}
}

func TestServiceGetTokenPriceRejectsInvalidSide(t *testing.T) {
	svc := NewService(config.Polymarket{
		GammaBaseURL: "https://gamma-api.polymarket.com",
		ClobBaseURL:  "https://clob.polymarket.com",
		Timeout:      1,
		DefaultLimit: 10,
	}, &mocks.MockLogger{})

	_, err := svc.GetTokenPrice("token-1", "hold")
	if err != ErrInvalidPriceSide {
		t.Fatalf("GetTokenPrice() error = %v, want %v", err, ErrInvalidPriceSide)
	}
}

func TestServiceGetOrderBook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/book" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("token_id"); got != "token-1" {
			t.Fatalf("unexpected token id: %s", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"market":"market-1",
			"asset_id":"token-1",
			"bids":[{"price":"0.4","size":"10"}],
			"asks":[{"price":"0.5","size":"20"}],
			"last_trade_price":"0.45"
		}`))
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL: server.URL,
		ClobBaseURL:  server.URL,
		Timeout:      1,
		DefaultLimit: 10,
	}, &mocks.MockLogger{})

	book, err := svc.GetOrderBook("token-1")
	if err != nil {
		t.Fatalf("GetOrderBook() error = %v", err)
	}
	if len(book.Bids) != 1 || len(book.Asks) != 1 {
		t.Fatalf("GetOrderBook() bids=%d asks=%d, want 1/1", len(book.Bids), len(book.Asks))
	}
	if book.LastTradePrice != "0.45" {
		t.Fatalf("GetOrderBook() last trade = %s, want 0.45", book.LastTradePrice)
	}
}

func TestServiceHandlesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL: server.URL,
		ClobBaseURL:  server.URL,
		Timeout:      1,
		DefaultLimit: 10,
	}, &mocks.MockLogger{})

	_, err := svc.GetMarketByID("42")
	if err == nil {
		t.Fatal("GetMarketByID() error = nil, want error")
	}
	if err.Error() != "polymarket api error: bad request" {
		t.Fatalf("GetMarketByID() error = %v", err)
	}
}
