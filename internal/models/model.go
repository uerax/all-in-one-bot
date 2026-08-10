package models

type MessageKind int

const (
	KindText     MessageKind = iota
	KindMarkdown             // 使用 Telegram Markdown 格式发送
)

type Message struct {
	ChatID int64
	Text   string
	Kind   MessageKind
}
