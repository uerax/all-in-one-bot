package video

import (
	"os/exec"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/goconf"
)

type VideoDownload struct {
	ch   chan<- common.AioEvent
	path string
}

func NewVideoDownload(ch ...chan<- common.AioEvent) *VideoDownload {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &VideoDownload{
		ch:   eventCh,
		path: goconf.VarStringOrDefault("/tmp/aio-tgbot/video/", "video", "path"),
	}
}

func (t *VideoDownload) sendVideo(path string) {
	common.Send(t.ch, common.Video(path))
}

func (t *VideoDownload) sendText(msg string) {
	common.Send(t.ch, common.Text(msg))
}

func (t *VideoDownload) sendAudio(thumb string, duration int, path string) {
	common.Send(t.ch, common.Audio(common.AioAudio{Thumb: thumb, Duration: duration, Path: path}))
}

func (t *VideoDownload) Cut(file, start, end, output string) error {
	args := []string{"-i", file, "-ss", start, "-to", end, "-c", "copy", output}
	cmd := exec.Command("ffmpeg", args...)
	err := cmd.Run()

	return err
}
