package handler

import (
	"fmt"

	tb "gopkg.in/telebot.v4"

	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)


type ChatIDHandle struct {
	Logger logger.Log
}

func (h *ChatIDHandle) Cmd() string {
	return "/chatid"
}

func (h *ChatIDHandle) Handle(c tb.Context) error {
	chat := c.Chat()

	h.Logger.Info(
        "command processed",
		"command", h.Cmd(),
        "chat_id", chat.ID, 
        "chat_type", chat.Type, // 例如：private, group
    )

	err := c.Send(fmt.Sprintf("This chat ID is: %d", chat.ID))

	if err != nil {
        h.Logger.Error(
            "Failed to send response",
            "chat_id", chat.ID,
            "error", err,
        )
    }

	return err
}
