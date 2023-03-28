package tg

import (
	"fmt"
	"log"
	"tg-aio-bot/crypto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Server() {

	// Create a new bot instance
	bot, err := tgbotapi.NewBotAPI("6296061886:AAEFywGZ9FDmDIX-0bHz75lxx_x8iLUex9o")
	if err != nil {
		log.Panic(err)
	}

	// Set bot options
	bot.Debug = true

	// Create a new update channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	s := crypto.NewCrypto("","")

	// Start listening for updates
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}

		// Process incoming messages
		switch update.Message.Text {
		case "/start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%v", s.Price("ARBUSDT")))
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		case "How are you?":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I'm doing great, thank you for asking!")
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}