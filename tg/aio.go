package tg

import (
	"log"
	"strings"

	"tg-aio-bot/chatgpt"
	"tg-aio-bot/cron"
	"tg-aio-bot/crypto"
	"tg-aio-bot/photo"
	"tg-aio-bot/video"
	"tg-aio-bot/vps"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var api = &Aio{}

type Aio struct {
	bot        *tgbotapi.BotAPI
	CryptoApi  *crypto.CryptoMonitor
	ChatGPTApi *chatgpt.ChatGPT
	VpsApi     *vps.VpsMonitor
	PhotoApi   *photo.Cutouts
	CryptoV2Api *crypto.Probe
	Cron *cron.Task
	Video *video.VideoDownload
	Gif *Gif
	Sticker *Sticker
}

func (t *Aio) SendMsg(id int64, msg string) {
	mc := tgbotapi.NewMessage(id, msg)
	t.bot.Send(mc)
}

func (t *Aio) SendImg(id int64, img string) {
	mc := tgbotapi.NewPhoto(id, tgbotapi.FilePath(img))
	t.bot.Send(mc)
}

func (t *Aio) SendVideo(id int64, video string) {
	mc := tgbotapi.NewVideo(id, tgbotapi.FilePath(video))
	t.bot.Send(mc)
}

func (t *Aio) SendFile(id int64, file string) {
	mc := tgbotapi.NewDocument(id, tgbotapi.FilePath(file))
	t.bot.Send(mc)
}

func (t *Aio) SendAudio(id int64, file string) {
	mc := tgbotapi.NewAudio(id, tgbotapi.FilePath(file))
	
	t.bot.Send(mc)
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
	t.VpsApi = vps.NewVpsMonitor()
	t.PhotoApi = photo.NewCutouts()
	t.CryptoV2Api = crypto.NewProbe()
	t.Cron = cron.NewTask()
	t.Video = video.NewVideoDownload()
	t.Gif = NewGif()
	t.Sticker = NewSticker()

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
					for i, j := 0, 4000; j < len(msg); j = j << 1 {
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

		case v := <-t.VpsApi.C:
			for k, v := range v {
				go t.SendMsg(k, v)
			}

		case v := <-t.PhotoApi.C:
			for k, v := range v {
				go t.SendImg(k, v)
			}
		case v := <-t.CryptoV2Api.Kline:
			go t.SendMsg(ChatId, v)
		case v := <-t.Cron.C:
			go t.SendMsg(ChatId, v)
		// Youtube
		case v := <-t.Video.C:
			go t.SendVideo(ChatId, v)
		case v := <-t.Video.AudioC:
			go t.SendAudio(ChatId, v)
		case v := <-t.Video.MsgC:
			go t.SendMsg(ChatId, v)
		// GIF
		case v := <-t.Gif.C:
			go t.SendFile(ChatId, v)
		case v := <-t.Gif.MsgC:
			go t.SendMsg(ChatId, v)
		// Sticker
		case v := <-t.Sticker.C:
			go t.SendFile(ChatId, v)
		case v := <-t.Sticker.MsgC:
			go t.SendMsg(ChatId, v)
		}

	}
}
