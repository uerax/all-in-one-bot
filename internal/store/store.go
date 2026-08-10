package store

import (
	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

type Store interface {
	Set(database string, key string) (map[string]struct{}, error)
	Load(database string, key string, target any) error
	Save(database string, key string, value any) error
}

func NewStore(cfg config.Database, logger logger.Log) Store {
	switch cfg.Type {
	case "file":
		return NewFileStore(cfg, logger)
	default:
		return NewFileStore(cfg, logger)
	}
}
