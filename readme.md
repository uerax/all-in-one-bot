# all-in-one-bot (lite)

轻量级、高性能的多功能 Telegram 机器人，基于 Go 语言（`gopkg.in/telebot.v4`）与模块化 `Handler` / `Provider` 架构重构。内置多源加密货币行情引擎（GeckoTerminal DEX + CoinGecko CEX）、Crocodile 鳄鱼量能突增监控、Polymarket 预测市场查询、RSS 社区监测及实用工具箱。

---

## 目录

- [一、 核心功能：Crocodile 鳄鱼量能突增监控引擎](#一-核心功能crocodile-鳄鱼量能突增监控引擎)
- [二、 核心功能：Coin 基础与 DEX 行情工具](#二-核心功能coin-基础与-dex-行情工具)
- [三、 其他功能与命令（Polymarket / RSS / 工具箱）](#三-其他功能与命令)
- [四、 快速部署与配置指南](#四-快速部署与配置指南)
  - [1. 直接部署（systemd / 二进制）](#1-直接部署systemd--二进制)
  - [2. Docker Compose 部署](#2-docker-compose-部署)
- [五、 配置文件与环境变量说明](#五-配置文件与环境变量说明)

---

## 一、 核心功能：Crocodile 鳄鱼量能突增监控引擎

**Crocodile（鳄鱼引擎）** 是专为加密货币（CEX 现货代币 & 链上 DEX 池子/Meme 币）设计的每日 **UTC 00:05** 量能异动自动检测系统。

### 1.1 工作原理与计算机制
- **数据源驱动**：优先从 **GeckoTerminal DEX API** 获取链上交易对的精确 1day OHLCV K 线（极速捕捉 Solana、Ethereum、Base、BSC 等链上代币的真实成交量）；未找到时自动平滑降级至 **CoinGecko CEX** 数据源。
- **UTC 00:00 自然日对齐与未闭合柱子剔除**：
  - 在每日 **UTC 00:05** 触发扫描时，自动过滤刚开盘 5 分钟的当天柱子，精准锁定 **上一个刚刚闭合的完整 24 小时 UTC 自然日 K 线**（`targetDay`）。
- **判定阈值算法**：
  - `yRatio`（对比昨日）：`最新闭合日成交量 / 前一天成交量 >= 昨日倍数`（默认 `3.0x`）。
  - `aRatio`（对比均值）：`最新闭合日成交量 / 过去 N 天平均成交量 >= 均值倍数`（默认 `2.0x`）。
  - **两者同时满足** 时立即触发量能信号告警！

### 1.2 Crocodile 命令列表

| 命令 | 功能说明 | 适用场景 / 返回说明 |
| :--- | :--- | :--- |
| **`/crocodile_check`** | **只读诊断扫描** | 手动即时触发，输出**完整诊断报告**：<br>• 扫描 UTC 时间与规则配置<br>• 🟢 **触发告警币种**：收盘价、量能、昨日倍数及均值倍数<br>• ⚪ **未触发币种现状**：展示全部监控币种的最新量能倍数<br>• ⚠️ **异常说明**：接口报错或无数据详情 |
| **`/crocodile_monitor`** | **24h 自动化后台静默监控** | 开启每日 UTC 00:05 自动扫描。未触发信号时**静默无打扰**；触发信号时经当日去重后推送告警消息 |
| **`/crocodile_stop`** | **关闭后台监控** | 停止 24h 自动定时任务 |
| **`/crocodile_list`** | **查看监控列表** | 列出当前正在监控的所有币种、合约或池子 |
| **`/crocodile_add <id> [名称]`** | **添加监控目标** | 支持填入 CoinGecko ID（如 `solana`）或链上合约/池子地址（如 `solana:JUPyi...` 或 `eth:0x6982...`） |
| **`/crocodile_rule`** | **查看当前阈值规则** | 展示回望天数 `Lookback`、昨日倍数 `YesterdayMultiple` 和均值倍数 `AverageMultiple` |
| **`/crocodile_update <回望> <昨日倍数> <均值倍数>`** | **动态修改阈值规则** | 例如发送 `/crocodile_update 5 3.0 2.0` |

---

## 二、 核心功能：Coin 基础与 DEX 行情工具

`coin` 模块采用解耦的 **Provider 接口架构**，集成 **CoinGecko (CEX)** 与 **GeckoTerminal (DEX)** 引擎，支持传统主流币与链上 Meme 币、合约地址的实时查询。

### 2.1 核心特性
- **智能地址路由**：输入 EVM 十六进制合约（`0x...`）、Solana Base58 地址或 `chain:address` 时，自动定位至 GeckoTerminal DEX 池子；输入 Token Symbol/名称时自动查 CEX 市场并回退 DEX。
- **高精度价格与量能显示**：根据价格量级自动动态选择小数位（高价币如 `$1234.56`，微价币如 `$0.00000854`）；大额流动性与成交量自动缩写为 `$1.25B` / `$45.20M` / `$850.50K`。

### 2.2 Coin 命令列表

| 命令 | 功能说明 | 示例 / 返回内容 |
| :--- | :--- | :--- |
| **`/coin <symbol\|address>`** | **单币 / DEX 池子快捷查询** | `/coin BTC`<br>`/coin 0x6982508145454ce325ddbe47a25d4ec3d2311933`<br>`/coin solana:JUPyiwrYF...`<br>返回：价格、5m/1h/6h/24h 涨跌幅、24h 成交量、流动性 Reserve USD、FDV、市值、DEX 平台及网络 |
| **`/coin_search <名称\|地址>`** | **CEX + DEX 混合搜索** | `/coin_search PEPE`<br>同时列出 CoinGecko 市值排名与 GeckoTerminal 链上流动性池 |
| **`/coin_price`** | **持仓价值实时计算** | 读取 `coingecko/list.json` 配置，计算当前总持有 USD 价值及各代币持仓明细（兼容 DEX 合约持仓） |
| **`/coin_trending [网络]`** | **DEX 热门池子榜单** | `/coin_trending`<br>`/coin_trending solana`<br>获取实时 DEX 热门交易池及 24h 涨跌幅/流动性 |
| **`/coin_pool [网络] <池子地址>`** | **指定 DEX 池子数据分析** | `/coin_pool eth 0x11950d141ecb863f010075d381097c119232698d` |
| **`/coin_monitor`** | **持仓 24h 定时播报** | 启动持仓价值的每日定时播报 |
| **`/coin_stop`** | **停止持仓定时播报** | 停止持仓定时任务 |

---

## 三、 其他功能与命令

### 3.1 Polymarket 预测市场
- `/polymarket_holdings <地址>`：查询指定钱包的 Polymarket 持仓与未结算订单。
- `/polymarket_l2_check`: 检查系统 Polymarket L2 凭据配置状态。

### 3.2 社区 RSS 监测
- `/bitcointalk_rss` / `/bitcointalk_rss_stop`: Bitcointalk 论坛新帖关键词订阅与监控。
- `/nodeseek_rss` / `/nodeseek_rss_stop`: NodeSeek 论坛 RSS 监测。

### 3.3 系统工具
- `/chatid`: 查询当前 Telegram 会话/群组的 ChatID。

---

## 四、 快速部署与配置指南

### 1. 直接部署（systemd / 二进制）

```bash
# 1. 编译二进制
go build -o all-in-one-bot .

# 2. 复制配置示例
cp config.example.yaml /etc/all-in-one-bot/config.yaml
vim /etc/all-in-one-bot/config.yaml    # 填写 Telegram Token 与相关配置
```

编写服务文件 `/etc/systemd/system/all-in-one-bot.service`：
```ini
[Unit]
Description=all-in-one-bot (lite)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/opt/aio/all-in-one-bot -config /etc/all-in-one-bot/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now all-in-one-bot
```

### 2. Docker Compose 部署

```bash
# 1. 拷贝环境变量配置文件
cp .env.example .env
vim .env    # 填入 TELEGRAM_TOKEN, TELEGRAM_ADMIN_IDS 等

# 2. 启动服务
docker compose up -d
```

---

## 五、 配置文件与环境变量说明

配置加载优先级为：`环境变量 (ENV / .env) > YAML 配置文件 (-config) > 内置默认值`。

```yaml
# ───────── Telegram ─────────
telegram:
  token: ""             # env TELEGRAM_TOKEN (必填)
  timeout: 10           # env TELEGRAM_TIMEOUT (秒, 默认 10)
  admin_ids: []         # env TELEGRAM_ADMIN_IDS (管理员 ID 白名单，逗号分隔)

# ───────── CoinGecko ─────────
coingecko:
  keys: ""              # env COINGECKO_KEYS (逗号分隔的 Demo API Keys)

# ───────── GeckoTerminal (DEX 行情接口) ─────────
geckoterminal:
  base_url: "https://api.geckoterminal.com/api/v2"  # env GECKOTERMINAL_BASE_URL
  timeout: 15                                       # env GECKOTERMINAL_TIMEOUT (秒, 默认 15)

# ───────── Crocodile (每日 UTC 00:05 量能突增扫描) ─────────
crocodile:
  lookback: 5           # env CROCODILE_LOOKBACK (回看天数, 默认 5)
  yesterday_multiple: 3.0 # env CROCODILE_YESTERDAY_MULTIPLE (相对昨日倍数, 默认 3.0)
  average_multiple: 2.0   # env CROCODILE_AVERAGE_MULTIPLE (相对均值倍数, 默认 2.0)
  interval: 86400       # env CROCODILE_INTERVAL (扫描间隔秒, 默认 86400)
```

---

### 💰 赞助商与特别致谢

[![yxvm_support.png](https://s2.loli.net/2025/04/09/JMyQZUKY2bX4G3q.png)](https://yxvm.com/)

[NodeSupport](https://github.com/NodeSeekDev/NodeSupport) 赞助了本项目。
