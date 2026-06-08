package tg

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/uerax/all-in-one-bot/bbs"
	"github.com/uerax/all-in-one-bot/chatgpt"
	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/all-in-one-bot/cron"
	"github.com/uerax/all-in-one-bot/crypto"
	"github.com/uerax/all-in-one-bot/crypto/crocodile"
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
	ch          chan common.AioEvent
	CryptoApi   *crypto.CryptoMonitor
	Coingecko   *crypto.Coingecko
	Crocodile   *crocodile.Crocodile
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
	Track       *crypto.Track
	Bbs         *bbs.Bbs
}

func (t *Aio) NewBot(token string, local string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	t.bot = bot
	t.local = local
	t.ch = make(chan common.AioEvent, 20)

	t.CryptoApi = crypto.NewCryptoMonitor(t.ch)
	t.Coingecko = crypto.NewCoingecko(t.ch)
	t.Crocodile = crocodile.NewCrocodile(t.ch)
	t.ChatGPTApi = chatgpt.NewChatGPT(t.ch)
	t.VpsApi = vps.NewVpsMonitor(t.ch)
	t.PhotoApi = photo.NewCutouts(t.ch)
	t.CryptoV2Api = crypto.NewProbe(t.ch)
	t.Cron = cron.NewTask(t.ch)
	t.Video = video.NewVideoDownload(t.ch)
	t.Gif = NewGif(t.ch)
	t.Sticker = NewSticker(t.ch)
	t.Utils = utils.NewUtils(t.ch)
	t.Lists = lists.NewLists(t.ch)
	t.Track = crypto.NewTrack(t.ch)
	t.Bbs = bbs.NewBbs(t.ch)

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

func (t *Aio) DeleteAfterSendMessage(msg string) {
	mc := tgbotapi.NewMessage(ChatId, msg)
	m, err := t.bot.Send(mc)
	if err == nil {
		go t.deleteAfterMinute(ChatId, m.MessageID, 2)
	}
}

func (t *Aio) SendMarkdown(id int64, msg string, preview bool) (tgbotapi.Message, error) {
	mc := tgbotapi.NewMessage(id, msg)
	mc.ParseMode = tgbotapi.ModeMarkdown
	mc.DisableWebPagePreview = preview
	return t.bot.Send(mc)
}

func (t *Aio) DeleteAfterSendMarkdown(id int64, msg string, preview bool) {
	mc := tgbotapi.NewMessage(id, msg)
	mc.ParseMode = tgbotapi.ModeMarkdown
	mc.DisableWebPagePreview = preview
	m, err := t.bot.Send(mc)
	if err == nil {
		go t.deleteAfterMinute(id, m.MessageID, 2)
	}
}

func (t *Aio) deleteAfterMinute(id int64, msgId int, minute int) {
	time.Sleep(time.Minute * time.Duration(minute))
	t.bot.Send(tgbotapi.NewDeleteMessage(id, msgId))
}

func (t *Aio) AppendMsg(id int64, msgId int, msg string) {
	mc := tgbotapi.NewEditMessageText(id, msgId, msg)
	mc.ParseMode = "Markdown"
	mc.DisableWebPagePreview = true
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

func (t *Aio) SendAudio(id int64, audio common.AioAudio) {
	mc := tgbotapi.NewAudio(id, tgbotapi.FilePath(audio.Path))
	mc.Duration = audio.Duration
	mc.Thumb = tgbotapi.FilePath(audio.Thumb)
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
		log.Println("Error:", err)
	}
}

func (t *Aio) WaitToSend() {
	for event := range t.ch {
		t.dispatch(event)
	}
}

func (t *Aio) dispatch(event common.AioEvent) {
	id := event.ChatID
	if id == 0 {
		id = ChatId
	}

	switch event.Kind {
	case common.EventText:
		go t.SendMsg(id, event.Text)
	case common.EventMarkdown:
		go t.SendMarkdown(id, event.Text, event.DisableWebPreview)
	case common.EventDeleteMarkdown:
		minutes := event.DeleteAfterMinutes
		if minutes <= 0 {
			minutes = 2
		}
		mc := tgbotapi.NewMessage(id, event.Text)
		mc.ParseMode = tgbotapi.ModeMarkdown
		mc.DisableWebPagePreview = event.DisableWebPreview
		m, err := t.bot.Send(mc)
		if err == nil {
			go t.deleteAfterMinute(id, m.MessageID, minutes)
		}
	case common.EventPhoto:
		go t.SendImg(id, event.Path)
	case common.EventVideo:
		go t.SendVideo(id, event.Path)
	case common.EventDocument:
		go t.SendFile(id, event.Path)
	case common.EventAudio:
		go t.SendAudio(id, event.Audio)
	}
}
