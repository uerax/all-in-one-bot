package bot

import (
	"context"

	"github.com/uerax/all-in-one-bot/lite/internal/models"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	tb "gopkg.in/telebot.v4"
)

type Dispatcher struct {
	bot   *tb.Bot
	msgCh <-chan models.Message
	log   logger.Log
}

func NewDispatcher(b *tb.Bot, msgCh <-chan models.Message, log logger.Log) *Dispatcher {
	return &Dispatcher{
		bot:   b,
		msgCh: msgCh,
		log:   log,
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-d.msgCh:
				var err error
				if msg.Kind == models.KindMarkdown {
					_, err = d.bot.Send(tb.ChatID(msg.ChatID), msg.Text, tb.ModeMarkdown, tb.NoPreview)
				} else {
					_, err = d.bot.Send(tb.ChatID(msg.ChatID), msg.Text)
				}
				if err != nil {
					d.log.Error("dispatcher: 发送消息失败", "chatID", msg.ChatID, "error", err)
				}
			}
		}
	}()
}
