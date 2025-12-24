package bitcointalk

import (
	"errors"
	"maps"
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/mocks"
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

