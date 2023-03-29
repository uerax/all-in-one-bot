package tg

import (
	"log"
	"strings"
	"tg-aio-bot/crypto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var api = &Aio{}

type Aio struct {
	bot       *tgbotapi.BotAPI
	CryptoApi *crypto.Monitor
}

func (t *Aio) NewBot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	t.bot = bot

	go t.WaitToSend()

	t.CryptoApi = crypto.NewCryptoMonitor()
	t.CryptoApi.Do()
}

func (t *Aio) WaitToSend() {
	for {
		select {
		case v := <-t.CryptoApi.C:
			for id, cryptoToPrice := range v {
				if len(cryptoToPrice) == 0 {
					continue
				}
				sb := strings.Builder{}
				sb.WriteString("当前监控的加密货币超过监控线的有:")
				for crypto, price := range cryptoToPrice {
					sb.WriteString("\n")
					sb.WriteString(crypto)
					sb.WriteString(" : ")
					sb.WriteString(price)
				}
				t.bot.Send(tgbotapi.NewMessage(id, sb.String()))
			}
		}
	}
}
