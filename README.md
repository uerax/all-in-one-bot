## 一. Usage

* 在 all-in-one-bot.yml 添加你的 telegram token
`telegram 搜索用户 @BotFather 发送 /newbot 获取`
* 在 all-in-one-bot.yml 添加你的 telegram chatId
`添加token后启动应用,去你的bot发送 /chatid 即可获取`
    
## 二. Command List
### 1. 加密货币监控功能清单

__单位默认USDT,可在配置文件crypto -> unit修改__

- [x] add_kline_strategy_probe 探测连续3根一直走势的k线 例: btcusdt
- [x] delete_kline_strategy_probe 删除探测 例: btcusdt
- [x] list_kline_strategy_probe 列出当前探测的加密货币
- [x] add_crypto_growth_monitor 加密货币 提示价格 例: BNB 1110
- [x] add_crypto_decline_monitor 加密货币 提示价格 例: BNB 1110
- [x] get_crypto_price 加密货币[可选]
- [x] delete_crypto_minitor 加密货币(多个用逗号隔开) 例子: BNB,ARB
- [x] get_crypto_ufutures_price u本位合约[可选,默认BTCUSDT] 例子: ETHBTC

### 2. ChatGPT功能清单

- [x] chatgpt

### 3. VPS库存监控功能清单(已弃用)

- [ ] vps_monitor_supported_list 查看支持监控的网站
- [ ] vps_add_supported_list 添加支持监控的网站 例: url keyword name desc(有空格需要引号)
- [ ] add_vps_monitor url(必须是vps_monitor_supported_list有的,或者系统站点模版的商家)

### 4. 抠图功能

- [x] cutout (需要在配置文件将 photo -> enable 设为true并添加apikey)

### 5. Telegram 信息获取

- [x] chatid

### 6. Cron 定时提醒

- [x] add_cron 每隔多久一次提醒,单位/秒 例: 15 提醒内容(必填)
- [x] delete_cron 删除 例: 1

### 7. 视频下载

- [x] youtube_download 下载ytb视频
- [x] youtube_audio_download 下载ytb音频
- [ ] bilibili_download 下载bilibili视频
- [x] youtube_download_cut 下载ytb的视频并裁剪(需要安装ffmpeg)
- [x] youtube_audio_download_cut 下载ytb音频并裁剪(需要安装ffmpeg)

### 8. 贴纸和GIF下载

- [x] sticker_download 下载贴纸表情
- [x] gif_download 下载GIF


## 三. 环境安装(可选)

__用到视频裁剪功能或者GIF下载功能需要安装 FFmpeg__

`Ubuntu或Debian`
```
sudo apt-get update
sudo apt-get install ffmpeg
```

`CentOS或RHEL`

```
sudo yum install epel-release
sudo yum install ffmpeg
```

`Fedora`

```
sudo dnf install ffmpeg
```

`Arch Linux`

```
sudo pacman -S ffmpeg
```

## 四. Telegram Commands

__通过 @BotFather /setcommands 添加__

```
chatid - 查询chatid
add_kline_strategy_probe - 探测连续3根一直走势的k线 例: btcusdt
delete_kline_strategy_probe - 删除探测 例: btcusdt
list_kline_strategy_probe - 列出当前探测的加密货币
add_cron - 每隔多久一次提醒,单位/秒 例: 15 提醒内容(必填)
delete_cron - 删除 例: 1
add_crypto_growth_monitor - 添加加密货币高线监控 例: BNB 1.1 (单位USD)
add_crypto_decline_monitor - 添加加密货币低线监控 例: BNB 1.1 (单位USD)
get_crypto_price - 查询当前价格(默认查询监控的加密货币) 例 : BNB
delete_crypto_minitor - 删除监控的加密货币 例: BNB,ARB
get_crypto_ufutures_price - 查询当前合约价格 例 : ETHUSDT
chatgpt - chatgpt功能
cutout - 抠图功能
youtube_download - 下载youtube的视频
youtube_audio_download - 下载ytb音频
youtube_download_cut - 下载youtube的视频并裁剪
youtube_audio_download_cut - 下载ytb音频并裁剪
sticker_download - 下载贴纸表情
gif_download - 下载GIF
```

__待实现__

```
bilibili_download - 下载bilibili的视频
```

__弃用__

```
vps_monitor_supported_list - 查看支持监控的网站
add_vps_monitor - 添加VPS库存监控 例: URL(vps_monitor_supported_list里的)
vps_add_supported_list - 添加支持监控的网站 例: url keyword name desc(有空格需要引号)
```