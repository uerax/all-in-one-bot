package polymarket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/mocks"
)

func newTestService(serverURL string, defaultLimit int) *Service {
	return NewService(config.Polymarket{
		GammaBaseURL: serverURL,
		ClobBaseURL:  serverURL,
		Timeout:      1,
		DefaultLimit: defaultLimit,
	}, &mocks.MockLogger{})
}

func TestServiceListMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("next_cursor"); got != InitialCursor {
			t.Fatalf("unexpected next_cursor: %s", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"limit":2,
			"count":2,
			"next_cursor":"LTE=",
			"data":[
				{"id":"1","question":"Will BTC hit 200k?","slug":"btc-200k","active":true,"closed":false},
				{"id":"2","question":"Will ETH hit 10k?","slug":"eth-10k","active":true,"closed":false}
			]
		}`))
	}))
	defer server.Close()

	svc := newTestService(server.URL, 10)
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

	svc := newTestService(server.URL, 10)
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
		if r.URL.Path != "/markets" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("next_cursor"); got != InitialCursor {
			t.Fatalf("unexpected next_cursor: %s", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"limit":2,
			"count":2,
			"next_cursor":"LTE=",
			"data":[
				{"id":"41","question":"Other market","slug":"other-market","active":true,"closed":false},
				{"id":"42","question":"Test market","slug":"test-market","active":true,"closed":false}
			]
		}`))
	}))
	defer server.Close()

	svc := newTestService(server.URL, 10)
	market, err := svc.GetMarketBySlug("test-market")
	if err != nil {
		t.Fatalf("GetMarketBySlug() error = %v", err)
	}
	if market.ID != "42" {
		t.Fatalf("GetMarketBySlug() id = %s, want 42", market.ID)
	}
}

func TestServiceListEventsAggregatesByEventSlug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"limit":3,
			"count":3,
			"next_cursor":"LTE=",
			"data":[
				{"id":"1","question":"Q1","slug":"m1","active":true,"closed":false,"event_id":"e-1","event_slug":"event-a","event_title":"Event A","event_category":"Politics"},
				{"id":"2","question":"Q2","slug":"m2","active":false,"closed":true,"event_id":"e-1","event_slug":"event-a","event_title":"Event A","event_category":"Politics"},
				{"id":"3","question":"Q3","slug":"m3","active":true,"closed":false,"event_id":"e-2","event_slug":"event-b","event_title":"Event B","event_category":"Sports"}
			]
		}`))
	}))
	defer server.Close()

	svc := newTestService(server.URL, 10)
	events, err := svc.ListEvents(10)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("ListEvents() len = %d, want 2", len(events))
	}
	if events[0].Slug != "event-a" {
		t.Fatalf("first event slug = %s, want event-a", events[0].Slug)
	}
	if len(events[0].Markets) != 2 {
		t.Fatalf("event-a markets len = %d, want 2", len(events[0].Markets))
	}
	if !events[0].Active {
		t.Fatal("event-a active = false, want true")
	}
	if events[0].Closed {
		t.Fatal("event-a closed = true, want false")
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

	svc := newTestService(server.URL, 10)
	price, err := svc.GetTokenPrice("token-1", PriceSideBuy)
	if err != nil {
		t.Fatalf("GetTokenPrice() error = %v", err)
	}
	if price.Price != "0.42" {
		t.Fatalf("GetTokenPrice() price = %s, want 0.42", price.Price)
	}
}

func TestServiceGetTokenPriceRejectsInvalidSide(t *testing.T) {
	svc := newTestService("https://clob.polymarket.com", 10)
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

	svc := newTestService(server.URL, 10)
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
		if r.URL.Path != "/markets/42" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	svc := newTestService(server.URL, 10)
	_, err := svc.GetMarketByID("42")
	if err == nil {
		t.Fatal("GetMarketByID() error = nil, want error")
	}
	if err.Error() != "polymarket api error: bad request" {
		t.Fatalf("GetMarketByID() error = %v", err)
	}
}

func TestServiceGetMarketBySlugNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"limit":1,"count":1,"next_cursor":"LTE=","data":[{"id":"1","slug":"other"}]}`))
	}))
	defer server.Close()

	svc := newTestService(server.URL, 10)
	_, err := svc.GetMarketBySlug("missing")
	if err == nil {
		t.Fatal("GetMarketBySlug() error = nil, want error")
	}
	if got := err.Error(); got != "polymarket api error: market not found" {
		t.Fatalf("GetMarketBySlug() error = %s, want market not found", got)
	}
}

