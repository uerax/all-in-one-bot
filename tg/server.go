package tg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/uerax/goconf"
)

func Server() {

	// Create a new bot instance
	token, err := goconf.VarString("telegram", "token")
	if err != nil {
		panic(err)
	}

	api.NewBot(token)

	// Create a new update channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	
	// Start listening for updates
	updates := api.bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
            continue
        }
		fmt.Println("receive msg : " + update.Message.Text)
		switch update.Message.Command() {
		case "addCryptoGrowthMonitor":
			addCryptoGrowthMonitor(update.Message.Chat.ID, update.Message.CommandArguments())
		case "addCryptoDeclineMonitor":
			addCryptoDeclineMonitor(update.Message.Chat.ID, update.Message.CommandArguments())
		case "getCryptoPrice":
			getCryptoPrice(update.Message.Chat.ID, update.Message.CommandArguments())
		}
		
	}
}
