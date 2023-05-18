package tg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	newFileName := strings.TrimSuffix(fileName, fileExt) + ".jpg"
	filePath := filepath.Join(t.path, newFileName)
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

	t.C <- filePath
}