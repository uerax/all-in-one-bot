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

func int64OrDefault(key string, defaultVal int64) int64 {
	envStr := os.Getenv(key)
	if envStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseInt(envStr, 10, 64)
	if err != nil {
        // 记录警告日志，并返回默认值 (Fail Safe)
        slog.Warn("WARN: Failed to parse '%s' as int64, using default value %d. Error: %v", envStr, defaultVal, err)
        return defaultVal
    }
	return val
}

