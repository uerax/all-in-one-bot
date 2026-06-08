package tg

import (
	"strings"
	"testing"

	"github.com/uerax/all-in-one-bot/crypto"
)

func TestFormatCoingeckoSearchResults(t *testing.T) {
	rank := 1234
	msg := formatCoingeckoSearchResults("wquil", []crypto.CoingeckoSearchCoin{
		{ID: "wrapped-quil", Name: "Wrapped QUIL", Symbol: "wquil", MarketCapRank: &rank},
	})

	for _, want := range []string{
		"CoinGecko 搜索结果: `wquil`",
		"Wrapped QUIL",
		"symbol: WQUIL",
		"id: `wrapped-quil`",
		"rank: 1234",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message missing %q:\n%s", want, msg)
		}
	}
}

func TestFormatCoingeckoSearchResultsEmpty(t *testing.T) {
	msg := formatCoingeckoSearchResults("none", nil)
	if !strings.Contains(msg, "未找到 CoinGecko 币种: `none`") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestFormatCoingeckoSearchResultsLimitsOutput(t *testing.T) {
	coins := make([]crypto.CoingeckoSearchCoin, 11)
	for i := range coins {
		coins[i] = crypto.CoingeckoSearchCoin{
			ID:     "coin",
			Name:   "Coin",
			Symbol: "coin",
		}
	}

	msg := formatCoingeckoSearchResults("coin", coins)
	if strings.Count(msg, "\nid: `coin`") != 10 {
		t.Fatalf("expected 10 ids, got message:\n%s", msg)
	}
	if !strings.Contains(msg, "其余 1 条结果已省略") {
		t.Fatalf("expected omitted count, got:\n%s", msg)
	}
}
