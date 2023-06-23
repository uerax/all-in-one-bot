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
		fmt.Println("启动失败: 没有填写bot的token")
		panic(err)
	}
	id := goconf.VarIntOrDefault(0, "telegram", "chatId")

	ChatId = int64(id)

	api.NewBot(token, goconf.VarStringOrDefault("http://localhost:8081/", "telegram", "local"))

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
			// Track
			case "wallet_tx_analyze":
				walletAnalyze(update.Message.Text)
			case "wallet_tracking":
				walletTracking(update.Message.Text)
			case "stop_wallet_tracking":
				stopWalletTracking(update.Message.Text)
			// Crypto
			case "set_smart_addr_probe_itv":
				setSmartAddrProbeItv(update.Message.Text)
			case "smart_addr_tx":
				smartAddrTx(update.Message.Text)
			case "smart_addr_probe":
				addSmartAddrProbe(update.Message.Text)
			case "delete_smart_addr_probe":
				deleteSmartAddrProbe(update.Message.Text)
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
			case "add_meme_growth_monitor":
				addMemeGrowthMonitor(update.Message.Text)
			case "add_meme_decline_monitor":
				addMemeDeclineMonitor(update.Message.Text)
			case "delete_meme_monitor":
				deleteMemeMonitor(update.Message.Text)
			case "get_meme":
				getMeme(update.Message.Text)
			// Vps
			case "vps_monitor_supported_list":
				vpsMonitorSupportedList(update.Message.Chat.ID)
			case "add_vps_monitor":
				addVpsMonitor(update.Message.Chat.ID, update.Message.Text)
			// ChatGPT
			case "chatgpt":
				chatGPT(update.Message.Chat.ID, update.Message.Text)
			// Cutout
			case "cutout":
				if update.Message.Photo != nil {
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
			case "bilibili_download":
				bilibiliDownload(update.Message.Text)
			case "douyin_download":
				douyinDownload(update.Message.Text)
			case "twitter_download":
				twitterDownload(update.Message.Text)
			// Sticker And Gif
			case "sticker_download":
				if update.Message.Sticker != nil {
					stickerDownload(update.Message.Sticker.FileID)
				}
			case "gif_download":
				if update.Message.Animation != nil {
					gifDownload(update.Message.Animation.FileID)
				}
			// Utils
			case "base64_encode":
				base64Encode(update.Message.Text)
			case "base64_decode":
				base64Decode(update.Message.Text)
			case "timestamp_convert":
				timestampConvert(update.Message.Text)
			case "time_convert":
				timeConvert(update.Message.Text)
			case "json_format":
				jsonFormat(update.Message.Text)
			}

		}

		switch update.Message.Command() {
		// Track
		case "list_wallet_tracking":
			listWalletTracking()
		case "list_smart_addr_probe":
			listSmartAddrProbe()
		case "dump_tracking_list":
			dumpTrackingList()
		case "wallet_tx_analyze":
			Cmd = "wallet_tx_analyze"
			tips(update.Message.Chat.ID, "分析钱包近n条交易的利润 例: \n`0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E 30`")
		case "wallet_tracking":
			Cmd = "wallet_tracking"
			tips(update.Message.Chat.ID, "追踪聪明钱包买卖动态 例: \n`0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E`")
		case "stop_wallet_tracking":
			Cmd = "stop_wallet_tracking"
			stopWalletTrackingTip("停止追踪聪明钱包买卖动态 例: \n`0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E`")
		// Crypto
		case "set_smart_addr_probe_itv":
			Cmd = "set_smart_addr_probe_itv"
			tips(update.Message.Chat.ID, "修改聪明地址探测频率(1-60分钟) 例: \n`15`")
		case "dump_smart_addr_probe_list":
			dumpSmartAddrList()
		case "smart_addr_tx":
			Cmd = "smart_addr_tx"
			tips(update.Message.Chat.ID, "输入聪明地址(eth)和近n条交易 例: \n`0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80 20`")
		case "smart_addr_probe":
			Cmd = "smart_addr_probe"
			tips(update.Message.Chat.ID, "输入需要监控的聪明地址(eth) 例: \n`0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80`")
		case "delete_smart_addr_probe":
			Cmd = "delete_smart_addr_probe"
			deleteSmartAddrProbeTip("输入关闭监控的聪明地址(eth) 例: \n`0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80`")
		case "add_kline_strategy_probe":
			Cmd = "add_kline_strategy_probe"
			tips(update.Message.Chat.ID, "输入要监控的u本位合约 例: \n`btcusdt`")
		case "delete_kline_strategy_probe":
			Cmd = "delete_kline_strategy_probe"
			deleteKlineStrategyProbeTip()
		case "list_kline_strategy_probe":
			listKlineStrategyProbe()
		case "meme_monitor_list":
			memeMonitorList()
		case "add_crypto_growth_monitor":
			Cmd = "add_crypto_growth_monitor"
			tips(update.Message.Chat.ID, "添加加密货币高线监控 例: \n`BNB 1.1`")
		case "add_crypto_decline_monitor":
			Cmd = "add_crypto_decline_monitor"
			tips(update.Message.Chat.ID, "添加加密货币低线监控 例: \n`BNB 1.1`")
		case "get_crypto_price":
			Cmd = "get_crypto_price"
			tips(update.Message.Chat.ID, "查询加密货币价格 例: \n`BNB`")
		case "delete_crypto_minitor":
			Cmd = "delete_crypto_minitor"
			tips(update.Message.Chat.ID, "删除加密货币监控线 例: \n`BNB`")
		case "get_crypto_ufutures_price":
			Cmd = "get_crypto_ufutures_price"
			tips(update.Message.Chat.ID, "查询加密货币合约价格 例: \n`BNBUSDT`")
		case "add_meme_growth_monitor":
			Cmd = "add_meme_growth_monitor"
			tips(update.Message.Chat.ID, "添加加meme币高线监控 token chain price 例: \n`0x6982508145454ce325ddbe47a25d4ec3d2311933 eth 0.00000123`")
		case "add_meme_decline_monitor":
			Cmd = "add_meme_decline_monitor"
			tips(update.Message.Chat.ID, "添加加meme币低线监控 token chain price 例: \n`0x6982508145454ce325ddbe47a25d4ec3d2311933 eth 0.00000123`")
		case "delete_meme_monitor":
			Cmd = "delete_meme_monitor"
			tips(update.Message.Chat.ID, "删除meme币监控 token chain 例: \n`0x6982508145454ce325ddbe47a25d4ec3d2311933 eth`")
		case "get_meme":
			Cmd = "get_meme"
			tips(update.Message.Chat.ID, "获取meme币信息 token chain(可选,默认eth) 例: \n`0x6982508145454ce325ddbe47a25d4ec3d2311933 eth`")
		// Vps
		case "vps_monitor_supported_list":
			tips(update.Message.Chat.ID, "该功能已弃用")
		case "add_vps_monitor":
			tips(update.Message.Chat.ID, "该功能已弃用")
		//ChatGPT
		case "chatgpt":
			Cmd = "chatgpt"
			tips(update.Message.Chat.ID, "发送你的问题 例: \n`今天的天气`")
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
			tips(update.Message.Chat.ID, "每隔多久一次提醒,单位/秒 例: `15 提醒内容(必填)`")
		case "delete_cron":
			Cmd = "delete_cron"
			tips(update.Message.Chat.ID, "输入id(每次触发通知发送的消息带有) 例: `1`")
		// Video
		case "youtube_download":
			Cmd = "youtube_download"
			tips(update.Message.Chat.ID, "输入url或者vid 例: `ISqZpXkgNNs`")
		case "youtube_audio_download":
			Cmd = "youtube_audio_download"
			tips(update.Message.Chat.ID, "输入url或者vid 例: `ISqZpXkgNNs`")
		case "youtube_download_cut":
			Cmd = "youtube_download_cut"
			tips(update.Message.Chat.ID, "输入url或者vid 开始时间 结束时间(s) \n例: `ISqZpXkgNNs 100 2000`")
		case "youtube_audio_download_cut":
			Cmd = "youtube_audio_download_cut"
			tips(update.Message.Chat.ID, "输入url或者vid 开始时间 结束时间(s) \n例: `ISqZpXkgNNs 0:12:22 0:33:22`")
		case "bilibili_download":
			Cmd = "bilibili_download"
			tips(update.Message.Chat.ID, "输入url 例: `https://www.bilibili.com/video/BV1xP411d74V`")
		case "douyin_download":
			Cmd = "douyin_download"
			tips(update.Message.Chat.ID, "输入url 例: `https://v.douyin.com/UBt7heL/`")
		case "twitter_download":
			Cmd = "twitter_download"
			tips(update.Message.Chat.ID, "输入url 例: `https://twitter.com/SpaceX/status/1660401510969212929?s=20`")
		// Sticker And Gif
		case "gif_download":
			Cmd = "gif_download"
			tips(update.Message.Chat.ID, "发送需要转换的GIF图")
		case "sticker_download":
			Cmd = "sticker_download"
			tips(update.Message.Chat.ID, "发送你的贴纸")
		// Utils
		case "base64_encode":
			Cmd = "base64_encode"
			tips(update.Message.Chat.ID, "发送你要base64加密的内容")
		case "base64_decode":
			Cmd = "base64_decode"
			tips(update.Message.Chat.ID, "发送你的base64编码")
		case "timestamp_convert":
			Cmd = "timestamp_convert"
			tips(update.Message.Chat.ID, "发送你的时间搓 例:1686384050")
		case "time_convert":
			Cmd = "time_convert"
			tips(update.Message.Chat.ID, "发送你的时间 格式为: 2023-06-10 16:00:50")
		case "json_format":
			Cmd = "json_format"
			tips(update.Message.Chat.ID, "发送json内容")
		// Lists
		case "cmd_list":
			allList()
		case "crypto_cmd_list":
			cryptoList()
		case "video_cmd_list":
			videoList()
		case "image_cmd_list":
			imageList()
		case "utils_cmd_list":
			utilsList()
		case "list_cmd_list":
			listsList()
		}

	}
}
