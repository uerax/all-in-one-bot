package router

import (
	"github.com/uerax/all-in-one-bot/lite/internal/handler/bbs/bitcointalk"
	"github.com/uerax/all-in-one-bot/lite/internal/handler/telegram"

	tb "gopkg.in/telebot.v4"
)

type Handler interface {
	Cmd() string
	Handle(c tb.Context) error
}


func Handlers(deps *Dependencies) []Handler {
	
	var handlers []Handler

	// telegram handlers
	// chatid
	handlers = append(handlers, telegram.NewChatIDHandle(deps.Logger))
	
	// bitcointalk handlers
	bitcointalkService := bitcointalk.NewBitcointalkHandle(deps.Config, deps.Logger)
	// bitcointalk_start	
	handlers = append(handlers, bitcointalk.NewBitcointalkStartHandle(bitcointalkService))
	// bitcointalk_stop	
	handlers = append(handlers, bitcointalk.NewBitcointalkStopHandle(bitcointalkService))
	

	return handlers
}

// RegisterHandlers 负责将所有 Handler 绑定到 Bot 实例。
func RegisterHandlers(b *tb.Bot, deps *Dependencies) {

// 1. 调用 Handlers 获取所有已经**配置好并注入了依赖**的 Handler 实例。
    //    Logger 实例必须在这里作为参数传入 Handlers。
    handlers := Handlers(deps)

    for _, h := range handlers {
        b.Handle(h.Cmd(), h.Handle)
    }
}