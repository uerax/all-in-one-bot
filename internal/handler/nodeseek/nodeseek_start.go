package nodeseek

import (
	"context"
	"fmt"

	tb "gopkg.in/telebot.v4"
)

type NodeseekStartHandle struct {
	*Nodeseek
}

func NewNodeseekStartHandle(n *Nodeseek) *NodeseekStartHandle {
	return &NodeseekStartHandle{
		Nodeseek: n,
	}
}

func (h *NodeseekStartHandle) Cmd() string {
	return "/nodeseek_start"
}

func (h *NodeseekStartHandle) Handle(c tb.Context) error {
	chat := c.Chat()
	h.Nodeseek.Log.Info(
		"command processed",
		"command", h.Cmd(),
		"chat_id", chat.ID,
		"chat_type", chat.Type, // 例如：private, group
	)

	err := h.Nodeseek.StartMonitor(context.Background(), chat.ID)

	if err != nil {
		return c.Send(fmt.Sprintf("%v", err))
	}

	return c.Send("Nodeseek 监控已启动")
}