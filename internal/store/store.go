package store

import "github.com/uerax/all-in-one-bot/lite/internal/config"

type Store interface {
    //Add(item string) error
    //Remove(item string) error
    //GetAll() (map[string]struct{}, error)
    //Contains(item string) bool
    Set(database string, key string) (map[string]struct{}, error)
    // Sync() error // 预留给 GitHub 下载或刷盘逻辑
}

func NewStore(cfg *config.Config) Store {
    switch cfg.Database.Type {
    case "file":
        return NewFileStore(cfg)
    default:
        return NewFileStore(cfg)
    }
}