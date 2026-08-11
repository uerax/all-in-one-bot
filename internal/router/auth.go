package router

import (
	tb "gopkg.in/telebot.v4"

	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

// AdminIDsSet 将配置中的管理员 ID 切片转换成 set 形式的 map，便于 O(1) 判定。
// nil/空切片返回 nil，语义与配置一致（未配置 → 全部放行）。
func AdminIDsSet(ids []int64) map[int64]bool {
	if len(ids) == 0 {
		return nil
	}
	set := make(map[int64]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

// isAuthorized 决定命令是否被允许执行：
//   - AdminIDs 为空（未配置）→ 放行，维持现状行为。
//   - cmd 属于放行白名单（如 /chatid）→ 放行，供任何人查询 chat ID 以便配置。
//   - senderID 在白名单内 → 放行。
func isAuthorized(adminIDs map[int64]bool, cmd string, senderID int64) bool {
	if len(adminIDs) == 0 {
		return true
	}
	if cmd == "/chatid" {
		return true
	}
	return adminIDs[senderID]
}

// authorizedOnly 构造一个鉴权中间件：仅放行管理员白名单用户的命令，其余静默丢弃。
func authorizedOnly(adminIDs map[int64]bool, cmd string, log logger.Log, next func(c tb.Context) error) func(c tb.Context) error {
	return func(c tb.Context) error {
		sender := c.Sender()
		senderID := int64(0) // sender 为 nil 时按未授权处理
		if sender != nil {
			senderID = sender.ID
		}
		if !isAuthorized(adminIDs, cmd, senderID) {
			log.Warn(
				"unauthorized command dropped",
				"command", cmd,
				"sender_id", senderID,
			)
			return nil // 静默丢弃，不回复
		}
		return next(c)
	}
}