package telegram

import (
	"log/slog"
	"os"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	tb "gopkg.in/telebot.v4"
)

func NewBot(cfg config.Telegram) (*tb.Bot, error) {


	// 1. 确保 Token 存在
	if cfg.Token == "" {
		slog.Error("FATAL: Telegram Token is required")
		os.Exit(1)
	}

    settings := tb.Settings{
        Token:  cfg.Token,
        Poller: &tb.LongPoller{Timeout: time.Duration(cfg.Timeout) * time.Second}, // 👈 从配置中读取 Timeout
    }

    return tb.NewBot(settings)
}