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

type Sticker struct {
	ch   chan<- common.AioEvent
	path string
}

func NewSticker(ch ...chan<- common.AioEvent) *Sticker {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &Sticker{
		ch:   eventCh,
		path: goconf.VarStringOrDefault("/tmp/aio-tgbot/sticker/", "sticker", "path"),
	}
}

func (t *Sticker) sendText(msg string) {
	common.Send(t.ch, common.Text(msg))
}

func (t *Sticker) sendDocument(path string) {
	common.Send(t.ch, common.Document(path))
}

func (t *Sticker) StickerDownload(fileId string, gif bool) {
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
	filePath := filepath.Join(t.path, newFileName)

	if _, err := os.Stat(t.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.path, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地临时文件夹失败")
			t.sendText("创建本地临时文件夹失败")
			return
		}
	}

	if !gif {
		filePath = filePath + ".jpg"
		fileLocal, err := os.Create(filePath)
		if err != nil {
			log.Printf("无法创建本地文件：%s\n", err.Error())
			t.sendText("无法创建本地文件")
			return
		}
		defer fileLocal.Close()

		// 将下载的贴纸写入本地文件
		_, err = io.Copy(fileLocal, resp.Body)
		if err != nil {
			log.Printf("无法写入本地文件：%s\n", err.Error())
			t.sendText("无法写入本地文件")
			return
		}
	} else {
		fileLocal, err := os.Create(t.path + fileName)
		if err != nil {
			log.Printf("无法创建本地文件：%s\n", err.Error())
			t.sendText("无法创建本地文件")
			return
		}
		defer fileLocal.Close()

		// 将下载的贴纸写入本地文件
		_, err = io.Copy(fileLocal, resp.Body)
		if err != nil {
			log.Printf("无法写入本地文件：%s\n", err.Error())
			t.sendText("无法写入本地文件")
			return
		}

		args := []string{"-i", t.path + fileName, "-b", "2048k", filePath + ".gif"}

		cmd := exec.Command("ffmpeg", args...)
		if err = cmd.Run(); err != nil {
			log.Printf("sticker转gif失败：%s\n", err.Error())
			t.sendText("sticker转gif失败, 检查是否安装ffmpeg")
			return
		}

		err = common.Zip(filePath+".gif", filePath+".zip")
		if err != nil {
			log.Printf("创建压缩包失败%s\n", err.Error())
			t.sendText("创建压缩包失败")
			return
		}

		filePath = filePath + ".zip"
	}

	t.sendDocument(filePath)
	go common.DeleteFileAfterTime(filePath, 5)
}
