package telegram

import (
	"errors"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	tb "gopkg.in/telebot.v4"
)

func NewBot(cfg config.Telegram) (*tb.Bot, error) {

	// 1. 确保 Token 存在
	if cfg.Token == "" {
		return nil, errors.New("telegram bot token cannot be empty")
	}

	settings := tb.Settings{
		Token:  cfg.Token,
		Poller: &tb.LongPoller{Timeout: time.Duration(cfg.Timeout) * time.Second}, // 👈 从配置中读取 Timeout
	}

	return tb.NewBot(settings)
}
