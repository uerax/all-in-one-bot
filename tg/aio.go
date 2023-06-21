package tg

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/uerax/all-in-one-bot/chatgpt"
	"github.com/uerax/all-in-one-bot/cron"
	"github.com/uerax/all-in-one-bot/crypto"
	"github.com/uerax/all-in-one-bot/lists"
	"github.com/uerax/all-in-one-bot/photo"
	"github.com/uerax/all-in-one-bot/utils"
	"github.com/uerax/all-in-one-bot/video"
	"github.com/uerax/all-in-one-bot/vps"
)

var api = &Aio{}

type Aio struct {
	local       string
	bot         *tgbotapi.BotAPI
	CryptoApi   *crypto.CryptoMonitor
	ChatGPTApi  *chatgpt.ChatGPT
	VpsApi      *vps.VpsMonitor
	PhotoApi    *photo.Cutouts
	CryptoV2Api *crypto.Probe
	Cron        *cron.Task
	Video       *video.VideoDownload
	Gif         *Gif
	Sticker     *Sticker
	Utils       *utils.Utils
	Lists       *lists.Lists
	Track 		*crypto.Track
}

func (t *Aio) NewBot(token string, local string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	t.bot = bot
	t.local = local

	t.CryptoApi = crypto.NewCryptoMonitor()
	t.ChatGPTApi = chatgpt.NewChatGPT()
	t.VpsApi = vps.NewVpsMonitor()
	t.PhotoApi = photo.NewCutouts()
	t.CryptoV2Api = crypto.NewProbe()
	t.Cron = cron.NewTask()
	t.Video = video.NewVideoDownload()
	t.Gif = NewGif()
	t.Sticker = NewSticker()
	t.Utils = utils.NewUtils()
	t.Lists = lists.NewLists()
	t.Track = crypto.NewTrack()

	go t.WaitToSend()
}

func (t *Aio) SendMsg(id int64, msg string) {
	mc := tgbotapi.NewMessage(id, msg)
	t.bot.Send(mc)
}

func (t *Aio) SendMessage(msg string) {
	mc := tgbotapi.NewMessage(ChatId, msg)
	t.bot.Send(mc)
}

func (t *Aio) SendMarkdown(id int64, msg string, preview bool) {
	mc := tgbotapi.NewMessage(id, msg)
	mc.ParseMode = "Markdown"
	mc.DisableWebPagePreview = preview
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

func (t *Aio) SendAudio(id int64, cfg []interface{}) {
	if len(cfg) != 3 {
		return
	}
	mc := tgbotapi.NewAudio(id, tgbotapi.FilePath(cfg[2].(string)))
	mc.Duration = cfg[1].(int)
	mc.Thumb = tgbotapi.FilePath(cfg[0].(string))
	t.bot.Send(mc)
}

func (t *Aio) LocalServerSendFile(id int64, filepath string, filename string) {
	cmd := exec.Command("curl",
		"-v",
		"-F", fmt.Sprintf("chat_id=%d", id),
		"-F", fmt.Sprintf("video=file://%s", filepath),
		"-F", "supports_streaming=true",
		"-F", fmt.Sprintf("caption=%s", filename),
		fmt.Sprintf("%sbot%s/sendVideo", t.local, t.bot.Token),
	)

	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	}
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
		case v := <-t.CryptoV2Api.Meme:
			go t.SendMarkdown(ChatId, v, true)
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
		// Utils
		case v := <-t.Utils.MsgC:
			go t.SendMarkdown(ChatId, v, false)
		case v := <-t.Utils.ErrC:
			go t.SendMsg(ChatId, v)
		// Lists
		case v := <-t.Lists.C:
			go t.SendMarkdown(ChatId, v, false)
		case v := <-t.Lists.ErrC:
			go t.SendMsg(ChatId, v)
		// Track
		case v := <-t.Track.C:
			go t.SendMarkdown(ChatId, v, true)
		}

	}
}
