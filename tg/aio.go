package tg

import (
	"log"
	"strings"
	"tg-aio-bot/chatgpt"
	"tg-aio-bot/crypto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var api = &Aio{}

type Aio struct {
	bot       *tgbotapi.BotAPI
	CryptoApi *crypto.CryptoMonitor
	ChatGPTApi *chatgpt.ChatGPT
}

func (t *Aio) SendMsg(id int64, msg string) {
	t.bot.Send(tgbotapi.NewMessage(id, msg))
}

func (t *Aio) NewBot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	t.bot = bot

	t.CryptoApi = crypto.NewCryptoMonitor()
	t.ChatGPTApi = chatgpt.NewChatGPT()

	go t.WaitToSend()
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
				sb.WriteString("有加密货币触发监控线 :")
				for crypto, price := range cryptoToPrice {
					sb.WriteString("\n")
					sb.WriteString(crypto)
					sb.WriteString(" : ")
					sb.WriteString(price)
				}
				go t.bot.Send(tgbotapi.NewMessage(id, sb.String()))
			}
		case v := <-t.ChatGPTApi.C:
			for id, msg := range v {
				if len(msg) > 4096 {
					for i,j := 0,4000; j < len(msg); j = j << 1 {
						if j > len(msg) {
							j = len(msg)
						}
						go t.SendMsg(id, msg[i:j])
						i = j
					}
				} else {
					go t.SendMsg(id, msg)
				}
			}
		}
		
	}
}
