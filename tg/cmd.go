package tg

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Crypto Start
func addKlineStrategyProbe(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		fmt.Printf("addKlineStrategyProbe 参数有误: %s", args)
		return
	}
	api.CryptoV2Api.AddKLineProbe(arg[0])
}

func deleteKlineStrategyProbe(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		fmt.Printf("deleteKlineStrategyProbe 参数有误: %s", args)
		return
	}
	api.CryptoV2Api.StopKLineProbe(arg[0])
}

func addCryptoGrowthMonitor(id int64, args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		fmt.Printf("addCryptoGrowthMonitor 参数有误: %s", args)
		return
	}
	
	api.CryptoApi.AddHighMonitor(id, arg[0], arg[1])

}

func addCryptoDeclineMonitor(id int64, args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		fmt.Printf("addCryptoDeclineMonitor 参数有误: %s", args)
		return
	}
	
	api.CryptoApi.AddLowMonitor(id, arg[0], arg[1])
}

func getCryptoPrice(id int64, args string) {
	arg := strings.Split(args, " ")
	if args == "" {
		arg = nil
	}
	ctp := api.CryptoApi.GetPrice(id, arg...)
	sb := strings.Builder{}
	sb.WriteString("查询的结果如下:")
	for k, v := range ctp {
		sb.WriteString("\n")
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(v)
	}
	api.SendMsg(id, sb.String())	
}

func deleteCryptoMinitor(id int64, args string) {
	arg := strings.Split(args, ",")
	if args == "" {
		return
	}
	api.CryptoApi.DeleteMonitor(id, arg...)
}

func getUFuturesCryptoPrice(id int64, args string) {
	if args == "" {
		args = "BTCUSDT"
	}
	ctp := api.CryptoApi.GetUFuturePrice(id, args)
	sb := strings.Builder{}
	sb.WriteString(args)
	sb.WriteString(":")
	sb.WriteString(ctp)
	api.SendMsg(id, sb.String())	
}

// ChatGPT
func chatGPT(id int64, args string) {
	if args == "" {
		return
	}

	api.ChatGPTApi.Ask(id, args)
}

// Vps
func vpsMonitorSupportedList(id int64) {
	list := api.VpsApi.GetList()
	sb := strings.Builder{}
	sb.WriteString("目前可监控列表如下:")
	for _, v := range list {
		sb.WriteString("\n[")
		sb.WriteString(v.Name)
		sb.WriteString("] - ")
		sb.WriteString(v.Url)
	}
	api.SendMsg(id, sb.String())
}

func addVpsMonitor(id int64, args string) {
	if args == "" {
		return
	}

	api.VpsApi.AddMonitor(id, args)
}

// Photo
func cutouts(id int64, photos []tgbotapi.PhotoSize) {
	photo := photos[len(photos)-1] // get the largest available photo
	fileID := photo.FileID
	file, err := api.bot.GetFileDirectURL(fileID)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	api.PhotoApi.RemoveBackground(id, file)
}

// Telegram
func chatid(id int64) {
	api.SendMsg(id, fmt.Sprintf("你的ChatId为 : %d", id))
}

// Cron
func addCron(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		fmt.Printf("addCron 参数有误: %s", args)
		return
	}
	api.Cron.AddTask(arg[0], arg[1])
}

func deleteCron(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		fmt.Printf("deleteCron 参数有误: %s", args)
		return
	}
	api.Cron.CloseTask(arg[0])
}

// Video
func ytbDownload(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		fmt.Printf("ytbDownload 参数有误: %s", args)
		return
	}

	api.Video.YoutubeDownload(arg[0])
}

func ytbAudioDownload(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		fmt.Printf("ytbDownload 参数有误: %s", args)
		return
	}

	api.Video.YoutubeDownload(arg[0])
}

func ytbDownloadCut(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 3 {
		fmt.Printf("ytbDownload 参数有误: %s", args)
		return
	}

	api.Video.YoutubeDownload(arg[0], arg[1], arg[2])
}

func ytbAudioDownloadCut(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 3 {
		fmt.Printf("ytbAudioDownloadCut 参数有误: %s", args)
		return
	}

	api.Video.YoutubeDownload(arg[0], arg[1], arg[2])
}

// Default
func execute(id int64, args string) {
	if args == "" {
		return
	}

	api.ChatGPTApi.Ask(id, args)
}

func tips(id int64, msg string) {
	api.SendMsg(id, msg)
}