package video

import "github.com/uerax/goconf"

type VideoDownload struct {
	C chan string
	MsgC chan string
	path string
}

func NewVideoDownload() *VideoDownload {
	return &VideoDownload{
		C: make(chan string, 3),
		MsgC: make(chan string, 3),
		path: goconf.VarStringOrDefault("/tmp/aio-tgbot/", "video", "path"),
	}
}