package tg

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/uerax/goconf"
)

var Cmd string
var ChatId int64

func Server() {
	// Create a new bot instance
	token, err := goconf.VarString("telegram", "token")
	if err != nil {
		log.Println("启动失败: 没有填写bot的token")
		return
	}
	id := goconf.VarIntOrDefault(0, "telegram", "chatId")

	ChatId = int64(id)

	api.NewBot(token, goconf.VarStringOrDefault("http://localhost:8081/", "telegram", "local"))

	// Create a new update channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	// Start listening for updates
	updates := api.bot.GetUpdatesChan(u)
	notice("System Notification: Aio 服务已启动")
	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}
		log.Println("receive msg : " + update.Message.Text)

		if ChatId != 0 && update.Message.Chat.ID != ChatId {
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			
		}

		// Cmd Tip
		
	}
}
