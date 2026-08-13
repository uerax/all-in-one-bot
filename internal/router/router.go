package router

import (
	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
	cgprovider "github.com/uerax/all-in-one-bot/lite/internal/crypto/provider/coingecko"
	gtprovider "github.com/uerax/all-in-one-bot/lite/internal/crypto/provider/geckoterminal"
	"github.com/uerax/all-in-one-bot/lite/internal/handler/bbs/bitcointalk"
	"github.com/uerax/all-in-one-bot/lite/internal/handler/bbs/nodeseek"
	coinhandler "github.com/uerax/all-in-one-bot/lite/internal/handler/crypto/coin"
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

	// crypto providers & manager
	cgProv := cgprovider.NewProvider(deps.Config.Coingecko, deps.Logger)
	gtProv := gtprovider.NewProvider(deps.Config.GeckoTerminal.BaseURL, deps.Config.GeckoTerminal.Timeout, deps.Logger)
	providerMgr := provider.NewManager(cgProv, cgProv, cgProv, gtProv, gtProv, gtProv, gtProv)

	// coin handlers
	coinService := coinhandler.NewService(deps.Store, providerMgr, r.msgCh, deps.Logger)
	handlers = append(handlers, coinhandler.NewCoinHandle(coinService))
	handlers = append(handlers, coinhandler.NewCoinPriceHandle(coinService))
	handlers = append(handlers, coinhandler.NewCoinSearchHandle(coinService))
	handlers = append(handlers, coinhandler.NewCoinTrendingHandle(coinService))
	handlers = append(handlers, coinhandler.NewCoinPoolHandle(coinService))
	handlers = append(handlers, coinhandler.NewCoinMonitorHandle(coinService))
	handlers = append(handlers, coinhandler.NewCoinStopHandle(coinService))

	// crocodile handlers
	crService := crhandler.NewCrocodile(deps.Store, providerMgr, r.msgCh, deps.Config.Crocodile, deps.Logger)
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
