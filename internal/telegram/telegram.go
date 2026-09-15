package telegram

import (
	"errors"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	tb "gopkg.in/telebot.v4"
)

func NewBot(cfg config.Telegram, log logger.Log) (*tb.Bot, error) {

	// 1. 确保 Token 存在
	if cfg.Token == "" {
		return nil, errors.New("telegram bot token cannot be empty")
	}

	settings := tb.Settings{
		Token:  cfg.Token,
		Poller: &tb.LongPoller{Timeout: time.Duration(cfg.Timeout) * time.Second},
		OnError: func(err error, c tb.Context) {
			if c != nil {
				log.Error("telegram context error", "error", err, "update_id", c.Update().ID)
			} else {
				log.Error("telegram poller error", "error", err)
			}
		},
	}

	return tb.NewBot(settings)
}
