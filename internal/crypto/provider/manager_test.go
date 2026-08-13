package provider

import (
	"testing"
)

func TestIsContractOrPool(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{"0x6982508145454ce325ddbe47a25d4ec3d2311933", true},
		{"solana:58oQChx4yWmvKdwLLZzBi4ChoCc2fqCUWBkwMihLYQo2", true},
		{"JUPyiwrYJFskRFiEabWBmtvGXdphMvK6Zidqs1gjb8N", true},
		{"BTC", false},
		{"PEPE", false},
		{"bitcoin", false},
	}

	for _, tt := range tests {
		got := isContractOrPool(tt.query)
		if got != tt.want {
			t.Errorf("isContractOrPool(%q) = %v, want %v", tt.query, got, tt.want)
		}
	}
}
