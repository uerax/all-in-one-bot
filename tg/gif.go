package tg

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/goconf"
)

type Gif struct {
	ch   chan<- common.AioEvent
	path string
}

func NewGif(ch ...chan<- common.AioEvent) *Gif {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &Gif{
		ch:   eventCh,
		path: goconf.VarStringOrDefault("/tmp/aio-tgbot/gif/", "sticker", "path"),
	}
}

func (t *Gif) sendText(msg string) {
	common.Send(t.ch, common.Text(msg))
}

func (t *Gif) sendDocument(path string) {
	common.Send(t.ch, common.Document(path))
}

func (t *Gif) GifDownload(fileId string) {
	fileConfig := tgbotapi.FileConfig{
		FileID: fileId,
	}
	file, err := api.bot.GetFile(fileConfig)
	if err != nil {
		log.Printf("无法获取贴纸文件：%s\n", err.Error())
		t.sendText("无法获取贴纸文件")
		return
	}

	token, err := goconf.VarString("telegram", "token")
	if err != nil {
		log.Printf("无法获取token：%s\n", err.Error())
		t.sendText("无法获取token")
		return
	}

	downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, file.FilePath)

	// 发送HTTP GET请求以下载贴纸文件
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Printf("无法下载贴纸文件：%s\n", err.Error())
		t.sendText("无法下载贴纸文件")
		return
	}
	defer resp.Body.Close()

	// 创建本地文件以保存下载的贴纸
	fileName := filepath.Base(file.FilePath)
	fileExt := filepath.Ext(fileName)
	newFileName := strings.TrimSuffix(fileName, fileExt)
	filePath := filepath.Join(t.path, newFileName+".mp4")
	if _, err := os.Stat(t.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.path, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地临时文件夹失败")
			t.sendText("创建本地临时文件夹失败")
			return
		}
	}
	fileLocal, err := os.Create(filePath + ".mp4")
	if err != nil {
		log.Printf("无法创建本地文件：%s\n", err.Error())
		t.sendText("无法创建本地文件")
		return
	}
	defer fileLocal.Close()

	// 将下载的视频写入本地文件
	_, err = io.Copy(fileLocal, resp.Body)
	if err != nil {
		log.Printf("无法写入本地文件：%s\n", err.Error())
		t.sendText("无法写入本地文件")
		return
	}

	args := []string{"-i", filePath + ".mp4", "-b", "2048k", filePath + ".gif"}

	cmd := exec.Command("ffmpeg", args...)
	if err = cmd.Run(); err != nil {
		log.Printf("mp4转gif失败：%s\n", err.Error())
		t.sendText("mp4转gif失败")
		return
	}

	err = common.Zip(filePath+".gif", filePath+".zip")
	if err != nil {
		log.Printf("创建压缩包失败%s\n", err.Error())
		t.sendText("创建压缩包失败")
		return
	}

	t.sendDocument(filePath + ".zip")
	go common.DeleteFileAfterTime(filePath+".zip", 5)
}
