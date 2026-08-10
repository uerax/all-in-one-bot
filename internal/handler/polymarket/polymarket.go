package polymarket

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

var ErrInvalidPriceSide = errors.New("invalid price side")
var ErrInvalidAddress = errors.New("invalid address")
var ErrMissingL2Credentials = errors.New("polymarket l2 credentials not configured")

const (
	PriceSideBuy  = "buy"
	PriceSideSell = "sell"
	InitialCursor = "MA=="
	EndCursor     = "LTE="
)

type ApiError struct {
	Message string `json:"error"`
	Status  int    `json:"-"`
}

func (e *ApiError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("polymarket api error: %s", e.Message)
	}
	if e.Status > 0 {
		return fmt.Sprintf("polymarket api status: %d", e.Status)
	}
	return "polymarket api error"
}

type Client struct {
	baseURL       string
	httpClient    *http.Client
	address       string
	apiKey        string
	apiSecret     string
	apiPassphrase string
}

func NewClient(baseURL string, httpClient *http.Client, address, apiKey, apiSecret, apiPassphrase string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:       strings.TrimRight(baseURL, "/"),
		httpClient:    httpClient,
		address:       strings.TrimSpace(address),
		apiKey:        strings.TrimSpace(apiKey),
		apiSecret:     strings.TrimSpace(apiSecret),
		apiPassphrase: strings.TrimSpace(apiPassphrase),
	}
}

func (c *Client) hasL2Creds() bool {
	return c.address != "" && c.apiKey != "" && c.apiSecret != "" && c.apiPassphrase != ""
}

