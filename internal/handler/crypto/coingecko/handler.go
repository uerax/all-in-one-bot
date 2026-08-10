package coingecko

import (
	"strings"

	tb "gopkg.in/telebot.v4"
)

// coinMonitorHandle 开启 24h 定时价格播报
type coinMonitorHandle struct{ svc *Coingecko }

func NewCoinMonitorHandle(svc *Coingecko) *coinMonitorHandle { return &coinMonitorHandle{svc} }
func (h *coinMonitorHandle) Cmd() string                     { return "/coin_monitor" }
func (h *coinMonitorHandle) Handle(c tb.Context) error {
	go h.svc.Monitor(c.Chat().ID)
	return nil
}

// coinStopHandle 停止 24h 定时监控
type coinStopHandle struct{ svc *Coingecko }

func NewCoinStopHandle(svc *Coingecko) *coinStopHandle { return &coinStopHandle{svc} }
func (h *coinStopHandle) Cmd() string                  { return "/coin_stop" }
func (h *coinStopHandle) Handle(c tb.Context) error {
	go h.svc.Stop(c.Chat().ID)
	return nil
}

// coinPriceHandle 立即查询价格
type coinPriceHandle struct{ svc *Coingecko }

func NewCoinPriceHandle(svc *Coingecko) *coinPriceHandle { return &coinPriceHandle{svc} }
func (h *coinPriceHandle) Cmd() string                   { return "/coin_price" }
func (h *coinPriceHandle) Handle(c tb.Context) error {
	go h.svc.Handle(c.Chat().ID)
	return nil
}

// coinSearchHandle 搜索币种（从 payload 或 text 获取关键字）
type coinSearchHandle struct{ svc *Coingecko }

func NewCoinSearchHandle(svc *Coingecko) *coinSearchHandle { return &coinSearchHandle{svc} }
func (h *coinSearchHandle) Cmd() string                    { return "/coin_search" }
func (h *coinSearchHandle) Handle(c tb.Context) error {
	query := strings.TrimSpace(c.Message().Payload)
	if query == "" {
		parts := strings.Fields(c.Message().Text)
		if len(parts) > 1 {
			query = strings.Join(parts[1:], " ")
		}
	}
	if query == "" {
		return c.Send("用法: /coin_search <币种名称>")
	}
	go h.svc.Search(c.Chat().ID, query)
	return nil
}
