package crocodile

import (
	"fmt"
	"strconv"
	"strings"

	tb "gopkg.in/telebot.v4"
)

// crocodileMonitorHandle 开启每日定时量能扫描
type crocodileMonitorHandle struct{ svc *Crocodile }

func NewCrocodileMonitorHandle(svc *Crocodile) *crocodileMonitorHandle {
	return &crocodileMonitorHandle{svc}
}
func (h *crocodileMonitorHandle) Cmd() string { return "/crocodile_monitor" }
func (h *crocodileMonitorHandle) Handle(c tb.Context) error {
	go h.svc.Monitor(c.Chat().ID)
	return nil
}

// crocodileStopHandle 停止监控
type crocodileStopHandle struct{ svc *Crocodile }

func NewCrocodileStopHandle(svc *Crocodile) *crocodileStopHandle {
	return &crocodileStopHandle{svc}
}
func (h *crocodileStopHandle) Cmd() string { return "/crocodile_stop" }
func (h *crocodileStopHandle) Handle(c tb.Context) error {
	go h.svc.Stop(c.Chat().ID)
	return nil
}

// crocodileCheckHandle 立即执行一次扫描
type crocodileCheckHandle struct{ svc *Crocodile }

func NewCrocodileCheckHandle(svc *Crocodile) *crocodileCheckHandle {
	return &crocodileCheckHandle{svc}
}
func (h *crocodileCheckHandle) Cmd() string { return "/crocodile_check" }
func (h *crocodileCheckHandle) Handle(c tb.Context) error {
	go h.svc.Handle(c.Chat().ID)
	return nil
}

// crocodileListHandle 查看监控列表
type crocodileListHandle struct{ svc *Crocodile }

func NewCrocodileListHandle(svc *Crocodile) *crocodileListHandle {
	return &crocodileListHandle{svc}
}
func (h *crocodileListHandle) Cmd() string { return "/crocodile_list" }
func (h *crocodileListHandle) Handle(c tb.Context) error {
	go h.svc.ListMonitor(c.Chat().ID)
	return nil
}

// crocodileAddHandle 添加监控币种（/crocodile_add <id> [name]）
type crocodileAddHandle struct{ svc *Crocodile }

func NewCrocodileAddHandle(svc *Crocodile) *crocodileAddHandle {
	return &crocodileAddHandle{svc}
}
func (h *crocodileAddHandle) Cmd() string { return "/crocodile_add" }
func (h *crocodileAddHandle) Handle(c tb.Context) error {
	args := strings.TrimSpace(c.Message().Payload)
	if args == "" {
		parts := strings.Fields(c.Message().Text)
		if len(parts) > 1 {
			args = strings.Join(parts[1:], " ")
		}
	}
	if args == "" {
		return c.Send("用法: /crocodile_add <id> [name]")
	}
	parts := strings.SplitN(args, " ", 2)
	id := parts[0]
	name := ""
	if len(parts) > 1 {
		name = parts[1]
	}
	go h.svc.AddMonitor(c.Chat().ID, id, name)
	return nil
}

// crocodileRuleHandle 两阶段：先展示当前规则，再解析新参数更新
// 第一次调用 -> 展示提示；此后 bot 在外部用 Cmd 状态机路由文本输入
type crocodileRuleHandle struct{ svc *Crocodile }

func NewCrocodileRuleHandle(svc *Crocodile) *crocodileRuleHandle {
	return &crocodileRuleHandle{svc}
}
func (h *crocodileRuleHandle) Cmd() string { return "/crocodile_rule" }
func (h *crocodileRuleHandle) Handle(c tb.Context) error {
	args := strings.TrimSpace(c.Message().Payload)
	if args == "" {
		parts := strings.Fields(c.Message().Text)
		if len(parts) > 1 {
			args = strings.Join(parts[1:], " ")
		}
	}
	// 如果直接带参数就更新，否则展示当前配置
	if args == "" {
		go h.svc.RuleTip(c.Chat().ID)
		return nil
	}
	lookback, yesterday, average, err := parseRuleArgs(args)
	if err != nil {
		return c.Send(fmt.Sprintf("参数错误: %v\n格式: /crocodile_rule <lookback> <today_multiple> <avg_multiple>", err))
	}
	go h.svc.UpdateRule(c.Chat().ID, lookback, yesterday, average)
	return nil
}

func parseRuleArgs(args string) (lookback int, yesterday, average float64, err error) {
	parts := strings.Fields(args)
	if len(parts) != 3 {
		err = fmt.Errorf("需要 3 个参数，收到 %d 个", len(parts))
		return
	}
	lookback, err = strconv.Atoi(parts[0])
	if err != nil {
		err = fmt.Errorf("lookback 非整数: %s", parts[0])
		return
	}
	yesterday, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		err = fmt.Errorf("today_multiple 非数字: %s", parts[1])
		return
	}
	average, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		err = fmt.Errorf("avg_multiple 非数字: %s", parts[2])
		return
	}
	return
}
