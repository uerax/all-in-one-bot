package polymarket

import (
	"fmt"

	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"

	tb "gopkg.in/telebot.v4"
)

type PolymarketL2CheckHandle struct {
	Service *Service
	Logger  logger.Log
}

func NewPolymarketL2CheckHandle(service *Service, log logger.Log) *PolymarketL2CheckHandle {
	return &PolymarketL2CheckHandle{
		Service: service,
		Logger:  log,
	}
}

func (h *PolymarketL2CheckHandle) Cmd() string {
	return "/polymarket_l2_check"
}

func (h *PolymarketL2CheckHandle) Handle(c tb.Context) error {
	chat := c.Chat()
	h.Logger.Info(
		"command processed",
		"command", h.Cmd(),
		"chat_id", chat.ID,
		"chat_type", chat.Type,
	)

	if err := h.Service.CheckL2Credentials(); err != nil {
		return c.Send(fmt.Sprintf("%v", err))
	}

	return c.Send("Polymarket L2 credentials are valid")
}