func (c *Client) request(ctx context.Context, method, reqPath string, query map[string]string, auth bool) ([]byte, error) {
	fullURL := c.baseURL + reqPath
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	if len(query) > 0 {
		q := req.URL.Query()
		for k, v := range query {
			if v != "" {
				q.Set(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	if auth && c.hasL2Creds() {
		headers, err := createL2Headers(c.address, c.apiKey, c.apiSecret, c.apiPassphrase, method, reqPath, nil)
		if err == nil {
			for k, v := range headers {
				req.Header[k] = v
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr ApiError
		apiErr.Status = resp.StatusCode
		if jsonErr := json.Unmarshal(body, &apiErr); jsonErr == nil && apiErr.Message != "" {
			return nil, &apiErr
		}
		return nil, &apiErr
	}

	return body, nil
}

func createL2Headers(address, apiKey, apiSecret, apiPassphrase, method, requestPath string, body []byte) (http.Header, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	secretStr := strings.ReplaceAll(strings.ReplaceAll(apiSecret, "-", "+"), "_", "/")
	secretBytes, err := base64.StdEncoding.DecodeString(secretStr)
	if err != nil {
		secretBytes = []byte(apiSecret)
	}

	sigMessage := timestamp + method + requestPath
	if len(body) > 0 {
		sigMessage += string(body)
	}

	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(sigMessage))
	rawSig := h.Sum(nil)

	sig := base64.StdEncoding.EncodeToString(rawSig)
	sig = strings.ReplaceAll(strings.ReplaceAll(sig, "+", "-"), "/", "_")

	header := make(http.Header)
	header.Set("POLY_ADDRESS", address)
	header.Set("POLY_SIGNATURE", sig)
	header.Set("POLY_TIMESTAMP", timestamp)
	header.Set("POLY_API_KEY", apiKey)
	header.Set("POLY_PASSPHRASE", apiPassphrase)

	return header, nil
}

type PagedResponse struct {
	NextCursor string          `json:"next_cursor"`
	Data       json.RawMessage `json:"data"`
}

type Service struct {
	defaultLimit int
	clobClient   *Client
	l2Enabled    bool
	log          logger.Log
}

type ListMarketsOptions struct {
	Limit  int
	Active *bool
	Closed *bool
	Query  string
}

type Market struct {
	ID             string  `json:"id"`
	Question       string  `json:"question"`
	Slug           string  `json:"slug"`
	Description    string  `json:"description"`
	EndDate        string  `json:"endDate"`
	Active         bool    `json:"active"`
	Closed         bool    `json:"closed"`
	Volume         string  `json:"volume"`
	Liquidity      string  `json:"liquidity"`
	VolumeNum      float64 `json:"volumeNum"`
	LiquidityNum   float64 `json:"liquidityNum"`
	Outcomes       string  `json:"outcomes"`
	OutcomePrices  string  `json:"outcomePrices"`
	ClobTokenIDs   string  `json:"clobTokenIds"`
	BestBid        float64 `json:"bestBid"`
	BestAsk        float64 `json:"bestAsk"`
	LastTradePrice float64 `json:"lastTradePrice"`
}

type Event struct {
	ID          string   `json:"id"`
	Ticker      string   `json:"ticker"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	Category    string   `json:"category"`
	Active      bool     `json:"active"`
	Closed      bool     `json:"closed"`
	Markets     []Market `json:"markets"`
}

type Price struct {
	Price string `json:"price"`
}

type BookLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

type OrderBook struct {
	Market         string      `json:"market"`
	AssetID        string      `json:"asset_id"`
	Timestamp      string      `json:"timestamp"`
	Hash           string      `json:"hash"`
	Bids           []BookLevel `json:"bids"`
	Asks           []BookLevel `json:"asks"`
	MinOrderSize   string      `json:"min_order_size"`
	TickSize       string      `json:"tick_size"`
	NegRisk        bool        `json:"neg_risk"`
	LastTradePrice string      `json:"last_trade_price"`
}

type rawOrderBook struct {
	Market         string      `json:"market"`
	AssetID        string      `json:"asset_id"`
	Timestamp      string      `json:"timestamp"`
	Hash           string      `json:"hash"`
	Bids           []BookLevel `json:"bids"`
	Asks           []BookLevel `json:"asks"`
	MinOrderSize   string      `json:"min_order_size"`
	TickSize       string      `json:"tick_size"`
	NegRisk        bool        `json:"neg_risk"`
	LastTradePrice string      `json:"last_trade_price"`
}

type TradeItem struct {
	Market string `json:"market"`
}

func NewService(cfg config.Polymarket, log logger.Log) *Service {
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
		Transport: &stripAcceptEncodingTransport{
			base: http.DefaultTransport,
		},
	}

	clobClient := NewClient(
		cfg.ClobBaseURL,
		httpClient,
		cfg.Address,
		cfg.ApiKey,
		cfg.ApiSecret,
		cfg.ApiPassphrase,
	)

	return &Service{
		defaultLimit: cfg.DefaultLimit,
		clobClient:   clobClient,
		l2Enabled:    clobClient.hasL2Creds(),
		log:          log,
	}
}

func (s *Service) ListMarkets(opts ListMarketsOptions) ([]Market, error) {
	limit := s.normalizeLimit(opts.Limit)
	rawMarkets, err := s.fetchRawMarkets(limit)
	if err != nil {
		return nil, err
	}

	allMarkets, err := decodeMarkets(rawMarkets)
	if err != nil {
		return nil, err
	}

	markets := applyMarketFlags(allMarkets, opts.Active, opts.Closed)
	markets = filterMarkets(markets, opts.Query)

	if len(markets) > limit {
		markets = markets[:limit]
	}

	return markets, nil
}

func (s *Service) SearchMarkets(query string, limit int) ([]Market, error) {
	return s.ListMarkets(ListMarketsOptions{
		Limit: limit,
		Query: query,
	})
}

func (s *Service) GetMarketByID(id string) (*Market, error) {
	body, err := s.clobClient.request(context.Background(), http.MethodGet, "/markets/"+id, nil, false)
	if err != nil {
		s.log.Error("polymarket get market by id failed", "id", id, "error", err)
		return nil, s.mapSDKError(err)
	}

	var market Market
	if err := json.Unmarshal(body, &market); err != nil {
		s.log.Error("polymarket decode market failed", "id", id, "error", err)
		return nil, err
	}

	return &market, nil
}

func (s *Service) GetMarketBySlug(slug string) (*Market, error) {
	rawMarkets, err := s.fetchRawMarkets(0)
	if err != nil {
		return nil, err
	}

	markets, err := decodeMarkets(rawMarkets)
	if err != nil {
		return nil, err
	}

	for _, market := range markets {
		if market.Slug == slug {
			m := market
			return &m, nil
		}
	}

	return nil, fmt.Errorf("polymarket api error: market not found")
}

func (s *Service) ListEvents(limit int) ([]Event, error) {
	rawMarkets, err := s.fetchRawMarkets(0)
	if err != nil {
		return nil, err
	}

	type rawEventFields struct {
		EventID          string `json:"event_id"`
		EventSlug        string `json:"event_slug"`
		EventTicker      string `json:"event_ticker"`
		EventTitle       string `json:"event_title"`
		EventDescription string `json:"event_description"`
		EventStartDate   string `json:"event_start_date"`
		EventEndDate     string `json:"event_end_date"`
		EventCategory    string `json:"event_category"`
	}

	groups := make(map[string]*Event)
	order := make([]string, 0)

	for _, raw := range rawMarkets {
		market, err := decodeAny[Market](raw)
		if err != nil {
			continue
		}

		meta, err := decodeAny[rawEventFields](raw)
		if err != nil {
			continue
		}

		key := strings.TrimSpace(meta.EventSlug)
		if key == "" {
			key = strings.TrimSpace(meta.EventID)
		}
		if key == "" {
			key = strings.TrimSpace(meta.EventTitle)
		}
		if key == "" {
			continue
		}

		evt, ok := groups[key]
		if !ok {
			evt = &Event{
				ID:          strings.TrimSpace(meta.EventID),
				Ticker:      strings.TrimSpace(meta.EventTicker),
				Slug:        strings.TrimSpace(meta.EventSlug),
				Title:       strings.TrimSpace(meta.EventTitle),
				Description: strings.TrimSpace(meta.EventDescription),
				StartDate:   strings.TrimSpace(meta.EventStartDate),
				EndDate:     strings.TrimSpace(meta.EventEndDate),
				Category:    strings.TrimSpace(meta.EventCategory),
				Active:      market.Active,
				Closed:      market.Closed,
				Markets:     make([]Market, 0, 2),
			}
			groups[key] = evt
			order = append(order, key)
		}

		evt.Active = evt.Active || market.Active
		evt.Closed = evt.Closed && market.Closed
		evt.Markets = append(evt.Markets, market)
	}

	if len(order) == 0 {
		return []Event{}, nil
	}

	resolvedLimit := s.normalizeLimit(limit)
	if limit <= 0 {
		resolvedLimit = len(order)
	}
	if resolvedLimit > len(order) {
		resolvedLimit = len(order)
	}

	events := make([]Event, 0, resolvedLimit)
	for i := 0; i < resolvedLimit; i++ {
		events = append(events, *groups[order[i]])
	}

	return events, nil
}

func (s *Service) GetTokenPrice(tokenID string, side string) (*Price, error) {
	if side != PriceSideBuy && side != PriceSideSell {
		return nil, ErrInvalidPriceSide
	}

	query := map[string]string{
		"token_id": tokenID,
		"side":     side,
	}

	body, err := s.clobClient.request(context.Background(), http.MethodGet, "/price", query, false)
	if err != nil {
		s.log.Error("polymarket get token price failed", "token_id", tokenID, "side", side, "error", err)
		return nil, s.mapSDKError(err)
	}

	var price Price
	if err := json.Unmarshal(body, &price); err != nil {
		s.log.Error("polymarket decode price failed", "token_id", tokenID, "side", side, "error", err)
		return nil, err
	}

	return &price, nil
}

func (s *Service) GetOrderBook(tokenID string) (*OrderBook, error) {
	query := map[string]string{
		"token_id": tokenID,
	}

	body, err := s.clobClient.request(context.Background(), http.MethodGet, "/book", query, false)
	if err != nil {
		s.log.Error("polymarket get order book failed", "token_id", tokenID, "error", err)
		return nil, s.mapSDKError(err)
	}

	var raw rawOrderBook
	if err := json.Unmarshal(body, &raw); err != nil {
		s.log.Error("polymarket decode order book failed", "token_id", tokenID, "error", err)
		return nil, err
	}

	return &OrderBook{
		Market:         raw.Market,
		AssetID:        raw.AssetID,
		Timestamp:      raw.Timestamp,
		Hash:           raw.Hash,
		Bids:           raw.Bids,
		Asks:           raw.Asks,
		MinOrderSize:   raw.MinOrderSize,
		TickSize:       raw.TickSize,
		NegRisk:        raw.NegRisk,
		LastTradePrice: raw.LastTradePrice,
	}, nil
}

func (s *Service) ListAddressUnresolvedMarkets(address string, limit int) ([]Market, error) {
	normalized, err := normalizeAddress(address)
	if err != nil {
		return nil, err
	}
	if !s.l2Enabled {
		return nil, ErrMissingL2Credentials
	}

	query := map[string]string{
		"maker_address": normalized,
		"next_cursor":   InitialCursor,
	}

	body, err := s.clobClient.request(context.Background(), http.MethodGet, "/data/trades", query, true)
	if err != nil {
		s.log.Error("polymarket get trades failed", "address", normalized, "error", err)
		return nil, s.mapSDKError(err)
	}

	var paged PagedResponse
	if err := json.Unmarshal(body, &paged); err != nil {
		return nil, err
	}

	var trades []TradeItem
	if err := json.Unmarshal(paged.Data, &trades); err != nil {
		return nil, err
	}

	conditionIDs := make(map[string]struct{})
	for _, trade := range trades {
		conditionID := strings.TrimSpace(trade.Market)
		if conditionID == "" {
			continue
		}
		conditionIDs[conditionID] = struct{}{}
	}

	resolvedLimit := s.normalizeLimit(limit)
	markets := make([]Market, 0, resolvedLimit)
	for conditionID := range conditionIDs {
		market, getErr := s.GetMarketByID(conditionID)
		if getErr != nil {
			s.log.Error("polymarket get market by condition id failed", "condition_id", conditionID, "error", getErr)
			continue
		}
		if market.Closed {
			continue
		}
		markets = append(markets, *market)
		if len(markets) >= resolvedLimit {
			break
		}
	}

	return markets, nil
}

func (s *Service) CheckL2Credentials() error {
	if !s.l2Enabled {
		return ErrMissingL2Credentials
	}
	query := map[string]string{
		"next_cursor": InitialCursor,
	}
	_, err := s.clobClient.request(context.Background(), http.MethodGet, "/data/trades", query, true)
	if err != nil {
		s.log.Error("polymarket check l2 credentials failed", "error", err)
		return s.mapSDKError(err)
	}
	return nil
}

func (s *Service) fetchRawMarkets(limit int) ([]map[string]any, error) {
	resolvedLimit := limit
	fetchAll := resolvedLimit <= 0
	if !fetchAll {
		resolvedLimit = s.normalizeLimit(resolvedLimit)
	}

	cursor := InitialCursor
	initialCap := s.normalizeLimit(0)
	if !fetchAll {
		initialCap = resolvedLimit
	}
	markets := make([]map[string]any, 0, initialCap)

	for {
		query := map[string]string{
			"next_cursor": cursor,
		}
		body, err := s.clobClient.request(context.Background(), http.MethodGet, "/markets", query, false)
		if err != nil {
			s.log.Error("polymarket list markets failed", "cursor", cursor, "error", err)
			return nil, s.mapSDKError(err)
		}

		var payload PagedResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			s.log.Error("polymarket unmarshal markets page failed", "cursor", cursor, "error", err)
			return nil, err
		}

		var batch []map[string]any
		if err := json.Unmarshal(payload.Data, &batch); err != nil {
			s.log.Error("polymarket decode markets failed", "cursor", cursor, "error", err)
			return nil, err
		}

		markets = append(markets, batch...)

		if !fetchAll && len(markets) >= resolvedLimit {
			return markets[:resolvedLimit], nil
		}

		next := strings.TrimSpace(payload.NextCursor)
		if next == "" || next == EndCursor {
			break
		}
		cursor = next
	}

	return markets, nil
}

func decodeMarkets(rawMarkets []map[string]any) ([]Market, error) {
	if len(rawMarkets) == 0 {
		return []Market{}, nil
	}
	markets := make([]Market, 0, len(rawMarkets))
	for _, raw := range rawMarkets {
		market, err := decodeAny[Market](raw)
		if err != nil {
			return nil, err
		}
		markets = append(markets, market)
	}
	return markets, nil
}

func (s *Service) normalizeLimit(limit int) int {
	if limit > 0 {
		return limit
	}
	if s.defaultLimit > 0 {
		return s.defaultLimit
	}
	return 10
}

func applyMarketFlags(markets []Market, active *bool, closed *bool) []Market {
	if active == nil && closed == nil {
		return markets
	}

	filtered := make([]Market, 0, len(markets))
	for _, market := range markets {
		if active != nil && market.Active != *active {
			continue
		}
		if closed != nil && market.Closed != *closed {
			continue
		}
		filtered = append(filtered, market)
	}
	return filtered
}

func filterMarkets(markets []Market, query string) []Market {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return markets
	}

	filtered := make([]Market, 0, len(markets))
	for _, market := range markets {
		if strings.Contains(strings.ToLower(market.Question), query) || strings.Contains(strings.ToLower(market.Slug), query) {
			filtered = append(filtered, market)
		}
	}
	return filtered
}

func (s *Service) mapSDKError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return err
}

func decodeAny[T any](value any) (T, error) {
	var out T
	b, err := json.Marshal(value)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}

var addressPattern = regexp.MustCompile(`(?i)^0x[0-9a-f]{40}$`)

func normalizeAddress(address string) (string, error) {
	trimmed := strings.TrimSpace(address)
	if !addressPattern.MatchString(trimmed) {
		return "", ErrInvalidAddress
	}
	return strings.ToLower(trimmed), nil
}

type stripAcceptEncodingTransport struct {
	base http.RoundTripper
}

func (t *stripAcceptEncodingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	clone.Header.Del("Accept-Encoding")
	return base.RoundTrip(clone)
}
