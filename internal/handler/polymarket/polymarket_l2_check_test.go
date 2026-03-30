package polymarket

import (
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/mocks"
)

func TestPolymarketL2CheckHandleCmd(t *testing.T) {
	h := NewPolymarketL2CheckHandle(newTestService("https://clob.polymarket.com", 10), &mocks.MockLogger{})
	if got := h.Cmd(); got != "/polymarket_l2_check" {
		t.Fatalf("Cmd() = %s, want /polymarket_l2_check", got)
	}
}
