package polymarket

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

var ErrInvalidPriceSide = errors.New("invalid price side")

const (
	PriceSideBuy  = "buy"
	PriceSideSell = "sell"
)

type Service struct {
	gammaBaseURL string
	clobBaseURL  string
	defaultLimit int
	client       *http.Client
	log          logger.Log
}

type ListMarketsOptions struct {
	Limit  int
	Active *bool
	Closed *bool
	Query  string
}

type Market struct {
	ID            string  `json:"id"`
	Question      string  `json:"question"`
	Slug          string  `json:"slug"`
	Description   string  `json:"description"`
	EndDate       string  `json:"endDate"`
	Active        bool    `json:"active"`
	Closed        bool    `json:"closed"`
	Volume        string  `json:"volume"`
	Liquidity     string  `json:"liquidity"`
	VolumeNum     float64 `json:"volumeNum"`
	LiquidityNum  float64 `json:"liquidityNum"`
	Outcomes      string  `json:"outcomes"`
	OutcomePrices string  `json:"outcomePrices"`
	ClobTokenIDs  string  `json:"clobTokenIds"`
	BestBid       float64 `json:"bestBid"`
	BestAsk       float64 `json:"bestAsk"`
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
	Market        string      `json:"market"`
	AssetID       string      `json:"asset_id"`
	Timestamp     string      `json:"timestamp"`
	Hash          string      `json:"hash"`
	Bids          []BookLevel `json:"bids"`
	Asks          []BookLevel `json:"asks"`
	MinOrderSize  string      `json:"min_order_size"`
	TickSize      string      `json:"tick_size"`
	NegRisk       bool        `json:"neg_risk"`
	LastTradePrice string     `json:"last_trade_price"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewService(cfg config.Polymarket, log logger.Log) *Service {
	return &Service{
		gammaBaseURL: strings.TrimRight(cfg.GammaBaseURL, "/"),
		clobBaseURL:  strings.TrimRight(cfg.ClobBaseURL, "/"),
		defaultLimit: cfg.DefaultLimit,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		log: log,
	}
}

func (s *Service) ListMarkets(opts ListMarketsOptions) ([]Market, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(s.normalizeLimit(opts.Limit)))
	if opts.Active != nil {
		query.Set("active", strconv.FormatBool(*opts.Active))
	}
	if opts.Closed != nil {
		query.Set("closed", strconv.FormatBool(*opts.Closed))
	}

	var markets []Market
	if err := s.getJSON(s.gammaURL("/markets", query), &markets); err != nil {
		return nil, err
	}

	if opts.Query == "" {
		return markets, nil
	}

	return filterMarkets(markets, opts.Query), nil
}

func (s *Service) SearchMarkets(query string, limit int) ([]Market, error) {
	return s.ListMarkets(ListMarketsOptions{
		Limit: limit,
		Query: query,
	})
}

func (s *Service) GetMarketByID(id string) (*Market, error) {
	var market Market
	if err := s.getJSON(s.gammaURL("/markets/"+url.PathEscape(id), nil), &market); err != nil {
		return nil, err
	}
	return &market, nil
}

func (s *Service) GetMarketBySlug(slug string) (*Market, error) {
	var market Market
	if err := s.getJSON(s.gammaURL("/markets/slug/"+url.PathEscape(slug), nil), &market); err != nil {
		return nil, err
	}
	return &market, nil
}

func (s *Service) ListEvents(limit int) ([]Event, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(s.normalizeLimit(limit)))

	var events []Event
	if err := s.getJSON(s.gammaURL("/events", query), &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Service) GetTokenPrice(tokenID string, side string) (*Price, error) {
	if side != PriceSideBuy && side != PriceSideSell {
		return nil, ErrInvalidPriceSide
	}

	query := url.Values{}
	query.Set("token_id", tokenID)
	query.Set("side", side)

	var price Price
	if err := s.getJSON(s.clobURL("/price", query), &price); err != nil {
		return nil, err
	}
	return &price, nil
}

func (s *Service) GetOrderBook(tokenID string) (*OrderBook, error) {
	query := url.Values{}
	query.Set("token_id", tokenID)

	var book OrderBook
	if err := s.getJSON(s.clobURL("/book", query), &book); err != nil {
		return nil, err
	}
	return &book, nil
}

func (s *Service) gammaURL(path string, query url.Values) string {
	return buildURL(s.gammaBaseURL, path, query)
}

func (s *Service) clobURL(path string, query url.Values) string {
	return buildURL(s.clobBaseURL, path, query)
}

func buildURL(base string, path string, query url.Values) string {
	u := base + path
	if len(query) == 0 {
		return u
	}
	return u + "?" + query.Encode()
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

func (s *Service) getJSON(rawURL string, target any) error {
	resp, err := s.client.Get(rawURL)
	if err != nil {
		s.log.Error("polymarket request failed", "url", rawURL, "error", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Error("polymarket read body failed", "url", rawURL, "error", err)
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr errorResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != "" {
			return fmt.Errorf("polymarket api error: %s", apiErr.Error)
		}
		return fmt.Errorf("polymarket api status: %d", resp.StatusCode)
	}

	if err := json.Unmarshal(body, target); err != nil {
		s.log.Error("polymarket decode failed", "url", rawURL, "error", err)
		return err
	}

	return nil
}
