package tg

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var api = &Aio{}

type Aio struct {
	local       string
	bot         *tgbotapi.BotAPI
}

func (t *Aio) NewBot(token string, local string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	t.bot = bot
	t.local = local

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
	mc.ParseMode = "Markdown"
	mc.DisableWebPagePreview = preview
	return t.bot.Send(mc)
}

func (t *Aio) DeleteAfterSendMarkdown(id int64, msg string, preview bool) {
	mc := tgbotapi.NewMessage(id, msg)
	mc.ParseMode = "Markdown"
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
	t.bot.Send(tgbotapi.NewEditMessageText(id, msgId, msg))
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

func (t *Aio) WaitToSend() {

}
