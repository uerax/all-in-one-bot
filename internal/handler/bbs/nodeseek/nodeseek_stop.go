package nodeseek

import (
	"fmt"

	tb "gopkg.in/telebot.v4"
)

type NodeseekStopHandle struct {
	*Nodeseek
}

func NewNodeseekStopHandle(n *Nodeseek) *NodeseekStopHandle {
	return &NodeseekStopHandle{
		Nodeseek: n,
	}
}

func (h *NodeseekStopHandle) Cmd() string {
	return "/nodeseek_stop"
}

func (h *NodeseekStopHandle) Handle(c tb.Context) error {
	chat := c.Chat()
	h.Nodeseek.Log.Info(
		"command processed",
		"command", h.Cmd(),
		"chat_id", chat.ID,
		"chat_type", chat.Type, // 例如：private, group
	)
	err := h.Nodeseek.StopMonitor()
	if err != nil {
		return c.Send(fmt.Sprintf("%v", err))
	}
	return c.Send("Nodeseek 监控已关闭")
}