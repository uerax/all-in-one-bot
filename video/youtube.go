package video

import (
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/kkdai/youtube/v2"
	"github.com/uerax/all-in-one-bot/common"
)

func (v *VideoDownload) YoutubeAudioDownload(url string, startAndEnd ...string) {

	client := youtube.Client{}

	video, err := client.GetVideo(url)
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		return
	}

	filename := v.path + replaceSpecialChars(video.Title)

	formats := video.Formats.FindByItag(140)

	stream, _, err := client.GetStream(video, formats)
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		log.Println(err)
		return
	}

	if _, err := os.Stat(v.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(v.path, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地临时文件夹失败")
			v.MsgC <- "创建本地临时文件夹失败"
			return
		}
	}
	file, err := os.Create(filename + ".m4a")
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		log.Println(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		log.Println(err)
		return
	}

	audio_cfg := make([]interface{}, 3)
	if len(video.Thumbnails) > 0 {
		idx := 0
		if len(video.Thumbnails) > 1 {
			idx = 1
		}
		thumb, err := http.Get(video.Thumbnails[idx].URL)
		if err == nil {
			defer thumb.Body.Close()
			tmp, err := os.Create(filename + ".jpg")
			if err == nil {
				defer tmp.Close()
				_, err := io.Copy(tmp, thumb.Body)
				if err == nil {
					audio_cfg[0] = filename + ".jpg"
				}
			}
		}
	}
	audio_cfg[1] = int(video.Duration.Seconds())
	audio_cfg[2] = filename + ".m4a"

	if len(startAndEnd) == 2 {
		err = v.Cut(filename+".m4a", startAndEnd[0], startAndEnd[1], filename+"_C.m4a")
		if err != nil {
			v.MsgC <- "请检查是否安装ffmpeg"
			log.Println(err)
			return
		}
		audio_cfg[1] = common.TimeIntervalSecond(startAndEnd[0], startAndEnd[1])
		audio_cfg[2] = filename + "_C.m4a"
		v.AudioC <- audio_cfg
		go common.DeleteFileAfterTime(filename+"_C.m4a", 5)
		go common.DeleteFileAfterTime(filename+".m4a", 5)
		go common.DeleteFileAfterTime(filename+".jpg", 5)
		return
	}

	v.AudioC <- audio_cfg
	go common.DeleteFileAfterTime(filename+".m4a", 5)
	go common.DeleteFileAfterTime(filename+".jpg", 5)

}

func (v *VideoDownload) YoutubeDownload(url string, startAndEnd ...string) {

	client := youtube.Client{}

	video, err := client.GetVideo(url)
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		log.Println(err)
		return
	}

	filename := v.path + replaceSpecialChars(video.Title)

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		log.Println(err)
		return
	}

	if _, err := os.Stat(v.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(v.path, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地临时文件夹失败")
			v.MsgC <- "创建本地临时文件夹失败"
			return
		}
	}
	file, err := os.Create(filename + ".mp4")
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		log.Println(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		v.MsgC <- "出现异常,请重试"
		log.Println(err)
		return
	}

	if len(startAndEnd) == 2 {
		err = v.Cut(filename+".mp4", startAndEnd[0], startAndEnd[1], filename+"_C.mp4")
		if err != nil {
			v.MsgC <- "请检查是否安装ffmpeg"
			log.Println(err)
			return
		}

		v.C <- filename + "_C.mp4"
		go common.DeleteFileAfterTime(filename+"_C.mp4", 5)
		go common.DeleteFileAfterTime(filename+".mp4", 5)
		return
	}

	v.C <- filename + ".mp4"
	go common.DeleteFileAfterTime(filename+".mp4", 5)
}

func replaceSpecialChars(fileName string) string {

	re := regexp.MustCompile(`[\p{So}\p{Sk}\/\\:*?"<>| ]`)
	result := re.ReplaceAllString(fileName, "_")

	return result
}
