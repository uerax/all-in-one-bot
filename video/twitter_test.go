package video

import (
	"testing"
)

func TestVideoDownload_TwitterDownload(t *testing.T) {
	vd := NewVideoDownload()
	vd.path = "./"
	vd.TwitterDownload("https://twitter.com/SpaceX/status/1660401510969212929?s=20")
}

func TestVideoDownload_DouyinDownload(t *testing.T) {
	vd := NewVideoDownload()
	vd.path = "./"
	vd.DouyinDownload("https://v.douyin.com/UBtftcU/")
}
