package video

import "testing"

func TestVideoDownload_BilibiliDownload(t *testing.T) {
	vd := NewVideoDownload()
	vd.path = ""
	vd.BilibiliDownload("https://www.bilibili.com/video/BV11M411g7fD")
}
