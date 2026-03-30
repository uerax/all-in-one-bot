package polymarket

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"

	tb "gopkg.in/telebot.v4"
)

type PolymarketHoldingsHandle struct {
	Service *Service
	Logger  logger.Log
}

func NewPolymarketHoldingsHandle(service *Service, log logger.Log) *PolymarketHoldingsHandle {
	return &PolymarketHoldingsHandle{
		Service: service,
		Logger:  log,
	}
}

func (h *PolymarketHoldingsHandle) Cmd() string {
	return "/polymarket_holdings"
}

func (h *PolymarketHoldingsHandle) Handle(c tb.Context) error {
	chat := c.Chat()
	h.Logger.Info(
		"command processed",
		"command", h.Cmd(),
		"chat_id", chat.ID,
		"chat_type", chat.Type,
	)

	address, limit, err := parseHoldingsArgs(extractPayload(c))
	if err != nil {
		return c.Send("Usage: /polymarket_holdings <address> [limit]")
	}

	markets, err := h.Service.ListAddressUnresolvedMarkets(address, limit)
	if err != nil {
		return c.Send(fmt.Sprintf("%v", err))
	}
	if len(markets) == 0 {
		return c.Send(fmt.Sprintf("No unresolved markets found for %s", address))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Address: %s\n", strings.ToLower(address)))
	b.WriteString(fmt.Sprintf("Unresolved markets: %d\n\n", len(markets)))
	for i, market := range markets {
		b.WriteString(fmt.Sprintf("%d) %s\n", i+1, market.Question))
		if market.Slug != "" {
			b.WriteString(fmt.Sprintf("   slug: %s\n", market.Slug))
		}
		b.WriteString(fmt.Sprintf("   id: %s\n", market.ID))
	}

	return c.Send(b.String())
}

func extractPayload(c tb.Context) string {
	if c == nil || c.Message() == nil {
		return ""
	}
	payload := strings.TrimSpace(c.Message().Payload)
	if payload != "" {
		return payload
	}
	text := strings.TrimSpace(c.Message().Text)
	if text == "" {
		return ""
	}
	parts := strings.Fields(text)
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[1:], " ")
}

func parseHoldingsArgs(payload string) (string, int, error) {
	parts := strings.Fields(strings.TrimSpace(payload))
	if len(parts) < 1 || len(parts) > 2 {
		return "", 0, fmt.Errorf("invalid args")
	}

	address := strings.TrimSpace(parts[0])
	if _, err := normalizeAddress(address); err != nil {
		return "", 0, err
	}

	if len(parts) == 1 {
		return address, 0, nil
	}

	limit, err := strconv.Atoi(parts[1])
	if err != nil || limit <= 0 {
		return "", 0, fmt.Errorf("invalid limit")
	}

	return address, limit, nil
}
