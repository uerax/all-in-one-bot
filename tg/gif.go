package tg

import (
	"fmt"
	"io"
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
	C    chan string
	MsgC chan string
	path string
}

func NewGif() *Gif {
	return &Gif{
		C:    make(chan string, 3),
		MsgC: make(chan string, 3),
		path: goconf.VarStringOrDefault("/tmp/aio-tgbot/gif/", "sticker", "path"),
	}
}

func (t *Gif) GifDownload(fileId string) {
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
	filePath := filepath.Join(t.path, newFileName+".mp4")
	if _, err := os.Stat(t.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.path, os.ModePerm) // 创建目录
		if err != nil {
			fmt.Println("创建本地临时文件夹失败")
			t.MsgC <- "创建本地临时文件夹失败"
			return
		}
	}
	fileLocal, err := os.Create(filePath + ".mp4")
	if err != nil {
		fmt.Printf("无法创建本地文件：%s\n", err.Error())
		t.MsgC <- "无法创建本地文件"
		return
	}
	defer fileLocal.Close()

	// 将下载的视频写入本地文件
	_, err = io.Copy(fileLocal, resp.Body)
	if err != nil {
		fmt.Printf("无法写入本地文件：%s\n", err.Error())
		t.MsgC <- "无法写入本地文件"
		return
	}

	args := []string{"-i", filePath + ".mp4", "-b", "2048k", filePath + ".gif"}

	cmd := exec.Command("ffmpeg", args...)
	if err = cmd.Run(); err != nil {
		fmt.Printf("mp4转gif失败：%s\n", err.Error())
		t.MsgC <- "mp4转gif失败"
		return
	}

	err = common.Zip(filePath+".gif", filePath+".zip")
	if err != nil {
		fmt.Printf("创建压缩包失败%s\n", err.Error())
		t.MsgC <- "创建压缩包失败"
		return
	}

	t.C <- filePath + ".zip"
	go common.DeleteFileAfterTime(filePath+".zip", 5)
}
