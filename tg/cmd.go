package tg

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Crypto Start

func priceHighest(args string) {
	arg := strings.Split(args, " ")

	if len(arg) < 2 {
		log.Printf("priceHighest 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	
	if len(arg) == 2 {
		arg = append(arg, "now")
	}

	go api.Track.PriceHighestAndNow(arg[0], arg[1], arg[2], false)
}

func smartAddrAnalyze(args string) {
	arg := strings.Split(args, " ")
	if len(arg) == 1 {
		arg = append(arg, "5")
		arg = append(arg, "1")
	}

	if len(arg) != 3 {
		log.Printf("smartAddrFinder 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Track.SmartAddrAnalyze(arg[0], arg[1], arg[2])
}

func getTax(args string) {
	go api.Track.GetTax(args)
}

func trackingTax(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 3 {
		log.Printf("trackingTax 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	go api.Track.CronTaxTracking(arg[0], arg[1], arg[2])
}

func walletTxInfo(args string) {
	go api.Track.WalletTxInfo(args)
}

func trackingLastestTx() {
	go api.Track.WalletLastTransaction()
}

func analyzeAddrTokenProfit(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		log.Printf("analyzeAddrTokenProfit 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	go api.Track.AnalyzeAddrTokenProfit(arg[0], arg[1])
}

func smartAddrFinder(args string) {
	arg := strings.Split(args, " ")
	if len(arg) == 1 {
		arg = append(arg, "50")
		arg = append(arg, "1")
	}

	if len(arg) != 3 {
		log.Printf("smartAddrFinder 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Track.SmartAddrFinder(arg[0], arg[1], arg[2])
}

func smartAddrFinderV2(args string) {
	arg := strings.Split(args, " ")
	if len(arg) == 1 {
		arg = append(arg, "50")
		arg = append(arg, "1")
	}

	if len(arg) != 3 {
		log.Printf("smartAddrFinder 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Track.SmartAddrFinderV2(arg[0], arg[1], arg[2])
}

func botAddrFinder(args string) {
	arg := strings.Split(args, " ")
	if len(arg) == 1 {
		arg = append(arg, "50")
		arg = append(arg, "1")
	}

	if len(arg) != 3 {
		log.Printf("smartAddrFinder 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	go api.Track.BotAddrFinder(arg[0], arg[1], arg[2])
}

func listWalletTracking() {
	go api.Track.TrackingList(false)
}

func listSmartAddrProbe() {
	go api.CryptoV2Api.ListSmartAddr(false)
}

func dumpTrackingList() {
	go api.Track.DumpTrackingList(true)
}

func walletAnalyze(args string) {
	arg := strings.Split(args, " ")
	if len(arg) < 1 {
		log.Printf("walletAnalyze 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	if len(arg) == 1 {
		arg = append(arg, "40")
	}

	go api.Track.WalletTxAnalyzeV2(arg[0], arg[1], false)
}

func walletTracking(args string) {
	if args == "" {
		log.Printf("walletTracking 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	arg := strings.Split(args, " ")
	if len(arg) == 1 {
		arg = append(arg, "bot")
	}
	go api.Track.CronTracking(arg[0], arg[1])
}

func stopWalletTracking(args string) {
	go api.Track.StopTracking(args)
}

func stopWalletTrackingTip(tip string) {
	s := api.Track.TrackingList(true)
	go api.SendMarkdown(ChatId, s + "\n\n" + tip, true)
}

func setSmartAddrProbeItv(args string) {
	go api.CryptoV2Api.SetSmartAddrProbeItv(args)
}

func dumpSmartAddrList() {
	go api.CryptoV2Api.DumpSmartAddrList(true)
}

func smartAddrTx(args string) {
	arg := strings.Split(args, " ")
	if len(arg) < 1 {
		log.Printf("smartAddrTx 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	if len(arg) == 1 {
		arg = append(arg, "20")
	}
	go api.CryptoV2Api.SmartAddr(arg[0], arg[1])
}

func deleteSmartAddrProbe(args string) {
	go api.CryptoV2Api.DeleteSmartAddr(args)
}

func deleteSmartAddrProbeTip(tip string) {
	s := api.CryptoV2Api.ListSmartAddr(true)
	go api.SendMarkdown(ChatId, s + "\n\n" + tip, true)
}

func addSmartAddrProbe(args string) {
	go api.CryptoV2Api.AddSmartAddr(args)
}

func addKlineStrategyProbe(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		log.Printf("addKlineStrategyProbe 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	go api.CryptoV2Api.AddKLineProbe(arg[0])
}

func deleteKlineStrategyProbe(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		log.Printf("deleteKlineStrategyProbe 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	go api.CryptoV2Api.StopKLineProbe(arg[0])
}

func deleteKlineStrategyProbeTip() {
	msg := api.CryptoV2Api.ListKLineProbe()
	go api.DeleteAfterSendMarkdown(ChatId, msg+"\n\n请输入要删除的u本位合约 例: `btcusdt`", false)
}

func listKlineStrategyProbe() {
	msg := api.CryptoV2Api.ListKLineProbe()
	go api.SendMarkdown(ChatId, msg, false)
}

func addCryptoGrowthMonitor(id int64, args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		log.Printf("addCryptoGrowthMonitor 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.CryptoApi.AddHighMonitor(id, arg[0], arg[1])

}

func addCryptoDeclineMonitor(id int64, args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		log.Printf("addCryptoDeclineMonitor 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.CryptoApi.AddLowMonitor(id, arg[0], arg[1])
}

func getCryptoPrice(id int64, args string) {
	arg := strings.Split(args, " ")
	if args == "" {
		go api.DeleteAfterSendMessage("参数有误")
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
	go api.SendMsg(id, sb.String())
}

func deleteCryptoMinitor(id int64, args string) {
	arg := strings.Split(args, ",")
	if args == "" {
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	go api.CryptoApi.DeleteMonitor(id, arg...)
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
	go api.SendMsg(id, sb.String())
}

func addMemeGrowthMonitor(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 3 {
		log.Printf("addMemeGrowthMonitor 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.CryptoV2Api.MemeGrowthMonitor(arg[0], arg[1], arg[2])
}

func addMemeDeclineMonitor(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 3 {
		log.Printf("addMemeDeclineMonitor 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.CryptoV2Api.MemeDeclineMonitor(arg[0], arg[1], arg[2])
}

func deleteMemeMonitor(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		log.Printf("deleteMemeMonitor 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.CryptoV2Api.CloseMemeMonitor(arg[0], arg[1])
}

func getMeme(args string) {
	arg := strings.Split(args, " ")
	if len(arg) == 1 {
		arg = append(arg, "")
	}

	go api.CryptoV2Api.MemePrice(arg[0], arg[1])
}

func memeMonitorList() {
	go api.CryptoV2Api.MemeMonitorList()
}

// ChatGPT
func chatGPT(id int64, args string) {
	if args == "" {
		go api.SendMessage("参数有误")
		return
	}

	go api.ChatGPTApi.Ask(id, args)
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
	go api.SendMsg(id, sb.String())
}

func addVpsMonitor(id int64, args string) {
	if args == "" {
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.VpsApi.AddMonitor(id, args)
}

// Photo
func cutouts(id int64, photos []tgbotapi.PhotoSize) {
	photo := photos[len(photos)-1] // get the largest available photo
	fileID := photo.FileID
	file, err := api.bot.GetFileDirectURL(fileID)
	if err != nil {
		log.Println(err)
		go api.DeleteAfterSendMessage("图片读取失败,请重新发送")
		return
	}

	go api.PhotoApi.RemoveBackground(id, file)
}

// Telegram
func chatid(id int64) {
	go api.SendMarkdown(id, fmt.Sprintf("你的ChatId为 : `%d`", id), false)
}

// Cron
func addCron(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		log.Printf("addCron 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	go api.Cron.AddTask(arg[0], arg[1])
}

func deleteCron(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		log.Printf("deleteCron 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}
	go api.Cron.CloseTask(arg[0])
}

// Video
func ytbDownload(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		log.Printf("ytbDownload 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Video.YoutubeDownload(arg[0])
}

func ytbAudioDownload(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		log.Printf("ytbAudioDownload 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Video.YoutubeAudioDownload(arg[0])
}

func ytbDownloadCut(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 3 {
		log.Printf("ytbDownloadCut 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Video.YoutubeDownload(arg[0], arg[1], arg[2])
}

func ytbAudioDownloadCut(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 3 {
		log.Printf("ytbAudioDownloadCut 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Video.YoutubeAudioDownload(arg[0], arg[1], arg[2])
}

func bilibiliDownload(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		log.Printf("bilibiliDownload 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Video.BilibiliDownload(arg[0])
}

func douyinDownload(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		log.Printf("douyinDownload 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Video.DouyinDownload(arg[0])
}

func twitterDownload(args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 1 {
		log.Printf("twitterDownload 参数有误: %s", args)
		go api.DeleteAfterSendMessage("参数有误")
		return
	}

	go api.Video.TwitterDownload(arg[0])
}

// Sticker And Gif
func stickerDownload(args string) {
	go api.Sticker.StickerDownload(args)
}

func gifDownload(args string) {
	go api.Gif.GifDownload(args)
}

// Utils
func base64Encode(args string) {
	go api.Utils.Base64Encode(args)
}

func base64Decode(args string) {
	go api.Utils.Base64Decode(args)
}

func timestampConvert(args string) {
	go api.Utils.TimestampConvert(args)
}

func timeConvert(args string) {
	go api.Utils.TimeConvert(args)
}

func jsonFormat(args string) {
	go api.Utils.JsonFormat(args)
}

func decimal2Binary(args string) {
	go api.Utils.DecimalConv(args, 10, 2)
}

func decimal2Hex(args string) {
	go api.Utils.DecimalConv(args, 10, 16)
}

func binary2Hex(args string) {
	go api.Utils.DecimalConv(args, 2, 16)
}

func binary2Decimal(args string) {
	go api.Utils.DecimalConv(args, 2, 10)
}

func hex2Decimal(args string) {
	go api.Utils.DecimalConv(args, 16, 10)
}

func hex2Binary(args string) {
	go api.Utils.DecimalConv(args, 16, 2)
}

func hex2String(args string) {
	go api.Utils.Hex2String(args)
}

func string2Hex(args string) {
	go api.Utils.String2Hex(args)
}

func string2binary(args string) {
	go api.Utils.StrDecimalConv(args, 2)
}

func string2Decimal(args string) {
	go api.Utils.StrDecimalConv(args, 10)
}

func decimal2string(args string) {
	go api.Utils.DecimalStrConv(args, 10)
}

func binary2string(args string) {
	go api.Utils.DecimalStrConv(args, 2)
}

// Lists
func cryptoList() {
	go api.Lists.Crypto()
}
func imageList() {
	go api.Lists.Image()
}
func listsList() {
	go api.Lists.List()
}
func utilsList() {
	go api.Lists.Utils()
}
func videoList() {
	go api.Lists.Video()
}
func allList() {
	go api.Lists.All()
}

// Default
func tips(id int64, msg string) {
	go api.DeleteAfterSendMarkdown(id, msg, false)
}

func notice(msg string) {
	go api.SendMessage(msg)
}