package video

import (
	"io"
	"os"
	"time"

	"github.com/kkdai/youtube/v2"
)

func (v *VideoDownload) YoutubeDownload(url string, audioOnly bool)  {

	client := youtube.Client{}

	video, err := client.GetVideo(url)
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		panic(err)
	}

	filename := v.path + time.Now().Format("200612150405")

	if audioOnly {
		formats := video.Formats.FindByItag(140)
	
		stream, _, err := client.GetStream(video, formats)
		if err != nil {
			v.MsgC <- "出现异常,请重试"
			panic(err)
		}

		filename = filename + ".m4a"

		file, err := os.Create(filename)
		if err != nil {
			v.MsgC <- "出现异常,请重试"
			panic(err)
		}
		defer file.Close()

		_, err = io.Copy(file, stream)
		if err != nil {
			v.MsgC <- "出现异常,请重试"
			panic(err)
		}
		
	} else {
		formats := video.Formats.WithAudioChannels() // only get videos with audio
		stream, _, err := client.GetStream(video, &formats[0])
		if err != nil {
			v.MsgC <- "出现异常,请重试"
			panic(err)
		}
		filename = filename + ".mp4"
		file, err := os.Create(filename)
		if err != nil {
			v.MsgC <- "出现异常,请重试"
			panic(err)
		}
		defer file.Close()

		_, err = io.Copy(file, stream)
		if err != nil {
			v.MsgC <- "出现异常,请重试"
			panic(err)
		}

		v.C <- filename

	}
	
}

