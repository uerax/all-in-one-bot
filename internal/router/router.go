package router

import (
	"github.com/uerax/all-in-one-bot/lite/internal/handler/bbs/bitcointalk"
	"github.com/uerax/all-in-one-bot/lite/internal/handler/bbs/nodeseek"
	cghandler "github.com/uerax/all-in-one-bot/lite/internal/handler/crypto/coingecko"
	crhandler "github.com/uerax/all-in-one-bot/lite/internal/handler/crypto/crocodile"
	"github.com/uerax/all-in-one-bot/lite/internal/handler/polymarket"
	"github.com/uerax/all-in-one-bot/lite/internal/handler/telegram"
	"github.com/uerax/all-in-one-bot/lite/internal/models"

	tb "gopkg.in/telebot.v4"
)

type Handler interface {
	Cmd() string
	Handle(c tb.Context) error
}

type Router struct {
	bot   *tb.Bot
	msgCh chan models.Message
}

func (r *Router) Handlers(deps *Dependencies) []Handler {

	var handlers []Handler

	// telegram handlers
	handlers = append(handlers, telegram.NewChatIDHandle(deps.Logger))

	// bitcointalk handlers
	bitcointalkService := bitcointalk.NewBitcointalkHandle(deps.Store, &deps.Config.Bitcointalk, deps.Logger, r.msgCh)
	handlers = append(handlers, bitcointalk.NewBitcointalkStartHandle(bitcointalkService))
	handlers = append(handlers, bitcointalk.NewBitcointalkStopHandle(bitcointalkService))

	// nodeseek
	nodeseekService := nodeseek.NewNodeseek(deps.Store, r.msgCh, deps.Config.Nodeseek, deps.Logger)
	handlers = append(handlers, nodeseek.NewNodeseekStartHandle(nodeseekService))
	handlers = append(handlers, nodeseek.NewNodeseekStopHandle(nodeseekService))

	// polymarket
	polymarketService := polymarket.NewService(deps.Config.Polymarket, deps.Logger)
	handlers = append(handlers, polymarket.NewPolymarketHoldingsHandle(polymarketService, deps.Logger))
	handlers = append(handlers, polymarket.NewPolymarketL2CheckHandle(polymarketService, deps.Logger))

	// coingecko
	cgService := cghandler.NewCoingecko(deps.Store, deps.Config.Coingecko, r.msgCh, deps.Logger)
	handlers = append(handlers, cghandler.NewCoinMonitorHandle(cgService))
	handlers = append(handlers, cghandler.NewCoinStopHandle(cgService))
	handlers = append(handlers, cghandler.NewCoinPriceHandle(cgService))
	handlers = append(handlers, cghandler.NewCoinSearchHandle(cgService))

	// crocodile
	crService := crhandler.NewCrocodile(deps.Store, cgService, r.msgCh, deps.Config.Crocodile, deps.Logger)
	handlers = append(handlers, crhandler.NewCrocodileMonitorHandle(crService))
	handlers = append(handlers, crhandler.NewCrocodileStopHandle(crService))
	handlers = append(handlers, crhandler.NewCrocodileCheckHandle(crService))
	handlers = append(handlers, crhandler.NewCrocodileListHandle(crService))
	handlers = append(handlers, crhandler.NewCrocodileAddHandle(crService))
	handlers = append(handlers, crhandler.NewCrocodileRuleHandle(crService))

	return handlers
}

func NewRouter(b *tb.Bot, c chan models.Message) *Router {
	return &Router{
		bot:   b,
		msgCh: c,
	}
}

// RegisterHandlers 负责将所有 Handler 绑定到 Bot 实例。
// 每个 handler 统一包裹一层 authorizedOnly 鉴权中间件，非管理员命令被静默丢弃。
func (r *Router) RegisterHandlers(b *tb.Bot, deps *Dependencies) {
	handlers := r.Handlers(deps)
	for _, h := range handlers {
		mw := authorizedOnly(deps.AdminIDs, h.Cmd(), deps.Logger, h.Handle)
		b.Handle(h.Cmd(), mw)
	}
}
