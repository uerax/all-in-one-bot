package tg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"tg-aio-bot/common"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/uerax/goconf"
)

type Sticker struct {
	C    chan string
	MsgC chan string
	path string
}

func NewSticker() *Sticker {
	return &Sticker{
		C:    make(chan string, 3),
		MsgC: make(chan string, 3),
		path: goconf.VarStringOrDefault("/tmp/aio-tgbot/sticker/", "sticker", "path"),
	}
}

func (t *Sticker) StickerDownload(fileId string) {
	fileConfig := tgbotapi.FileConfig{
		FileID: fileId,
	}
	file, err := api.bot.GetFile(fileConfig)
	if err != nil {
		fmt.Printf("无法获取贴纸文件：%s\n", err.Error())
		t.MsgC <- "无法获取贴纸文件"
		return
	}

	token, err := goconf.VarString("telegram", "token")
	if err != nil {
		fmt.Printf("无法获取token：%s\n", err.Error())
		t.MsgC <- "无法获取token"
		return
	}

	downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, file.FilePath)

	// 发送HTTP GET请求以下载贴纸文件
	resp, err := http.Get(downloadURL)
	if err != nil {
		fmt.Printf("无法下载贴纸文件：%s\n", err.Error())
		t.MsgC <- "无法下载贴纸文件"
		return
	}
	defer resp.Body.Close()

	// 创建本地文件以保存下载的贴纸
	fileName := filepath.Base(file.FilePath)
	fileExt := filepath.Ext(fileName)
	newFileName := strings.TrimSuffix(fileName, fileExt)
	filePath := filepath.Join(t.path, newFileName)

	if strings.ToLower(fileExt) == ".webp" {
		filePath = filePath + ".jpg"
		if _, err := os.Stat(t.path); os.IsNotExist(err) { // 检查目录是否存在
			err := os.Mkdir(t.path, os.ModePerm) // 创建目录
			if err != nil {
				fmt.Println("创建本地临时文件夹失败")
				t.MsgC <- "创建本地临时文件夹失败"
				return
			}
		}
		fileLocal, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("无法创建本地文件：%s\n", err.Error())
			t.MsgC <- "无法创建本地文件"
			return
		}
		defer fileLocal.Close()

		// 将下载的贴纸写入本地文件
		_, err = io.Copy(fileLocal, resp.Body)
		if err != nil {
			fmt.Printf("无法写入本地文件：%s\n", err.Error())
			t.MsgC <- "无法写入本地文件"
			return
		}
	} else if strings.ToLower(fileExt) == ".webm" {

		fileLocal, err := os.Create(filePath + ".webm")
		if err != nil {
			fmt.Printf("无法创建本地文件：%s\n", err.Error())
			t.MsgC <- "无法创建本地文件"
			return
		}
		defer fileLocal.Close()

		// 将下载的贴纸写入本地文件
		_, err = io.Copy(fileLocal, resp.Body)
		if err != nil {
			fmt.Printf("无法写入本地文件：%s\n", err.Error())
			t.MsgC <- "无法写入本地文件"
			return
		}

		args := []string{"-i", filePath + ".webm", "-b", "2048k", filePath + ".gif"}

		cmd := exec.Command("ffmpeg", args...)
		if err = cmd.Run(); err != nil {
			fmt.Printf("sticker转gif失败：%s\n", err.Error())
			t.MsgC <- "sticker转gif失败, 检查是否安装ffmpeg"
			return
		}

		err = common.Zip(filePath+".gif", filePath+".zip")
		if err != nil {
			fmt.Printf("创建压缩包失败%s\n", err.Error())
			t.MsgC <- "创建压缩包失败"
			return
		}

		filePath = filePath + ".zip"
	}

	t.C <- filePath
	go common.DeleteFileAfterTime(filePath, 5)
}
