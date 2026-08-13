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
