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
		fmt.Println("receive msg : " + update.Message.Text)
		
		if !update.Message.IsCommand() { // ignore any non-command Messages
			if goconf.VarBoolOrDefault(false, "photo", "enable") && update.Message.Photo != nil {
				cutouts(update.Message.Chat.ID, update.Message.Photo)
			}
			if goconf.VarBoolOrDefault(false, "telegram", "chat") {
				execute(update.Message.Chat.ID, update.Message.Text)
			}
            continue
        }
		
		switch update.Message.Command() {
		// Crypto
		case "add_crypto_growth_monitor":
			addCryptoGrowthMonitor(update.Message.Chat.ID, update.Message.CommandArguments())
		case "add_crypto_decline_monitor":
			addCryptoDeclineMonitor(update.Message.Chat.ID, update.Message.CommandArguments())
		case "get_crypto_price":
			getCryptoPrice(update.Message.Chat.ID, update.Message.CommandArguments())
		case "delete_crypto_minitor":
			deleteCryptoMinitor(update.Message.Chat.ID, update.Message.CommandArguments())
		case "get_crypto_ufutures_price":
			getUFuturesCryptoPrice(update.Message.Chat.ID, update.Message.CommandArguments())
		// ChatGPT
		case "chatgpt":
			chatGPT(update.Message.Chat.ID, update.Message.CommandArguments())
		// Vps
		case "vps_monitor_supported_list":
			vpsMonitorSupportedList(update.Message.Chat.ID)
		case "add_vps_monitor":
			addVpsMonitor(update.Message.Chat.ID, update.Message.CommandArguments())
		}
		
	}
}
