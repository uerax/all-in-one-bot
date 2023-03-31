## Usage

在 all-in-one-bot.yml 添加你的 telegram token(从telegram的 @BotFather /newbot 获取)

## 加密货币监控功能清单

__所有单位均为USDT__

1. add_crypto_growth_monitor 加密货币 提示价格 例: BNB 1110
2. add_crypto_decline_monitor 加密货币 提示价格 例: BNB 1110
3. get_crypto_price 加密货币[可选]
4. delete_crypto_minitor 加密货币(多个用逗号隔开) 例子: BNB,ARB

5. chatgpt msg (如果在配置文件把chat设为true, 可以省略/chatgpt直接发送信息进行交互)


6. vps_monitor_supported_list 查看支持监控的网站 TODO
7. add_vps_monitor vps地址(必须是vps_monitor_supported_list有的,或者系统站点模版的商家) TODO
## Telegram Commands
```
add_crypto_growth_monitor  - 添加加密货币高线监控 例: BNB 1.1 (单位USD)
add_crypto_decline_monitor  - 添加加密货币低线监控 例: BNB 1.1 (单位USD)
get_crypto_price - 查询当前价格(默认查询监控的加密货币) 例 : BNB
delete_crypto_minitor - 删除监控的加密货币 例: BNB,ARB
chatgpt - chatgpt功能
vps_monitor_supported_list - 查看支持监控的网站
add_vps_monitor - 添加VPS库存监控
```