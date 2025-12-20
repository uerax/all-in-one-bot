package store

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

var (
	ErrorPathNotSet = errors.New("没有填写数据库文件地址")
	ErrorDownloadFailed = errors.New("下载JSON文件失败")
	ErrorReadFileFailed = errors.New("无法读取JSON文件")
	ErrorParseFileFailed = errors.New("无法解析JSON文件")
)

type FileStore struct {
    mu       sync.RWMutex
    path     string // GitHub Raw URL
    data     map[string]struct{}
	log		 logger.Log
}

func NewFileStore(cfg *config.Config) *FileStore {
	
    file := &FileStore{
        path:   cfg.Database.FilePath,
        data:   make(map[string]struct{}),
    }
	return file
}

func (f *FileStore) Set(database string, key string) (map[string]struct{}, error) {
	if f.path == "" {
		f.log.Error("没有填写数据库文件地址")
        return nil, ErrorPathNotSet
    }

	resp, err := http.Get("/" + f.path + database + "/" + key + ".dat")
	if err != nil {
		f.log.Error("无法下载JSON文件:", "error:", err.Error())
		return nil, ErrorDownloadFailed
	}
	defer resp.Body.Close()

	// 读取JSON文件内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		f.log.Error("无法读取JSON文件:", "error:", err.Error())
		return nil, ErrorReadFileFailed
	}

	// 解析JSON为Map
	data := make(map[string]struct{})
	if err := json.Unmarshal(body, &data); err != nil {
		f.log.Error("无法解析JSON文件:", "error:", err.Error())
		return nil, ErrorParseFileFailed
	}
	return data, nil
}
