package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_YAMLAndEnvOverride(t *testing.T) {
	// 创建临时 YAML 配置文件
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
telegram:
  token: "yaml-token"
  timeout: 20
bitcointalk:
  limit: 100
  interval: 30
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入测试 YAML 失败: %v", err)
	}

	// 1. 无 env 时，测试 YAML 能否正确作为 base 基础值填充
	cfg := LoadConfig(yamlPath)
	if cfg.Telegram.Token != "yaml-token" {
		t.Errorf("期望 Telegram.Token 为 yaml-token, 实际为 %s", cfg.Telegram.Token)
	}
	if cfg.Telegram.Timeout != 20 {
		t.Errorf("期望 Telegram.Timeout 为 20, 实际为 %d", cfg.Telegram.Timeout)
	}
	if cfg.Bitcointalk.Limit != 100 {
		t.Errorf("期望 Bitcointalk.Limit 为 100, 实际为 %d", cfg.Bitcointalk.Limit)
	}
	// YAML 中未配置的字段应回退到默认值
	if cfg.Bitcointalk.Url != "https://bitcointalk.org/index.php?board=159.0" {
		t.Errorf("期望 Bitcointalk.Url 默认值, 实际为 %s", cfg.Bitcointalk.Url)
	}

	// 2. 环境变量应能覆盖 YAML 配置
	t.Setenv("TELEGRAM_TOKEN", "env-token")
	t.Setenv("BITCOINTALK_LIMIT", "200")

	cfgOverridden := LoadConfig(yamlPath)
	if cfgOverridden.Telegram.Token != "env-token" {
		t.Errorf("期望 TELEGRAM_TOKEN 覆盖为 env-token, 实际为 %s", cfgOverridden.Telegram.Token)
	}
	if cfgOverridden.Bitcointalk.Limit != 200 {
		t.Errorf("期望 BITCOINTALK_LIMIT 覆盖为 200, 实际为 %d", cfgOverridden.Bitcointalk.Limit)
	}
	// 未被 env 覆盖的字段仍保持 YAML 值
	if cfgOverridden.Telegram.Timeout != 20 {
		t.Errorf("期望 Telegram.Timeout 保持 YAML 值 20, 实际为 %d", cfgOverridden.Telegram.Timeout)
	}
}
