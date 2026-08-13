package coin

import (
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
)

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		price float64
		want  string
	}{
		{1234.56, "1234.56"},
		{0.001234, "0.001234"},
		{0.000000012345, "0.0000000123"},
	}

	for _, tt := range tests {
		got := formatPrice(tt.price)
		if got != tt.want {
			t.Errorf("formatPrice(%f) = %q, want %q", tt.price, got, tt.want)
		}
	}
}

func TestFormatUSD(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{1_500_000_000, "$1.50B"},
		{45_200_000, "$45.20M"},
		{850_500, "$850.50K"},
		{12.34, "$12.34"},
		{0, "$0.00"},
	}

	for _, tt := range tests {
		got := formatUSD(tt.v)
		if got != tt.want {
			t.Errorf("formatUSD(%f) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestFormatMarketData(t *testing.T) {
	data := &provider.MarketData{
		Symbol:          "PEPE",
		Name:            "Pepe",
		PriceUSD:        0.00000854,
		Change24h:       12.45,
		Volume24hUSD:    45230100,
		ReserveUSD:      12850000,
		FDV:             3587000000,
		Chain:           "eth",
		DEX:             "uniswap_v3",
		ContractAddress: "0x6982508145454ce325ddbe47a25d4ec3d2311933",
		Source:          "geckoterminal",
	}

	output := FormatMarketData(data)
	if output == "" {
		t.Fatal("FormatMarketData output is empty")
	}
}
