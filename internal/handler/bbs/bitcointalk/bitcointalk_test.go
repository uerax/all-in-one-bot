package bitcointalk

import (
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/mocks"
	"github.com/uerax/all-in-one-bot/lite/internal/models"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
	"github.com/uerax/all-in-one-bot/lite/internal/store"
)

func TestBitcointalkHandle_syncFilter(t *testing.T) {
	// 💡 优化点 1: 仅保留 syncFilter 逻辑中真正用到的依赖
	type fields struct {
		db     store.Store
		logger logger.Log
		// filter 的初始状态（可选，用于测试是否正确覆盖旧数据）
		initialFilter map[string]struct{}
	}

	tests := []struct {
		name   string
		fields fields
		want   map[string]struct{}
	}{
		{
			name: "正常同步：数据库返回新关键字列表",
			fields: fields{
				db: &MockStore{
					SetFunc: func(db, k string) (map[string]struct{}, error) {
						return map[string]struct{}{"coldwater": {}, "STRONGS Coin": {}}, nil
					},
				},
				logger:        &mocks.MockLogger{},
				initialFilter: map[string]struct{}{"old_data": {}}, // 模拟已有旧数据
			},
			want: map[string]struct{}{"coldwater": {}, "STRONGS Coin": {}},
		},
		{
			name: "异常处理：数据库报错时不应清空原列表",
			fields: fields{
				db: &MockStore{
					SetFunc: func(db, k string) (map[string]struct{}, error) {
						return nil, errors.New("db error")
					},
				},
				logger:        &mocks.MockLogger{},
				initialFilter: map[string]struct{}{"stay_safe": {}},
			},
			want: map[string]struct{}{"stay_safe": {}}, // 期望保持原样
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 💡 优化点 2: 仅初始化必要的字段，其余字段会自动设为零值
			tr := &BitcointalkHandle{
				db:     tt.fields.db,
				Logger: tt.fields.logger,
				filter: tt.fields.initialFilter,
			}

			tr.syncFilter()

			// 💡 优化点 3: 使用标准报错格式，清晰展示失败原因
			if !maps.Equal(tt.want, tr.filter) {
				t.Errorf("\n[Case: %s]\ngot:  %v\nwant: %v", tt.name, tr.filter, tt.want)
			}
		})
	}
}

func TestBitcointalkHandle_monitor(t *testing.T) {
	path := filepath.Join("testdata", "testdata.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("无法读取测试样本文件: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	mockChan := make(chan models.Message, 100) // 真实数据多，通道开大点
	type fields struct {
		url      string
		filter   map[string]struct{}
		limit    int
		active   bool
		notified store.LRU
		C        chan models.Message
		client   *http.Client
		Logger   logger.Log
		Config   *config.Bitcointalk
	}
	type args struct {
		chatID int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   chan<- models.Message
	}{
		// TODO: Add test cases.
		{
			name: "正常启动测试",
			fields: fields{
				url: server.URL,
				filter: map[string]struct{}{
					"privacy-focused": {},
					"CPU Mineable":    {},
				},
				limit:  100,
				active: true,
				notified: &MockLRU{
					SeenFunc: func(key string) bool {
						return false
					},
				},
				C:      mockChan,
				client: server.Client(),
				Logger: &mocks.MockLogger{},
				Config: &config.Bitcointalk{},
			},
			want: mockChan,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BitcointalkHandle{
				url:      tt.fields.url,
				filter:   tt.fields.filter,
				limit:    tt.fields.limit,
				active:   tt.fields.active,
				notified: tt.fields.notified,
				C:        tt.fields.C,
				client:   tt.fields.client,
				Logger:   tt.fields.Logger,
				Config:   tt.fields.Config,
			}
			b.monitor(tt.args.chatID)
			expectedCount := 9
			actualCount := len(mockChan)
			if actualCount != expectedCount {
				t.Errorf("解析数量不符: 期望 %d, 实际得到 %d", expectedCount, actualCount)
			}
		})
	}
}
