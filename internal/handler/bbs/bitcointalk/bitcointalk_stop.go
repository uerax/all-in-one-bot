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
	return "/bitcointalk_stop"
}

func (h *BitcointalkStopHandle) Handle(c tb.Context) error {
	chat := c.Chat()

	h.Bitcointalk.Logger.Info(
		"command processed",
		"command", h.Cmd(),
		"chat_id", chat.ID,
		"chat_type", chat.Type, // 例如：private, group
	)

	err := h.Bitcointalk.StopMonitor()
	if err != nil {
		return c.Send(fmt.Sprintf("%v", err))
	}

	return c.Send("Bitcointalk 监控已停止")
}
