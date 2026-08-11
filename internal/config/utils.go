package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// strOrDefaultBase 在 env 为空时回退到 YAML base 值（非空），base 也未填则用内置默认值。
func strOrDefaultBase(key, baseVal, defaultVal string) string {
	if envStr := os.Getenv(key); envStr != "" {
		return envStr
	}
	if baseVal != "" {
		return baseVal
	}
	return defaultVal
}

// intOrDefaultBase 在 env 为空时回退到 YAML base 值（非零），base 也未填则用内置默认值。
// YAML 中未配置的数值字段以 0 表示"未填"，本项目所有 int 字段合法值均 >0。
func intOrDefaultBase(key string, baseVal, defaultVal int) int {
	envStr := os.Getenv(key)
	if envStr != "" {
		val, err := strconv.Atoi(envStr)
		if err != nil {
			slog.Warn(
				"Configuration parsing failed, using default value",
				"env_value", envStr,
				"default_value", defaultVal,
				"error", err)
			return defaultVal
		}
		return val
	}
	if baseVal != 0 {
		return baseVal
	}
	return defaultVal
}

// int64OrDefaultBase 同 intOrDefaultBase，针对 int64 字段。
func int64OrDefaultBase(key string, baseVal, defaultVal int64) int64 {
	envStr := os.Getenv(key)
	if envStr != "" {
		val, err := strconv.ParseInt(envStr, 10, 64)
		if err != nil {
			slog.Warn(
				"Configuration parsing failed, using default value",
				"env_value", envStr,
				"default_value", defaultVal,
				"error", err)
			return defaultVal
		}
		return val
	}
	if baseVal != 0 {
		return baseVal
	}
	return defaultVal
}

// parseAdminIDs 将逗号分隔的 ID 字符串解析为 ID 切片。
// 非法项记 slog.Warn 后跳过；空或全部非法时返回 nil（语义为“全部放行”）。
func parseAdminIDs(raw string) []int64 {
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			slog.Warn(
				"Configuration parsing failed, admin id ignored",
				"raw_value", s,
				"error", err)
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}
	return ids
}

// telegramAdminIDs 在 env 为空时回退到 YAML base 值（非空），base 也未填则返回 nil。
// env 用逗号分隔字符串，YAML 用列表形式；nil 表示未配置 → 全部放行。
func telegramAdminIDs(key string, base []int64) []int64 {
	if envStr := os.Getenv(key); envStr != "" {
		return parseAdminIDs(envStr)
	}
	return base
}

// float64OrDefaultBase 同 float64OrDefaultBase 语义，针对 float64 字段。
func float64OrDefaultBase(key string, baseVal, defaultVal float64) float64 {
	envStr := os.Getenv(key)
	if envStr != "" {
		val, err := strconv.ParseFloat(envStr, 64)
		if err != nil {
			slog.Warn(
				"Configuration parsing failed, using default value",
				"env_value", envStr,
				"default_value", defaultVal,
				"error", err)
			return defaultVal
		}
		return val
	}
	if baseVal != 0 {
		return baseVal
	}
	return defaultVal
}
