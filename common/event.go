package common

type AioEventKind string

const (
	EventText           AioEventKind = "text"
	EventMarkdown       AioEventKind = "markdown"
	EventDeleteMarkdown AioEventKind = "delete_markdown"
	EventPhoto          AioEventKind = "photo"
	EventVideo          AioEventKind = "video"
	EventDocument       AioEventKind = "document"
	EventAudio          AioEventKind = "audio"
)

type AioEvent struct {
	Kind               AioEventKind
	ChatID             int64
	Text               string
	Path               string
	DisableWebPreview  bool
	DeleteAfterMinutes int
	Audio              AioAudio
}

type AioAudio struct {
	Thumb    string
	Duration int
	Path     string
}

func Text(msg string) AioEvent {
	return AioEvent{Kind: EventText, Text: msg}
}

func TextTo(chatID int64, msg string) AioEvent {
	return AioEvent{Kind: EventText, ChatID: chatID, Text: msg}
}

func Markdown(msg string, preview bool) AioEvent {
	return AioEvent{Kind: EventMarkdown, Text: msg, DisableWebPreview: preview}
}

func MarkdownTo(chatID int64, msg string, preview bool) AioEvent {
	return AioEvent{Kind: EventMarkdown, ChatID: chatID, Text: msg, DisableWebPreview: preview}
}

func DeleteMarkdownTo(chatID int64, msg string, preview bool, minutes int) AioEvent {
	return AioEvent{Kind: EventDeleteMarkdown, ChatID: chatID, Text: msg, DisableWebPreview: preview, DeleteAfterMinutes: minutes}
}

func PhotoTo(chatID int64, path string) AioEvent {
	return AioEvent{Kind: EventPhoto, ChatID: chatID, Path: path}
}

func Video(path string) AioEvent {
	return AioEvent{Kind: EventVideo, Path: path}
}

func Document(path string) AioEvent {
	return AioEvent{Kind: EventDocument, Path: path}
}

func Audio(audio AioAudio) AioEvent {
	return AioEvent{Kind: EventAudio, Audio: audio}
}

func Send(ch chan<- AioEvent, event AioEvent) {
	if ch == nil {
		return
	}
	ch <- event
}
