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
	return c.Send(fmt.Sprintf("This chat ID is: %d", chat.ID))
}
