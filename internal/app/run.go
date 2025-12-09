package app

import (
	_ "github.com/joho/godotenv/autoload"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/router"
	"github.com/uerax/all-in-one-bot/lite/internal/telegram"
)

func Run() {

	log := logger.NewLogger()

	cfg := config.LoadConfig()

	b, err := telegram.NewBot(cfg.Telegram)
	if err != nil {
		log.Error("FATAL: Could not create bot:", err)
	}

	// 3. 确保依赖注入
	dependencies := &router.Dependencies{}
	dependencies.Logger = log
	dependencies.Config = cfg

	// 3. 核心步骤：注册所有 Handler
	// 由于我们导入了 handlers 包，所有 init() 函数都已经运行，GlobalHandlers 已填充完毕。
	router.RegisterHandlers(b, dependencies)

	// 4. 启动 Bot
	log.Info("🚀 Bot @%s is starting up...", b.Me.Username)
	b.Start() // Bot 会阻塞在这里，直到程序停止
}