package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	pm "github.com/uerax/polymarket-go/polymarket"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

var ErrInvalidPriceSide = errors.New("invalid price side")
var ErrInvalidAddress = errors.New("invalid address")
var ErrMissingL2Credentials = errors.New("polymarket l2 credentials not configured")

const (
	PriceSideBuy  = "buy"
	PriceSideSell = "sell"
)

type Service struct {
	defaultLimit int
	gammaClient  *pm.Client
	clobClient   *pm.Client
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

func NewService(cfg config.Polymarket, log logger.Log) *Service {
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
		Transport: &stripAcceptEncodingTransport{
			base: http.DefaultTransport,
		},
	}

	gammaClient := pm.NewClient(
		strings.TrimRight(cfg.GammaBaseURL, "/"),
		pm.ChainPolygon,
		nil,
		nil,
		pm.WithHTTPClient(httpClient),
		pm.WithThrowOnError(true),
	)

	var signer pm.ClobSigner
	var creds *pm.ApiKeyCreds
	if strings.TrimSpace(cfg.Address) != "" && strings.TrimSpace(cfg.ApiKey) != "" && strings.TrimSpace(cfg.ApiSecret) != "" && strings.TrimSpace(cfg.ApiPassphrase) != "" {
		signer = &staticAddressSigner{address: strings.TrimSpace(cfg.Address)}
		creds = &pm.ApiKeyCreds{
			Key:        strings.TrimSpace(cfg.ApiKey),
			Secret:     strings.TrimSpace(cfg.ApiSecret),
			Passphrase: strings.TrimSpace(cfg.ApiPassphrase),
		}
	}

	clobClient := pm.NewClient(
		strings.TrimRight(cfg.ClobBaseURL, "/"),
		pm.ChainPolygon,
		signer,
		creds,
		pm.WithHTTPClient(httpClient),
		pm.WithThrowOnError(true),
	)

	return &Service{
		defaultLimit: cfg.DefaultLimit,
		gammaClient:  gammaClient,
		clobClient:   clobClient,
		l2Enabled:    signer != nil && creds != nil,
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
	res, err := s.clobClient.GetMarket(context.Background(), id)
	if err != nil {
		s.log.Error("polymarket get market by id failed", "id", id, "error", err)
		return nil, s.mapSDKError(err)
	}

	market, err := decodeAny[Market](res)
	if err != nil {
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

	res, err := s.clobClient.GetPrice(context.Background(), tokenID, side)
	if err != nil {
		s.log.Error("polymarket get token price failed", "token_id", tokenID, "side", side, "error", err)
		return nil, s.mapSDKError(err)
	}

	price, err := decodeAny[Price](res)
	if err != nil {
		s.log.Error("polymarket decode price failed", "token_id", tokenID, "side", side, "error", err)
		return nil, err
	}

	return &price, nil
}

func (s *Service) GetOrderBook(tokenID string) (*OrderBook, error) {
	book, err := s.clobClient.GetOrderBook(context.Background(), tokenID)
	if err != nil {
		s.log.Error("polymarket get order book failed", "token_id", tokenID, "error", err)
		return nil, s.mapSDKError(err)
	}

	return &OrderBook{
		Market:         book.Market,
		AssetID:        book.AssetID,
		Timestamp:      book.Timestamp,
		Hash:           book.Hash,
		Bids:           mapOrderLevels(book.Bids),
		Asks:           mapOrderLevels(book.Asks),
		MinOrderSize:   book.MinOrderSize,
		TickSize:       book.TickSize,
		NegRisk:        book.NegRisk,
		LastTradePrice: book.LastTradePrice,
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

	trades, err := s.clobClient.GetTrades(context.Background(), &pm.TradeParams{MakerAddress: normalized}, false, pm.InitialCursor)
	if err != nil {
		s.log.Error("polymarket get trades failed", "address", normalized, "error", err)
		return nil, s.mapSDKError(err)
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
	_, err := s.clobClient.GetTrades(context.Background(), &pm.TradeParams{}, false, pm.InitialCursor)
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

	cursor := pm.InitialCursor
	initialCap := s.normalizeLimit(0)
	if !fetchAll {
		initialCap = resolvedLimit
	}
	markets := make([]map[string]any, 0, initialCap)

	for {
		payload, err := s.clobClient.GetMarkets(context.Background(), cursor)
		if err != nil {
			s.log.Error("polymarket list markets failed", "cursor", cursor, "error", err)
			return nil, s.mapSDKError(err)
		}

		batch, err := decodeAny[[]map[string]any](payload.Data)
		if err != nil {
			s.log.Error("polymarket decode markets failed", "cursor", cursor, "error", err)
			return nil, err
		}

		markets = append(markets, batch...)

		if !fetchAll && len(markets) >= resolvedLimit {
			return markets[:resolvedLimit], nil
		}

		next := strings.TrimSpace(payload.NextCursor)
		if next == "" || next == pm.EndCursor {
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
	var apiErr *pm.ApiError
	if errors.As(err, &apiErr) {
		msg := strings.TrimSpace(apiErr.Message)
		if msg != "" {
			return fmt.Errorf("polymarket api error: %s", msg)
		}
		if apiErr.Status > 0 {
			return fmt.Errorf("polymarket api status: %d", apiErr.Status)
		}
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

type staticAddressSigner struct {
	address string
}

func (s *staticAddressSigner) Address(ctx context.Context) (string, error) {
	_ = ctx
	return s.address, nil
}

func (s *staticAddressSigner) SignClobAuth(ctx context.Context, chainID pm.Chain, timestamp int64, nonce int64) (string, error) {
	_ = ctx
	_ = chainID
	_ = timestamp
	_ = nonce
	return "", nil
}

func (s *staticAddressSigner) SignOrderTypedData(ctx context.Context, payload pm.OrderTypedDataPayload) (string, error) {
	_ = ctx
	_ = payload
	return "", nil
}

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

func mapOrderLevels(levels []pm.OrderSummary) []BookLevel {
	if len(levels) == 0 {
		return nil
	}
	mapped := make([]BookLevel, 0, len(levels))
	for _, level := range levels {
		mapped = append(mapped, BookLevel{Price: level.Price, Size: level.Size})
	}
	return mapped
}
