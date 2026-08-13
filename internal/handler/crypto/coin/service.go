package coin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/store"
)

type Service struct {
	mu      sync.Mutex
	mgr     *provider.Manager
	db      store.Store
	ch      chan<- models.Message
	list    map[string]float64 // coin_id or contract -> holding count
	ctx     context.Context
	cancel  context.CancelFunc
	log     logger.Log
}

func NewService(db store.Store, mgr *provider.Manager, ch chan<- models.Message, log logger.Log) *Service {
	s := &Service{
		db:  db,
		mgr: mgr,
		ch:  ch,
		log: log,
	}
	s.syncList()
	return s
}

// QueryCoin handles /coin <query> for single asset/contract/pool lookup.
func (s *Service) QueryCoin(chatID int64, query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		s.ch <- models.Message{ChatID: chatID, Text: "用法: /coin <代币名称 | 合约地址 | 链:地址>"}
		return
	}

	data, err := s.mgr.GetPrice(query)
	if err != nil {
		s.ch <- models.Message{ChatID: chatID, Text: fmt.Sprintf("查询失败: %v", err)}
		return
	}

	s.ch <- models.Message{
		ChatID: chatID,
		Text:   FormatMarketData(data),
		Kind:   models.KindMarkdown,
	}
}

// Handle handles /coin_price portfolio holdings calculation and reporting.
func (s *Service) Handle(chatID int64) {
	s.syncList()

	s.mu.Lock()
	list := s.list
	s.mu.Unlock()

	if len(list) == 0 {
		s.ch <- models.Message{ChatID: chatID, Text: "持仓列表为空，请在数据库 coingecko/list 中配置持仓数据"}
		return
	}

	var total float64
	var lines []string

	for key, count := range list {
		data, err := s.mgr.GetPrice(key)
		if err != nil || data == nil {
			s.log.Error("coin portfolio fetch failed", "key", key, "error", err)
			continue
		}

		itemTotal := count * data.PriceUSD
		total += itemTotal
		symbol := data.Symbol
		if symbol == "" {
			symbol = key
		}

		lines = append(lines, fmt.Sprintf("%s: *$%s* 持仓: *$%.2f* 24h: *%s*",
			symbol, formatPrice(data.PriceUSD), itemTotal, formatPercent(data.Change24h)))
	}

	header := fmt.Sprintf("当前总持有价值为 *$%.2f*", total)
	msgText := header + "\n" + strings.Join(lines, "\n")

	s.ch <- models.Message{
		ChatID: chatID,
		Text:   msgText,
		Kind:   models.KindMarkdown,
	}
}

// Search handles /coin_search <query> searching CEX and DEX markets.
func (s *Service) Search(chatID int64, query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		s.ch <- models.Message{ChatID: chatID, Text: "用法: /coin_search <币种名称 | 合约地址>"}
		return
	}

	res, err := s.mgr.Search(query)
	if err != nil {
		s.ch <- models.Message{ChatID: chatID, Text: "搜索失败: " + err.Error()}
		return
	}

	s.ch <- models.Message{
		ChatID: chatID,
		Text:   FormatMultiSearchResult(res),
		Kind:   models.KindMarkdown,
	}
}

// Trending handles /coin_trending [network].
func (s *Service) Trending(chatID int64, network string) {
	network = strings.TrimSpace(network)

	list, err := s.mgr.GetTrending(network)
	if err != nil {
		s.ch <- models.Message{ChatID: chatID, Text: "获取 DEX 热门池子失败: " + err.Error()}
		return
	}

	s.ch <- models.Message{
		ChatID: chatID,
		Text:   FormatTrending(network, list),
		Kind:   models.KindMarkdown,
	}
}

// PoolDetail handles /coin_pool <network:address> or <network> <pool_address>.
func (s *Service) PoolDetail(chatID int64, network, poolAddress string) {
	query := poolAddress
	if network != "" && poolAddress != "" {
		query = fmt.Sprintf("%s:%s", network, poolAddress)
	} else if network != "" {
		query = network
	}

	s.QueryCoin(chatID, query)
}

// Monitor starts 24h ticker loop for portfolio holdings reporting.
func (s *Service) Monitor(chatID int64) {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		s.ch <- models.Message{ChatID: chatID, Text: "Coin 监控已在运行中"}
		return
	}
	ctx, cf := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cf
	s.mu.Unlock()

	s.syncList()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		s.log.Info("Coin 24h 定时监控已启动")
		for {
			select {
			case <-ctx.Done():
				s.log.Info("Coin 定时监控已停止")
				return
			case <-ticker.C:
				s.Handle(chatID)
			}
		}
	}()
}

// Stop stops portfolio 24h monitor.
func (s *Service) Stop(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel == nil {
		s.ch <- models.Message{ChatID: chatID, Text: "Coin 监控未运行"}
		return
	}
	s.cancel()
	s.cancel = nil
	s.ch <- models.Message{ChatID: chatID, Text: "Coin 定时监控已关闭"}
}

func (s *Service) syncList() {
	if s.db == nil {
		return
	}
	var list map[string]float64
	if err := s.db.Load("coingecko", "list", &list); err != nil {
		s.log.Error("coin 从 Store 加载持仓列表失败", "error", err)
		return
	}
	s.mu.Lock()
	s.list = list
	s.mu.Unlock()
}