func TestServiceFetchMarketsPaginatesWhenLimitZero(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		calls++
		cursor := r.URL.Query().Get("next_cursor")
		if calls == 1 {
			if cursor != InitialCursor {
				t.Fatalf("first cursor = %s, want %s", cursor, InitialCursor)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"limit":1,"count":1,"next_cursor":"MQ==","data":[{"id":"1","slug":"m1","event_slug":"e1","event_title":"E1"}]}`))
			return
		}
		if calls == 2 {
			if cursor != "MQ==" {
				t.Fatalf("second cursor = %s, want MQ==", cursor)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"limit":1,"count":1,"next_cursor":"LTE=","data":[{"id":"2","slug":"m2","event_slug":"e2","event_title":"E2"}]}`))
			return
		}
		t.Fatalf("unexpected call count: %d", calls)
	}))
	defer server.Close()

	svc := newTestService(server.URL, 10)
	events, err := svc.ListEvents(0)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("ListEvents() len = %d, want 2", len(events))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestServiceMapSDKErrorStatusFallback(t *testing.T) {
	svc := newTestService("https://clob.polymarket.com", 10)
	err := svc.mapSDKError(&ApiError{Status: 500})
	if got := fmt.Sprint(err); got != "polymarket api status: 500" {
		t.Fatalf("mapSDKError() = %s, want polymarket api status: 500", got)
	}
}

func TestServiceListAddressUnresolvedMarketsMissingCreds(t *testing.T) {
	svc := newTestService("https://clob.polymarket.com", 10)
	_, err := svc.ListAddressUnresolvedMarkets("0x1111111111111111111111111111111111111111", 5)
	if err != ErrMissingL2Credentials {
		t.Fatalf("ListAddressUnresolvedMarkets() error=%v, want %v", err, ErrMissingL2Credentials)
	}
}

func TestServiceListAddressUnresolvedMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/data/trades":
			if got := r.URL.Query().Get("maker_address"); got != "0x1111111111111111111111111111111111111111" {
				t.Fatalf("unexpected maker_address: %s", got)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"next_cursor":"LTE=","data":[{"market":"mkt-1"},{"market":"mkt-2"}]}`))
		case r.URL.Path == "/markets/mkt-1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"mkt-1","question":"Q1","slug":"s1","closed":false}`))
		case r.URL.Path == "/markets/mkt-2":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"mkt-2","question":"Q2","slug":"s2","closed":true}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL:  server.URL,
		ClobBaseURL:   server.URL,
		Timeout:       1,
		DefaultLimit:  10,
		Address:       "0x1111111111111111111111111111111111111111",
		ApiKey:        "k",
		ApiSecret:     "c2VjcmV0",
		ApiPassphrase: "p",
	}, &mocks.MockLogger{})

	markets, err := svc.ListAddressUnresolvedMarkets("0x1111111111111111111111111111111111111111", 10)
	if err != nil {
		t.Fatalf("ListAddressUnresolvedMarkets() error=%v", err)
	}
	if len(markets) != 1 {
		t.Fatalf("ListAddressUnresolvedMarkets() len=%d, want 1", len(markets))
	}
	if markets[0].ID != "mkt-1" {
		t.Fatalf("ListAddressUnresolvedMarkets() first id=%s, want mkt-1", markets[0].ID)
	}
}

func TestServiceCheckL2CredentialsMissingCreds(t *testing.T) {
	svc := newTestService("https://clob.polymarket.com", 10)
	if err := svc.CheckL2Credentials(); err != ErrMissingL2Credentials {
		t.Fatalf("CheckL2Credentials() error=%v, want %v", err, ErrMissingL2Credentials)
	}
}

func TestServiceCheckL2CredentialsUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/trades" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Unauthorized/Invalid api key"}`))
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL:  server.URL,
		ClobBaseURL:   server.URL,
		Timeout:       1,
		DefaultLimit:  10,
		Address:       "0x1111111111111111111111111111111111111111",
		ApiKey:        "k",
		ApiSecret:     "c2VjcmV0",
		ApiPassphrase: "p",
	}, &mocks.MockLogger{})

	err := svc.CheckL2Credentials()
	if err == nil {
		t.Fatal("CheckL2Credentials() error=nil, want error")
	}
	if got := err.Error(); got != "polymarket api error: Unauthorized/Invalid api key" {
		t.Fatalf("CheckL2Credentials() error=%s", got)
	}
}

func TestServiceCheckL2CredentialsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/trades" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"next_cursor":"LTE=","data":[]}`))
	}))
	defer server.Close()

	svc := NewService(config.Polymarket{
		GammaBaseURL:  server.URL,
		ClobBaseURL:   server.URL,
		Timeout:       1,
		DefaultLimit:  10,
		Address:       "0x1111111111111111111111111111111111111111",
		ApiKey:        "k",
		ApiSecret:     "c2VjcmV0",
		ApiPassphrase: "p",
	}, &mocks.MockLogger{})

	if err := svc.CheckL2Credentials(); err != nil {
		t.Fatalf("CheckL2Credentials() error=%v", err)
	}
}
