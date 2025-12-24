package mocks

import (
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

type MockLogger struct{
	logger.Log
}

func (l *MockLogger) Info(msg string, args ...any) {
}

func (l *MockLogger) Error(msg string, args ...any) {
}