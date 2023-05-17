package tg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/uerax/goconf"
)

var Cmd string
var ChatId int64

func Server() {

	// Create a new bot instance
	token, err := goconf.VarString("telegram", "token")
	if err != nil {
		panic(err)
	}
	id := goconf.VarIntOrDefault(0, "telegram", "chatId")
	
	ChatId = int64(id)

	api.NewBot(token)

	// Create a new update channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Start listening for updates
	updates := api.bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}
		fmt.Println("receive msg : " + update.Message.Text)

		if ChatId != 0 && update.Message.Chat.ID != ChatId {
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			switch Cmd {
			// Crypto
			case "add_kline_strategy_probe":
				addKlineStrategyProbe(update.Message.Text)
			case "delete_kline_strategy_probe":
				deleteKlineStrategyProbe(update.Message.Text)
			case "add_crypto_growth_monitor":
				addCryptoGrowthMonitor(update.Message.Chat.ID, update.Message.Text)
			case "add_crypto_decline_monitor":
				addCryptoDeclineMonitor(update.Message.Chat.ID, update.Message.Text)
			case "get_crypto_price":
				getCryptoPrice(update.Message.Chat.ID, update.Message.Text)
			case "delete_crypto_minitor":
				deleteCryptoMinitor(update.Message.Chat.ID, update.Message.Text)
			case "get_crypto_ufutures_price":
				getUFuturesCryptoPrice(update.Message.Chat.ID, update.Message.Text)
			// Vps
			// case "vps_monitor_supported_list":
			// 	vpsMonitorSupportedList(update.Message.Chat.ID)
			// case "add_vps_monitor":
			// 	addVpsMonitor(update.Message.Chat.ID, update.Message.Text)
			// ChatGPT
			case "chatgpt":
				if goconf.VarBoolOrDefault(false, "telegram", "chat") {
					execute(update.Message.Chat.ID, update.Message.Text)
				}
			// Cutout
			case "cutout":
				if goconf.VarBoolOrDefault(false, "photo", "enable") && update.Message.Photo != nil {
					cutouts(update.Message.Chat.ID, update.Message.Photo)
				}
			case "delete_cron":
				deleteCron(update.Message.Text)
			case "add_cron":
				addCron(update.Message.Text)
			// Video
			case "youtube_download":
				ytbDownload(update.Message.Text)
			case "youtube_audio_download":
				ytbAudioDownload(update.Message.Text)
			case "youtube_download_cut":
				ytbDownloadCut(update.Message.Text)
			case "youtube_audio_download_cut":
				ytbAudioDownloadCut(update.Message.Text)
			}
			
		}

		switch update.Message.Command() {
		case "add_kline_strategy_probe":
			Cmd = "add_kline_strategy_probe"
			tips(update.Message.Chat.ID, "输入要监控的u本位合约 例: \nbtcusdt")
		case "delete_kline_strategy_probe":
			Cmd = "delete_kline_strategy_probe"
			tips(update.Message.Chat.ID, "输入要删除的u本位合约监控 例: \nbtcusdt")
		case "add_crypto_growth_monitor":
			Cmd = "add_crypto_growth_monitor"
			tips(update.Message.Chat.ID, "添加加密货币高线监控 例: \nBNB 1.1 (单位USDT)")
		case "add_crypto_decline_monitor":
			Cmd = "add_crypto_decline_monitor"
			tips(update.Message.Chat.ID, "添加加密货币低线监控 例: \nBNB 1.1 (单位USDT)")
		case "get_crypto_price":
			Cmd = "get_crypto_price"
			tips(update.Message.Chat.ID, "查询加密货币价格 例: \nBNB")
		case "delete_crypto_minitor":
			Cmd = "delete_crypto_minitor"
			tips(update.Message.Chat.ID, "删除加密货币监控线 例: \nBNB")
		case "get_crypto_ufutures_price":
			Cmd = "get_crypto_ufutures_price"
			tips(update.Message.Chat.ID, "查询加密货币合约价格 例: \nBNBUSDT")
		// Vps
		// case "vps_monitor_supported_list":
		// 	Cmd = "get_crypto_ufutures_price"
		// 	vpsMonitorSupportedList(update.Message.Chat.ID)
		// case "add_vps_monitor":
		// 	addVpsMonitor(update.Message.Chat.ID, update.Message.Text)
		// ChatGPT
		case "chatgpt":
			Cmd = "chatgpt"
			tips(update.Message.Chat.ID, "发送你的问题 例: \n今天的天气")
		// Cutout
		case "cutout":
			Cmd = "cutout"
			tips(update.Message.Chat.ID, "发送图片")
		// Telegram Info
		case "chatid":
			chatid(update.Message.Chat.ID)
		// Cron
		case "add_cron":
			Cmd = "add_cron"
			tips(update.Message.Chat.ID, "每隔多久一次提醒,单位/秒 例: 15 提醒内容(必填)")
		case "delete_cron":
			Cmd = "delete_cron"
			tips(update.Message.Chat.ID, "输入id(每次触发通知发送的消息带有) 例: 1")
		// Video
		case "youtube_download":
			Cmd = "youtube_download"
			tips(update.Message.Chat.ID, "输入url或者vid 例: ISqZpXkgNNs")			
		case "youtube_audio_download":
			Cmd = "youtube_audio_download"
			tips(update.Message.Chat.ID, "输入url或者vid 例: ISqZpXkgNNs")			
		case "youtube_download_cut":
			Cmd = "youtube_download_cut"
			tips(update.Message.Chat.ID, "输入url或者vid 开始时间 结束时间 例: ISqZpXkgNNs 100 2000(s)")
		case "youtube_audio_download_cut":
			Cmd = "youtube_download_cut"
			tips(update.Message.Chat.ID, "输入url或者vid 开始时间 结束时间 例: ISqZpXkgNNs 0:12:22 0:33:22")			
		}
		
	}
}
