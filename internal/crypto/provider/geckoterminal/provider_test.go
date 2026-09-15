package geckoterminal

import (
	"testing"
)

func TestParseQuery(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		query          string
		expectedNet    string
		expectedAddr   string
		expectedIsPool bool
	}{
		{
			query:          "solana:JUPyiwrYJFskRFiEabWBmtvGXdphMvK6Zidqs1gjb8N",
			expectedNet:    "solana",
			expectedAddr:   "JUPyiwrYJFskRFiEabWBmtvGXdphMvK6Zidqs1gjb8N",
			expectedIsPool: false,
		},
		{
			query:          "0x6982508145454ce325ddbe47a25d4ec3d2311933",
			expectedNet:    "eth",
			expectedAddr:   "0x6982508145454ce325ddbe47a25d4ec3d2311933",
			expectedIsPool: false,
		},
		{
			query:          "JUPyiwrYJFskRFiEabWBmtvGXdphMvK6Zidqs1gjb8N",
			expectedNet:    "solana",
			expectedAddr:   "JUPyiwrYJFskRFiEabWBmtvGXdphMvK6Zidqs1gjb8N",
			expectedIsPool: false,
		},
		{
			query:          "BTC",
			expectedNet:    "",
			expectedAddr:   "",
			expectedIsPool: false,
		},
	}

	for _, tt := range tests {
		net, addr, isPool := p.parseQuery(tt.query)
		if net != tt.expectedNet || addr != tt.expectedAddr || isPool != tt.expectedIsPool {
			t.Errorf("parseQuery(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.query, net, addr, isPool, tt.expectedNet, tt.expectedAddr, tt.expectedIsPool)
		}
	}
}

func TestExtractNetwork(t *testing.T) {
	tests := []struct {
		name     string
		pool     *PoolData
		expected string
	}{
		{
			name:     "nil pool",
			pool:     nil,
			expected: "",
		},
		{
			name: "from pool.Relationships.Network",
			pool: &PoolData{
				ID: "base_0x123",
				Relationships: PoolRelationships{
					Network: RelationshipItem{
						Data: RelationshipData{ID: "ethereum"},
					},
				},
			},
			expected: "ethereum",
		},
		{
			name: "from pool.ID (GeckoTerminal standard format)",
			pool: &PoolData{
				ID: "base_0xb099c658e784b41ee435d48a8eb67e8f27285c93",
			},
			expected: "base",
		},
		{
			name: "from pool.Relationships.BaseToken",
			pool: &PoolData{
				ID: "unknownformat",
				Relationships: PoolRelationships{
					BaseToken: RelationshipItem{
						Data: RelationshipData{ID: "solana_58oQChx4yWmvKdwLLZzBi4ChoCc2fqCUWBkwMihLYQo2"},
					},
				},
			},
			expected: "solana",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNetwork(tt.pool)
			if got != tt.expected {
				t.Errorf("extractNetwork() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseKlineQuery(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		query          string
		expectedNet    string
		expectedAddr   string
		expectedSearch string
	}{
		{
			query:          "base:auki",
			expectedNet:    "base",
			expectedAddr:   "",
			expectedSearch: "auki",
		},
		{
			query:          "eth:wquil",
			expectedNet:    "eth",
			expectedAddr:   "",
			expectedSearch: "wquil",
		},
		{
			query:          "base:0x2fa9d6085c91151200e61a3e627d35001772c0d1",
			expectedNet:    "base",
			expectedAddr:   "0x2fa9d6085c91151200e61a3e627d35001772c0d1",
			expectedSearch: "",
		},
		{
			query:          "0x6982508145454ce325ddbe47a25d4ec3d2311933",
			expectedNet:    "eth",
			expectedAddr:   "0x6982508145454ce325ddbe47a25d4ec3d2311933",
			expectedSearch: "",
		},
		{
			query:          "auki",
			expectedNet:    "",
			expectedAddr:   "",
			expectedSearch: "auki",
		},
	}

	for _, tt := range tests {
		net, addr, search := p.parseKlineQuery(tt.query)
		if net != tt.expectedNet || addr != tt.expectedAddr || search != tt.expectedSearch {
			t.Errorf("parseKlineQuery(%q) = (%q, %q, %q), want (%q, %q, %q)",
				tt.query, net, addr, search, tt.expectedNet, tt.expectedAddr, tt.expectedSearch)
		}
	}
}
