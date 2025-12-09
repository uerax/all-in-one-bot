package app

import (
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/router"

	tb "gopkg.in/telebot.v4"
)

func Run(token string) {
	log := logger.NewLogger()
		// 1. 确保 Token 存在
	if token == "" {
		log.Error("FATAL: Bot Token not set. Please update the token constant.")
	}

	// 2. Telebot 初始化设置
	settings := tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second}, // 使用 LongPoller 或 Webhook
	}

	b, err := tb.NewBot(settings)
	if err != nil {
		log.Error("FATAL: Could not create bot:", err)
	}

	// 3. 确保依赖注入
	dependencies := &router.Dependencies{}
	dependencies.Logger = log

	// 3. 核心步骤：注册所有 Handler
	// 由于我们导入了 handlers 包，所有 init() 函数都已经运行，GlobalHandlers 已填充完毕。
	router.RegisterHandlers(b, dependencies)

	// 4. 启动 Bot
	log.Info("🚀 Bot @%s is starting up...", b.Me.Username)
	b.Start() // Bot 会阻塞在这里，直到程序停止
}