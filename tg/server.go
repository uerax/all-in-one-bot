package tg

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/uerax/goconf"
)

var Cmd string
var ChatId int64

func Server() {
	// Create a new bot instance
	token, err := goconf.VarString("telegram", "token")
	if err != nil {
		log.Println("启动失败: 没有填写bot的token")
		return
	}
	id := goconf.VarIntOrDefault(0, "telegram", "chatId")

	ChatId = int64(id)

	api.NewBot(token, goconf.VarStringOrDefault("http://localhost:8081/", "telegram", "local"))

	// Create a new update channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	// Start listening for updates
	updates := api.bot.GetUpdatesChan(u)
	notice("System Notification: Aio 服务已启动")
	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}
		log.Println("receive msg : " + update.Message.Text)

		if ChatId != 0 && update.Message.Chat.ID != ChatId {
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			switch Cmd {
			// Track
			case "wallet_tx_interest_rate":
				walletTxInterestRate(update.Message.Text)
			case "price_highest":
				priceHighest(update.Message.Text)
			case "tacking_tax":
				trackingTax(update.Message.Text)
			case "get_tax":
				getTax(update.Message.Text)
			case "wallet_tx_info":
				walletTxInfo(update.Message.Text)
			case "bot_addr_finder":
				botAddrFinder(update.Message.Text)
			case "analyze_addr_token_profit":
				analyzeAddrTokenProfit(update.Message.Text)
			case "smart_addr_analyze":
				smartAddrAnalyze(update.Message.Text)
			case "smart_addr_finder":
				smartAddrFinder(update.Message.Text)
			case "smart_addr_finder_v2":
				smartAddrFinderV2(update.Message.Text)
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
			case "once":
				onceCron(update.Message.Text)
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
					stickerDownload(update.Message.Sticker.FileID, update.Message.Sticker.IsAnimated)
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
			case "ts_convert":
				timestampConvert(update.Message.Text)
			case "time_convert":
				timeConvert(update.Message.Text)
			case "json_format":
				jsonFormat(update.Message.Text)
			case "decimal2binary":
				decimal2Binary(update.Message.Text)
			case "decimal2hex":
				decimal2Hex(update.Message.Text)
			case "binary2hex":
				binary2Hex(update.Message.Text)
			case "binary2decimal":
				binary2Decimal(update.Message.Text)
			case "hex2decimal":
				hex2Decimal(update.Message.Text)
			case "hex2binary":
				hex2Binary(update.Message.Text)
			case "hex2string":
				hex2String(update.Message.Text)
			case "string2hex":
				string2Hex(update.Message.Text)
			case "string2binary":
				string2binary(update.Message.Text)
			case "string2decimal":
				string2Decimal(update.Message.Text)
			case "decimal2string":
				decimal2string(update.Message.Text)
			case "binary2string":
				binary2string(update.Message.Text)
			case "mining_cal":
				miningRewardCal(update.Message.Text)
			}
		}

		// Cmd Tip
		switch update.Message.Command() {
		// BBS
		case "nodeseek_rss":
			nodeseekMoniter()
		case "bitcointalk_rss":
			bitcointalkMoniter()
		case "bitcointalk_rss_stop":
			stopBitcointalkMoniter()
		// Track
		case "tracking_wallet_analyze":
			trackingWalletAnalyze()
		case "wallet_tx_interest_rate":
			Cmd = "wallet_tx_interest_rate"
			tips(update.Message.Chat.ID, "分析钱包近n条交易的可盈利率 例: \n`0xC100D16B937Cd4bD2672F3D2636602267aD65A8e 50`")
		case "price_highest":
			Cmd = "price_highest"
			tips(update.Message.Chat.ID, "查看时间区间最高价格(now可以是具体时间) 例: \n`0x9eac760d89805558d1a657b59bed313766e09e61 2023-08-15_02:36:35 now`")
		case "tacking_tax":
			Cmd = "tacking_tax"
			tips(update.Message.Chat.ID, "设置tax监控线20分钟后自动取消 例: \n`0x2890df158d76e584877a1d17a85fea3aeeb85aa6 10 10`")
		case "get_tax":
			Cmd = "get_tax"
			tips(update.Message.Chat.ID, "获取当前tax 例: \n`0x2890df158d76e584877a1d17a85fea3aeeb85aa6`")
		case "wallet_tx_info":
			Cmd = "wallet_tx_info"
			tips(update.Message.Chat.ID, "获取两日内买入的加密货币和时间 例: \n`0x2890df158d76e584877a1d17a85fea3aeeb85aa6`")
		case "bot_addr_finder":
			Cmd = "bot_addr_finder"
			tips(update.Message.Chat.ID, "分析高涨幅度币的早期买入地址 例: \n`0x2890df158d76e584877a1d17a85fea3aeeb85aa6 50 1`")
		case "tracking_lastest_tx":
			trackingLastestTx()
		case "analyze_addr_token_profit":
			Cmd = "analyze_addr_token_profit"
			tips(update.Message.Chat.ID, "分析钱包的指定加密货币总收益情况(钱包地址 加加密货币合约地址) 例:\n`0x1c8075cfc18cd17f5fb7743fba811603b819234c 0x808a57ef754c18e1d2cea5d6cf30f00eeeaa1273`")
		case "smart_addr_analyze":
			Cmd = "smart_addr_analyze"
			tips(update.Message.Chat.ID, "分析最早买入的钱包近期30次交易收益 例: \n`0x2890df158d76e584877a1d17a85fea3aeeb85aa6 10 1`")
		case "smart_addr_finder":
			Cmd = "smart_addr_finder"
			tips(update.Message.Chat.ID, "分析高涨幅度币的地址收益来寻找聪明地址 例: \n`0x2890df158d76e584877a1d17a85fea3aeeb85aa6 50 1`")
		case "smart_addr_finder_v2":
			Cmd = "smart_addr_finder_v2"
			tips(update.Message.Chat.ID, "分析高涨幅度币的地址收益来寻找聪明地址 例: \n`0x2890df158d76e584877a1d17a85fea3aeeb85aa6 50 1`")
		case "list_wallet_tracking":
			listWalletTracking()
		case "list_smart_addr_probe":
			listSmartAddrProbe()
		case "dump_tracking_list":
			dumpTrackingList()
		case "wallet_tx_analyze":
			Cmd = "wallet_tx_analyze"
			tips(update.Message.Chat.ID, "分析钱包近n条交易的利润 例: \n`0xC100D16B937Cd4bD2672F3D2636602267aD65A8e 30`")
		case "wallet_tracking":
			Cmd = "wallet_tracking"
			tips(update.Message.Chat.ID, "追踪聪明钱包买卖动态 例: \n`0xC100D16B937Cd4bD2672F3D2636602267aD65A8e alias`")
		case "stop_wallet_tracking":
			Cmd = "stop_wallet_tracking"
			stopWalletTrackingTip("停止追踪聪明钱包买卖动态 例: \n`0xC100D16B937Cd4bD2672F3D2636602267aD65A8e`")
		// Crypto
		case "coin_monitor":
			coingeckoMoniter()
		case "coin_stop":
			coingeckoStop()
		case "coin_price":
			coingeckoNow()
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
		case "once":
			onceCron(update.Message.Text)
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
		case "ts_convert":
			Cmd = "ts_convert"
			tips(update.Message.Chat.ID, "发送你的时间搓 例: `1686384050`")
		case "time_convert":
			Cmd = "time_convert"
			tips(update.Message.Chat.ID, "发送你的时间 格式为: `" + time.Now().Format("2006-01-02 15:04:05") + "`")
		case "json_format":
			Cmd = "json_format"
			tips(update.Message.Chat.ID, "发送json内容")
		case "decimal2binary":
			Cmd = "decimal2binary"
			tips(update.Message.Chat.ID, "发送十进制数")
		case "decimal2hex":
			Cmd = "decimal2hex"
			tips(update.Message.Chat.ID, "发送十进制数")
		case "binary2hex":
			Cmd = "binary2hex"
			tips(update.Message.Chat.ID, "发送二进制数")
		case "binary2decimal":
			Cmd = "binary2decimal"
			tips(update.Message.Chat.ID, "发送二进制数")
		case "hex2decimal":
			Cmd = "hex2decimal"
			tips(update.Message.Chat.ID, "发送十六进制数")
		case "hex2binary":
			Cmd = "hex2binary"
			tips(update.Message.Chat.ID, "发送十六进制数")
		case "hex2string":
			Cmd = "hex2string"
			tips(update.Message.Chat.ID, "发送十六进制数")
		case "string2hex":
			Cmd = "string2hex"
			tips(update.Message.Chat.ID, "发送字符串")
		case "mining_cal":
			Cmd = "mining_cal"
			tips(update.Message.Chat.ID, "发送你参数: 算力(k) Diff难度 区块奖励 时间(小时,可选) 价格(可选) 例: `40 214174659009 11 24 0.00000333`")
		case "qubic":
			qubic(update.Message.Text)
		case "qubic_sol":
			qubicSol(update.Message.Text)
		case "qubic_sol_acc":
			qubicAccSol(update.Message.Text)
		case "qubic_sol_all":
			qubicAccAll()
		case "qubic_token_refresh":
			qubicTokenRefresh()
		case "orge_withdraw":
			orgeWithdraw(update.Message.Text)
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
		// Test
		case "test":
			tips(update.Message.Chat.ID, "[发送字符串](`fdfffdafdf`)")
		}
	}
}
