## 一. all-in-one-bot

__Telegram机器人, 目前支持监控加密货币价格, ChatGPT, 自动抠图, Youtube视频/音频下载和剪切, Telegram贴纸Sticker下载, Telegram的gif图片下载, Bilibili视频下载, Douyin视频下载, 土狗币查询, 通用工具箱(base64,json格式化,时间戳转换)__

## 二. Usage

`安装`

```
bash -c "$(curl -L https://raw.githubusercontent.com/uerax/all-in-one-bot/master/install.sh)" @ install
```

> 注: 配置文件token必须添加,否则会启动失败, chatId不添加的情况下只能执行 /chatid 命令获取chatid, 获取到后添加到配置文件并重启服务(可以通过`其他`脚本输出`8` -> `2`进行添加)
* 在 all-in-one-bot.yml 添加你的 telegram token
`telegram 搜索用户 @BotFather 发送 /newbot 获取`
* 在 all-in-one-bot.yml 添加你的 telegram chatId
`添加token后启动应用,去你的bot发送 /chatid 即可获取`

`更新`

```
bash -c "$(curl -L https://raw.githubusercontent.com/uerax/all-in-one-bot/master/install.sh)" @ update
```

`卸载`

```
bash -c "$(curl -L https://raw.githubusercontent.com/uerax/all-in-one-bot/master/install.sh)" @ uninstall
```

`其他`

```
bash -c "$(curl -L https://raw.githubusercontent.com/uerax/all-in-one-bot/master/install.sh)" @
```

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

## 三. Command List
### 1. 加密货币监控功能清单

- [x] list_wallet_tracking 列出正在追踪的聪明钱包地址
- [x] list_smart_addr_probe 列出正在探测的聪明钱包地址
- [x] wallet_tx_analyze 分析钱包近n条交易的利润 例: 0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E 30
- [x] wallet_tracking 追踪聪明钱包买卖动态 例: 0x7431931094e8BAe1ECAA7D0b57d2284e121F760e
- [x] stop_wallet_tracking 停止追踪聪明钱包买卖动态 例: 0x7431931094e8BAe1ECAA7D0b57d2284e121F760e
- [x] set_smart_addr_probe_itv 修改聪明地址探测频率 例: 15
- [x] dump_smart_addr_probe_list dump聪明地址的过滤合约(建议每次准备重启服务的时候执行一次)
- [x] smart_addr_tx 输入聪明地址(eth)和近n条交易 例: 0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80 20
- [x] smart_addr_probe 监控聪明地址(eth)购买情况 例:  0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80
- [x] delete_smart_addr_probe 输入关闭监控的聪明地址(eth) 例: 0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80
- [x] add_kline_strategy_probe 探测连续3根一直走势的k线 例: btcusdt
- [x] delete_kline_strategy_probe 删除探测 例: btcusdt
- [x] get_meme 获取meme币信息 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 bsc(可选填)
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
- [x] json_format 格式化json

## 四. 环境安装(可选)

* __Telegram 50M上传限制的解决思路__

1. 前往[Guide](https://tdlib.github.io/telegram-bot-api/build.html)根据自己的系统选择参数,根据他提供的命令执行安装 Local Telegram Api
2. 需要先去 https://my.telegram.org ，登录后，点API development tools可以看到你的api-id和api-hash
3. 执行以下命令,用上面的api-id和api-hash替换里面的<arg>
```
telegram-bot-api --api-id=<arg> --api-hash=<arg> --local -l /var/logs/tgserver.log -v 3
```
4. 通过golang执行该命令发送文件
```
curl -v -F chat_id="<chat_id>" -F video="file://<filepath>" -F supports_streaming=true -F caption="<filename>" http://localhost:8081/bot<token>/sendVideo
```

* __用到视频裁剪功能或者GIF下载功能需要安装 FFmpeg__

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

## 五. Telegram Commands

__通过 @BotFather /setcommands 发送添加__

* 由于功能不断添加 command列表过长命令难找,采用分组形式自行查询获取,建议只填加以下常用命令到command列表,有需要其他功能进行查询获取

`常用命令`

```
chatid - 查询chatid
get_meme - 获取meme币信息 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 eth(可选填)
wallet_tx_analyze - 分析钱包近n条交易的利润 例: 0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E 30
wallet_tracking - 追踪聪明钱包买卖动态 例: 0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E
stop_wallet_tracking - 停止追踪聪明钱包买卖动态 例: 0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E
smart_addr_tx - 输入聪明地址和近n条交易 例: 0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80 20
smart_addr_probe - 监控聪明地址购买情况 例:  0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80
set_smart_addr_probe_itv - 修改聪明地址探测频率 例: 15
chatgpt - chatgpt功能
youtube_audio_download_cut - 下载ytb音频并裁剪
cmd_list - 列出全部功能
crypto_cmd_list - 加密货币相关功能列表
video_cmd_list - 音视频下载处理功能列表
image_cmd_list - 图片处理/下载功能列表
utils_cmd_list - 工具类功能列表
list_cmd_list - 功能分类列表
```

`全部命令`

```
chatid - 查询chatid
list_wallet_tracking - 列出正在追踪的聪明钱包地址
list_smart_addr_probe - 列出正在探测的聪明钱包地址
wallet_tx_analyze - 分析钱包近n条交易的利润 例: 0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E 30
wallet_tracking - 追踪聪明钱包买卖动态 例: 0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E
stop_wallet_tracking - 停止追踪聪明钱包买卖动态 例: 0xaA6a1993Ec0BC72dc44B8E18e1DCDeD11A69302E
set_smart_addr_probe_itv - 修改聪明地址探测频率 例: 15
smart_addr_tx - 输入聪明地址(eth)和近n条交易 例: 0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80 50
dump_smart_addr_probe_list - dump聪明地址的过滤合约(建议每次准备重启服务的时候执行一次)
smart_addr_probe - 监控聪明地址(eth)购买情况 例:  0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80
delete_smart_addr_probe - 输入关闭监控的聪明地址(eth) 例: 0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80
add_kline_strategy_probe - 探测连续3根一直走势的k线 例: btcusdt
delete_kline_strategy_probe - 删除探测 例: btcusdt
get_meme - 获取meme币信息 例: 0x6982508145454ce325ddbe47a25d4ec3d2311933 eth(可选填)
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
json_format - 格式化json
youtube_download - 下载youtube的视频
youtube_audio_download - 下载ytb音频
youtube_download_cut - 下载youtube的视频并裁剪
youtube_audio_download_cut - 下载ytb音频并裁剪
bilibili_download - 下载bilibili的视频
douyin_download - 下载douyin的视频
sticker_download - 下载贴纸表情
gif_download - 下载GIF(非贴纸)
cmd_list - 列出全部功能
crypto_cmd_list - 加密货币相关功能列表
video_cmd_list - 音视频下载处理功能列表
image_cmd_list - 图片处理/下载功能列表
utils_cmd_list - 工具类功能列表
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