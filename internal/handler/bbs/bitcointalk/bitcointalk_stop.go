package bitcointalk

import (
	"fmt"

	tb "gopkg.in/telebot.v4"
)

type BitcointalkStopHandle struct {
	Bitcointalk *BitcointalkHandle
}

func NewBitcointalkStopHandle(bitcointalk *BitcointalkHandle) *BitcointalkStopHandle {
	return &BitcointalkStopHandle{
		Bitcointalk: bitcointalk,
	}
}

func (h *BitcointalkStopHandle) Cmd() string {
	return "/bitcointalk_start"
}

func (h *BitcointalkStopHandle) Handle(c tb.Context) error {
	chat := c.Chat()

	h.Bitcointalk.Logger.Info(
		"command processed",
		"command", h.Cmd(),
		"chat_id", chat.ID,
		"chat_type", chat.Type, // 例如：private, group
	)

	err := c.Send(fmt.Sprintf("This chat ID is: %d", chat.ID))

	if err != nil {
		h.Bitcointalk.Logger.Error(
			"Failed to send response",
			"chat_id", chat.ID,
			"error", err,
		)
	}

	return err
}
