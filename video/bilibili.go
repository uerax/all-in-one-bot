package video

import (
	"fmt"

	"github.com/iawia002/lux/downloader"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/bilibili"
)


func (v *VideoDownload) BilibiliDownload(url string) {
	d, _ := bilibili.New().Extract(url, extractors.Options{Playlist: false})
	if len(d) == 0 {
		v.MsgC <- "链接无效,解析视频为空"
		fmt.Println("链接无效,解析视频为空")
		return
	}
	if _, ok := d[0].Streams["32-12"]; ok {
		d[0].Streams["32-12"].Ext = "mp4"
		err := downloader.New(downloader.Options{OutputPath: v.path, Stream: "32-12"}).Download(d[0])
		if err != nil {
			v.MsgC <- "视频下载失败"
			fmt.Println(err)
			return
		}
		filename := v.path + d[0].Title + ".mp4"
		v.C <- filename
	}
	
}