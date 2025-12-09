package handler

import (
	tb "gopkg.in/telebot.v4"
)

// Handler 接口是所有命令处理程序必须实现的契约。
type Handler interface {
	Cmd() string
	Handle(c tb.Context) error
}