package video

import (
	"fmt"
	"tg-aio-bot/common"

	"github.com/iawia002/lux/downloader"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/douyin"
)

func (v *VideoDownload) DouyinDownload(url string) {
	d, err := douyin.New().Extract(url, extractors.Options{Playlist: false})
	if err != nil || len(d) == 0 {
		v.MsgC <- "链接无效,解析视频为空"
		fmt.Println("链接无效,解析视频为空")
		return
	}
	if _, ok := d[0].Streams["default"]; ok {
		err := downloader.New(downloader.Options{OutputPath: v.path, Stream: "default"}).Download(d[0])
		if err != nil {
			v.MsgC <- "视频下载失败"
			fmt.Println(err)
			return
		}
		filename := v.path + d[0].Title + ".mp4"
		v.C <- filename
		go common.DeleteFileAfterTime(v.path+d[0].Title+".mp4", 5)
	}
}
