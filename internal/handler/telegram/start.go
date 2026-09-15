package telegram

import (
	tb "gopkg.in/telebot.v4"

	"github.com/uerax/all-in-one-bot/lite/internal/models"
)

type StartHandle struct {
	msgCh chan<- models.Message
}

func NewStartHandle(msgCh chan<- models.Message) *StartHandle {
	return &StartHandle{
		msgCh: msgCh,
	}
}

func (h *StartHandle) Cmd() string {
	return "/start"
}

func (h *StartHandle) Handle(c tb.Context) error {
	welcomeMsg := `🤖 *all-in-one-bot (lite)*

欢迎使用多功能 Telegram 机器人！

*常用命令*：
• /coin <代码|合约|链:地址> - 查询代币行情与 DEX 池子
• /coin_price - 持仓实时估值
• /coin_search <关键词> - 混合检索 CEX 与 DEX
• /coin_trending - DEX 热门交易池
• /crocodile_check - 鳄鱼量能异动即时诊断
• /crocodile_monitor - 开启每日 UTC 00:01 自动量能监控
• /crocodile_list - 查看量能监控列表
• /chatid - 查询当前会话 Chat ID`

	h.msgCh <- models.Message{
		ChatID: c.Chat().ID,
		Text:   welcomeMsg,
		Kind:   models.KindMarkdown,
	}
	return nil
}
