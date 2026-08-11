package config

import (
	"reflect"
	"testing"
)

func TestParseAdminIDs(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []int64
	}{
		{"空字符串", "", nil},
		{"仅分隔符与空白", " , , ", nil},
		{"单个 ID", "111", []int64{111}},
		{"多个 ID", "111,222,333", []int64{111, 222, 333}},
		{"带空格", " 111 , 222 ", []int64{111, 222}},
		{"含非法项时跳过并保留合法项", "111,abc,222,", []int64{111, 222}},
		{"全部非法", "abc,def", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseAdminIDs(tc.raw)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseAdminIDs(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

func TestTelegramAdminIDs_Priority(t *testing.T) {
	// 1. env 覆盖 YAML：env 非空时直接采用 env 解析结果
	t.Setenv("TELEGRAM_ADMIN_IDS", "111,222")
	got := telegramAdminIDs("TELEGRAM_ADMIN_IDS", []int64{999})
	want := []int64{111, 222}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("env 覆盖失败: got %v, want %v", got, want)
	}

	// 2. env 为空时回退到 YAML base
	t.Setenv("TELEGRAM_ADMIN_IDS", "")
	got = telegramAdminIDs("TELEGRAM_ADMIN_IDS", []int64{999, 888})
	want = []int64{999, 888}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("YAML 回退失败: got %v, want %v", got, want)
	}

	// 3. env 与 base 均为空 → nil（语义为全部放行）
	got = telegramAdminIDs("TELEGRAM_ADMIN_IDS", nil)
	if got != nil {
		t.Errorf("双空应返回 nil, got %v", got)
	}
}

func TestLoadConfig_AdminIDs(t *testing.T) {
	// env 提供的管理员 ID 应进入 Telegram.AdminIDs
	t.Setenv("TELEGRAM_ADMIN_IDS", "111,222")
	t.Setenv("TELEGRAM_TOKEN", "env-token") // 复用现有测试惯例，确保 LoadConfig 正常初始化

	cfg := LoadConfig("")
	want := []int64{111, 222}
	if !reflect.DeepEqual(cfg.Telegram.AdminIDs, want) {
		t.Errorf("LoadConfig 未正确解析 AdminIDs: got %v, want %v", cfg.Telegram.AdminIDs, want)
	}

	// 未配置 env 时，AdminIDs 为 nil（全部放行）
	t.Setenv("TELEGRAM_ADMIN_IDS", "")
	cfg2 := LoadConfig("")
	if cfg2.Telegram.AdminIDs != nil {
		t.Errorf("未配置 AdminIDs 应为 nil, got %v", cfg2.Telegram.AdminIDs)
	}
}
