package main

import (
	"os"

	"github.com/uerax/all-in-one-bot/lite/internal/app"
)

const BotToken = "YOUR_TELEGRAM_BOT_TOKEN" 

func main() {
	app.Run(os.Getenv("TELEGRAM_TOKEN"))
}