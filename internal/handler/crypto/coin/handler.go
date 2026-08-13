package coin

import (
	"strings"

	tb "gopkg.in/telebot.v4"
)

type coinHandle struct {
	svc *Service
}

func NewCoinHandle(svc *Service) *coinHandle {
	return &coinHandle{svc: svc}
}

func (h *coinHandle) Cmd() string {
	return "/coin"
}

func (h *coinHandle) Handle(c tb.Context) error {
	payload := strings.TrimSpace(c.Message().Payload)
	if payload == "" {
		payload = strings.TrimSpace(c.Message().Text)
		parts := strings.Fields(payload)
		if len(parts) > 1 {
			payload = strings.Join(parts[1:], " ")
		} else {
			payload = ""
		}
	}
	go h.svc.QueryCoin(c.Chat().ID, payload)
	return nil
}

type coinPriceHandle struct {
	svc *Service
}

func NewCoinPriceHandle(svc *Service) *coinPriceHandle {
	return &coinPriceHandle{svc: svc}
}

func (h *coinPriceHandle) Cmd() string {
	return "/coin_price"
}

func (h *coinPriceHandle) Handle(c tb.Context) error {
	go h.svc.Handle(c.Chat().ID)
	return nil
}

type coinSearchHandle struct {
	svc *Service
}

func NewCoinSearchHandle(svc *Service) *coinSearchHandle {
	return &coinSearchHandle{svc: svc}
}

func (h *coinSearchHandle) Cmd() string {
	return "/coin_search"
}

func (h *coinSearchHandle) Handle(c tb.Context) error {
	payload := strings.TrimSpace(c.Message().Payload)
	if payload == "" {
		payload = strings.TrimSpace(c.Message().Text)
		parts := strings.Fields(payload)
		if len(parts) > 1 {
			payload = strings.Join(parts[1:], " ")
		} else {
			payload = ""
		}
	}
	go h.svc.Search(c.Chat().ID, payload)
	return nil
}

type coinTrendingHandle struct {
	svc *Service
}

func NewCoinTrendingHandle(svc *Service) *coinTrendingHandle {
	return &coinTrendingHandle{svc: svc}
}

func (h *coinTrendingHandle) Cmd() string {
	return "/coin_trending"
}

func (h *coinTrendingHandle) Handle(c tb.Context) error {
	network := strings.TrimSpace(c.Message().Payload)
	go h.svc.Trending(c.Chat().ID, network)
	return nil
}

type coinPoolHandle struct {
	svc *Service
}

func NewCoinPoolHandle(svc *Service) *coinPoolHandle {
	return &coinPoolHandle{svc: svc}
}

func (h *coinPoolHandle) Cmd() string {
	return "/coin_pool"
}

func (h *coinPoolHandle) Handle(c tb.Context) error {
	payload := strings.TrimSpace(c.Message().Payload)
	parts := strings.Fields(payload)
	network := ""
	poolAddress := ""
	if len(parts) >= 2 {
		network = parts[0]
		poolAddress = parts[1]
	} else if len(parts) == 1 {
		poolAddress = parts[0]
	}
	go h.svc.PoolDetail(c.Chat().ID, network, poolAddress)
	return nil
}

type coinMonitorHandle struct {
	svc *Service
}

func NewCoinMonitorHandle(svc *Service) *coinMonitorHandle {
	return &coinMonitorHandle{svc: svc}
}

func (h *coinMonitorHandle) Cmd() string {
	return "/coin_monitor"
}

func (h *coinMonitorHandle) Handle(c tb.Context) error {
	go h.svc.Monitor(c.Chat().ID)
	return nil
}

type coinStopHandle struct {
	svc *Service
}

func NewCoinStopHandle(svc *Service) *coinStopHandle {
	return &coinStopHandle{svc: svc}
}

func (h *coinStopHandle) Cmd() string {
	return "/coin_stop"
}

func (h *coinStopHandle) Handle(c tb.Context) error {
	go h.svc.Stop(c.Chat().ID)
	return nil
}
