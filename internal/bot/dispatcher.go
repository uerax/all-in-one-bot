package bot

import (
	"context"

	"github.com/uerax/all-in-one-bot/lite/internal/models"
	tb "gopkg.in/telebot.v4"
)

type Dispatcher struct {
	bot   *tb.Bot
	msgCh <-chan models.Message
}

func NewDispatcher(b *tb.Bot, msgCh <-chan models.Message) *Dispatcher {
	return &Dispatcher{
		bot:   b,
		msgCh: msgCh,
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-d.msgCh:
				d.bot.Send(tb.ChatID(msg.ChatID), msg.Text)
			}
		}
	}()
}
