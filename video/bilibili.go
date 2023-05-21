package video

import (
	"fmt"

	"github.com/iawia002/lux/downloader"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/bilibili"
)


func (v *VideoDownload) BilibiliDownload(url string) {
	d, _ := bilibili.New().Extract(url, extractors.Options{Playlist: false})
	for _, stream := range d {
		err := downloader.New(downloader.Options{OutputPath: v.path, Stream: "32-12"}).Download(stream)
		fmt.Println(err)
	}
	
}