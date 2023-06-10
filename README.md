## 一. Usage

`安装`

```
bash -c "$(curl -L https://raw.githubusercontent.com/uerax/all-in-one-bot/master/install.sh)" @ install
```

`卸载`

```
bash -c "$(curl -L https://raw.githubusercontent.com/uerax/all-in-one-bot/master/install.sh)" @ uninstall
```

* 在 all-in-one-bot.yml 添加你的 telegram token
`telegram 搜索用户 @BotFather 发送 /newbot 获取`
* 在 all-in-one-bot.yml 添加你的 telegram chatId
`添加token后启动应用,去你的bot发送 /chatid 即可获取`

`操作`

```
// 启动
systemctl start aio
// 关闭
systemctl stop aio
// 自动启动
systemctl enable aio
// 状态
systemctl status aio
```

## 二. Command List
### 1. 加密货币监控功能清单

__单位默认USDT,可在配置文件crypto -> unit修改__

- [x] add_kline_strategy_probe 探测连续3根一直走势的k线 例: btcusdt
- [x] delete_kline_strategy_probe 删除探测 例: btcusdt
- [x] get_meme 获取meme币信息 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 bsc(可选,默认eth)
- [x] add_meme_growth_monitor 添加加meme币高线监控 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 0.00000123 (单位USD)
- [x] add_meme_decline_monitor 添加加meme币低线监控 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 0.0000012 (单位USD)
- [x] meme_monitor_list 列出当前探测的meme币
- [x] delete_meme_monitor 删除meme币监控 例子: 0x6982508145454ce325ddbe47a25d4ec3d2311933 eth
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

- [x] cutout (需要在配置文件添加apikey)

### 5. Telegram 信息获取

- [x] chatid

### 6. Cron 定时提醒

- [x] add_cron 每隔多久一次提醒,单位/秒 例: 15 提醒内容(必填)
- [x] delete_cron 删除 例: 1

### 7. 视频下载

- [x] youtube_download 下载ytb视频
- [x] youtube_audio_download 下载ytb音频
- [x] bilibili_download 下载bilibili视频
- [x] youtube_download_cut 下载ytb的视频并裁剪(需要安装ffmpeg)
- [x] youtube_audio_download_cut 下载ytb音频并裁剪(需要安装ffmpeg)
- [ ] twitter_download 下载twitter的视频
- [x] douyin_download 下载douyin的视频

### 8. 贴纸和GIF下载

- [x] sticker_download 下载贴纸表情
- [x] gif_download 下载GIF(非贴纸)

### 9. 工具箱

- [x] base64_encode 进行base64加密
- [x] base64_decode 进行base64解密
- [x] timestamp_convert 时间戳转换为时间"2006-01-02 15:04:05"
- [x] time_convert 时间转换为时间戳"2006-01-02 15:04:05"

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
get_meme - 获取meme币信息 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 eth(可选,默认eth)
add_meme_growth_monitor - 添加加meme币高线监控 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 eth 0.00000123 (单位USD)
add_meme_decline_monitor - 添加加meme币低线监控 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 bsc 0.0000012 (单位USD)
meme_monitor_list - 列出当前探测的meme币
delete_meme_monitor - 删除meme币监控 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 eth
list_kline_strategy_probe - 列出当前探测的加密货币
add_crypto_growth_monitor - 添加加密货币高线监控 例: BNB 1.1 (单位USD)
add_crypto_decline_monitor - 添加加密货币低线监控 例: BNB 1.1 (单位USD)
get_crypto_price - 查询当前价格(默认查询监控的加密货币) 例 : BNB
delete_crypto_minitor - 删除监控的加密货币 例: BNB,ARB
get_crypto_ufutures_price - 查询当前合约价格 例 : ETHUSDT
add_cron - 每隔多久一次提醒,单位/秒 例: 15 提醒内容(必填)
delete_cron - 删除 例: 1
chatgpt - chatgpt功能
cutout - 抠图功能
base64_encode - 进行base64加密
base64_decode - 进行base64解密
timestamp_convert - 时间戳转换为时间"2006-01-02 15:04:05"
time_convert - 时间转换为时间戳"2006-01-02 15:04:05"
youtube_download - 下载youtube的视频
youtube_audio_download - 下载ytb音频
youtube_download_cut - 下载youtube的视频并裁剪
youtube_audio_download_cut - 下载ytb音频并裁剪
bilibili_download - 下载bilibili的视频
douyin_download - 下载douyin的视频
sticker_download - 下载贴纸表情
gif_download - 下载GIF(非贴纸)
```

__待实现__

```
twitter_download - 下载twitter的视频

```

__弃用__

```
vps_monitor_supported_list - 查看支持监控的网站
add_vps_monitor - 添加VPS库存监控 例: URL(vps_monitor_supported_list里的)
vps_add_supported_list - 添加支持监控的网站 例: url keyword name desc(有空格需要引号)
```