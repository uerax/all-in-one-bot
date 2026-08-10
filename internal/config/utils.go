package config

import (
	"log/slog"
	"os"
	"strconv"
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
