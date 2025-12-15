package config

import (
	"log/slog"
	"os"
	"strconv"
)

func strOrDefault(key, defaultVal string) string {
	if os.Getenv(key) == "" {
		return defaultVal
	}
	return os.Getenv(key)
}

func intOrDefault(key string, defaultVal int) int {
	envStr := os.Getenv(key)
	if envStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(envStr)
	if err != nil {
        // 记录警告日志，并返回默认值 (Fail Safe)
        slog.Warn(
			"Configuration parsing failed, using default value", 
			"env_value", envStr, 
			"default_value", defaultVal, 
			"error", err)
        return defaultVal
    }
	return val
}

func int64OrDefault(key string, defaultVal int64) int64 {
	envStr := os.Getenv(key)
	if envStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseInt(envStr, 10, 64)
	if err != nil {
        // 记录警告日志，并返回默认值 (Fail Safe)
        slog.Warn(
			"Configuration parsing failed, using default value", 
			"env_value", envStr, 
			"default_value", defaultVal, 
			"error", err)
        return defaultVal
    }
	return val
}

