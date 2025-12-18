package app

import (
	"context"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"github.com/uerax/all-in-one-bot/lite/internal/bot"
	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/router"
	"github.com/uerax/all-in-one-bot/lite/internal/telegram"
)

func Run() {

	log := logger.NewLogger()

	cfg := config.LoadConfig()

	b, err := telegram.NewBot(cfg.Telegram)
	if err != nil {
		log.Error(
			"FATAL: Application initialization failed",
			"error", err,
		)
		os.Exit(1)
	}

	// 依赖注入
	dependencies := &router.Dependencies{}
	dependencies.Logger = log
	dependencies.Config = cfg

	// 创建 Dispatcher 用于消息分发
	c := make(chan models.Message, 20)
	dispatcher := bot.NewDispatcher(b, c)
	dispatcher.Start(context.Background())


	// 注册所有 Handler
	router := router.NewRouter(b, c)
	router.RegisterHandlers(b, dependencies)

	// 启动 Bot
	log.Info("🚀 Bot is starting up...")
	b.Start() // Bot 会阻塞在这里，直到程序停止
}
