package geckoterminal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	cache      *ttlCache
	limiter    *rateLimiter
	log        logger.Log
}

func NewClient(baseURL string, timeoutSec int, log logger.Log) *Client {
	if baseURL == "" {
		baseURL = "https://api.geckoterminal.com/api/v2"
	}
	if timeoutSec <= 0 {
		timeoutSec = 15
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
		cache:   newTTLCache(30 * time.Second),
		limiter: newRateLimiter(25), // 25 req/min (under 30 req/min limit)
		log:     log,
	}
}

func (c *Client) doGet(endpoint string) ([]byte, error) {
	if cached, ok := c.cache.Get(endpoint); ok {
		return cached.([]byte), nil
	}

	if !c.limiter.Allow() {
		return nil, provider.ErrRateLimited
	}

	fullURL := c.baseURL + endpoint
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("geckoterminal: create request error: %w", err)
	}

	req.Header.Set("Accept", "application/json;version=20230302")
	req.Header.Set("User-Agent", "all-in-one-bot/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geckoterminal: http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, provider.ErrRateLimited
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, provider.ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("geckoterminal: http status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("geckoterminal: read body error: %w", err)
	}

	c.cache.Set(endpoint, body)
	return body, nil
}

// SearchPools queries DEX pools by symbol, name, or contract address.
func (c *Client) SearchPools(query string) (*PoolListResponse, error) {
	endpoint := fmt.Sprintf("/search/pools?query=%s", url.QueryEscape(query))
	data, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var res PoolListResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("geckoterminal: unmarshal search error: %w", err)
	}
	return &res, nil
}

// GetTokenPools fetches top pools for a given token contract on a network.
func (c *Client) GetTokenPools(network, tokenAddress string) (*PoolListResponse, error) {
	endpoint := fmt.Sprintf("/networks/%s/tokens/%s/pools", network, tokenAddress)
	data, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var res PoolListResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("geckoterminal: unmarshal token pools error: %w", err)
	}
	return &res, nil
}

// GetPool fetches detailed data for a specific DEX pool.
func (c *Client) GetPool(network, poolAddress string) (*PoolResponse, error) {
	endpoint := fmt.Sprintf("/networks/%s/pools/%s", network, poolAddress)
	data, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var res PoolResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("geckoterminal: unmarshal pool error: %w", err)
	}
	return &res, nil
}

// GetTrendingPools fetches top trending pools on a network or globally.
func (c *Client) GetTrendingPools(network string) (*PoolListResponse, error) {
	endpoint := "/networks/trending_pools"
	if network != "" {
		endpoint = fmt.Sprintf("/networks/%s/trending_pools", network)
	}

	data, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var res PoolListResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("geckoterminal: unmarshal trending pools error: %w", err)
	}
	return &res, nil
}

// GetPoolOHLCV fetches daily/intraday K-lines for a pool.
func (c *Client) GetPoolOHLCV(network, poolAddress, timeframe string, limit int) (*OHLCVResponse, error) {
	if timeframe == "" {
		timeframe = "day"
	}
	if limit <= 0 {
		limit = 100
	}

	endpoint := fmt.Sprintf("/networks/%s/pools/%s/ohlcv/%s?limit=%d", network, poolAddress, timeframe, limit)
	data, err := c.doGet(endpoint)
	if err != nil {
		return nil, err
	}

	var res OHLCVResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("geckoterminal: unmarshal ohlcv error: %w", err)
	}
	return &res, nil
}
