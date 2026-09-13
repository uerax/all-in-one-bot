package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

const defaultLocalDataDir = "/tmp/aio/data"

var (
	ErrorPathNotSet      = errors.New("没有填写数据库文件地址")
	ErrorDownloadFailed  = errors.New("下载JSON文件失败")
	ErrorReadFileFailed  = errors.New("无法读取JSON文件")
	ErrorParseFileFailed = errors.New("无法解析JSON文件")
)

type FileStore struct {
	mu           sync.RWMutex
	path         string // GitHub Raw URL 或 本地路径
	localDataDir string // 远程 URL 模式下的本地持久化目录，留空默认为 /tmp/aio/data
	log          logger.Log
	client       *http.Client
}

func (f *FileStore) getLocalDataDir() string {
	if f.localDataDir != "" {
		return f.localDataDir
	}
	return defaultLocalDataDir
}

func NewFileStore(cfg config.Database, logger logger.Log) *FileStore {
	return &FileStore{
		path:   cfg.FilePath,
		client: &http.Client{Timeout: 10 * time.Second},
		log:    logger,
	}
}

// Load 支持从网络 URL 或本地路径加载 JSON 数据到 target
func (f *FileStore) Load(database string, key string, target any) error {
	if f.path == "" {
		f.log.Error("没有填写数据库文件地址")
		return ErrorPathNotSet
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	var body []byte
	var err error

	if strings.HasPrefix(f.path, "http://") || strings.HasPrefix(f.path, "https://") {
		// 远程 URL 模式下，若本地已有持久化文件（由 Save 产生），优先读取本地修改
		localBase := f.getLocalDataDir()
		localPath := filepath.Join(localBase, database, key+".json")
		if _, statErr := os.Stat(localPath); os.IsNotExist(statErr) {
			localPath = filepath.Join(localBase, database, key+".dat")
		}
		if _, statErr := os.Stat(localPath); statErr == nil {
			body, err = os.ReadFile(localPath)
		} else {
			cleanPath := strings.TrimRight(f.path, "/")
			url := cleanPath + "/" + database + "/" + key + ".json"
			body, err = f.fetchURL(url)
			if err != nil {
				url = cleanPath + "/" + database + "/" + key + ".dat"
				body, err = f.fetchURL(url)
			}
		}
	} else {
		filePath := filepath.Join(f.path, database, key+".json")
		if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
			filePath = filepath.Join(f.path, database, key+".dat")
		}
		body, err = os.ReadFile(filePath)
	}

	if err != nil {
		f.log.Error("无法读取文件:", "database", database, "key", key, "error", err)
		return ErrorReadFileFailed
	}

	if err := json.Unmarshal(body, target); err != nil {
		f.log.Error("无法解析JSON数据:", "database", database, "key", key, "error", err)
		return ErrorParseFileFailed
	}

	return nil
}

func (f *FileStore) fetchURL(url string) ([]byte, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Save 将数据写入本地持久化目录
func (f *FileStore) Save(database string, key string, value any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		f.log.Error("序列化数据失败:", "error", err)
		return err
	}

	baseDir := f.path
	if strings.HasPrefix(f.path, "http://") || strings.HasPrefix(f.path, "https://") {
		baseDir = f.getLocalDataDir()
	}

	dir := filepath.Join(baseDir, database)
	if err := os.MkdirAll(dir, 0755); err != nil {
		f.log.Error("创建数据库目录失败:", "dir", dir, "error", err)
		return err
	}

	filePath := filepath.Join(dir, key+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		f.log.Error("保存数据库文件失败:", "file", filePath, "error", err)
		return err
	}

	return nil
}

// Set 保持兼容旧的 API
func (f *FileStore) Set(database string, key string) (map[string]struct{}, error) {
	data := make(map[string]struct{})
	err := f.Load(database, key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
