package router

import (
	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/store"
)

// Dependencies 结构体持有所有 Handler 可能需要的服务（例如 DB连接、外部API客户端）。
type Dependencies struct {
	Config   *config.Config
	Logger   logger.Log
	Store    store.Store
	AdminIDs map[int64]bool // 管理员 ID 白名单；空则全部放行
}
