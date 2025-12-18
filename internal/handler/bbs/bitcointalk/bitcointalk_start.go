package bitcointalk

import (
	"context"

	tb "gopkg.in/telebot.v4"
)

type BitcointalkStartHandle struct {
	Bitcointalk *BitcointalkHandle
}

func NewBitcointalkStartHandle(bitcointalk *BitcointalkHandle) *BitcointalkStartHandle {
	return &BitcointalkStartHandle{
		Bitcointalk: bitcointalk,
	}
}

func (h *BitcointalkStartHandle) Cmd() string {
	return "/bitcointalk_start"
}

func (h *BitcointalkStartHandle) Handle(c tb.Context) error {
	chat := c.Chat()
	
	h.Bitcointalk.Logger.Info(
		"command processed",
		"command", h.Cmd(),
		"chat_id", chat.ID,
		"chat_type", chat.Type, // 例如：private, group
	)
	err := h.Bitcointalk.StartMonitor(context.Background(), chat.ID)

	if err != nil {
		c.Send(err.Error())
		h.Bitcointalk.Logger.Error(
			"Failed to send response",
			"chat_id", chat.ID,
			"error", err,
		)
	}

	return err
}