# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# 项目协作规则与开发指南

> 本文件在每次 session 启动时都会被加载，请严格遵守以下规则。
> 规则冲突时，按编号从小到大优先级递减，第 0 条优先级最高。

---

## 0. 自我保护规则（最高优先级）

- **禁止修改本文件（CLAUDE.md）**，除非用户在当前对话中明确说"帮我改 CLAUDE.md"或"更新一下项目规则"。
- 发现本文件内容与实际情况不符（比如规则过时）时，**不要自行修改**，而是提醒用户："CLAUDE.md 里的 XX 规则似乎需要更新，需要我改吗？"
- 批量任务（重构、格式化、批量替换等）**默认排除** `CLAUDE.md`、`.claude/` 目录下所有文件，除非用户明确点名要改这些文件。

---

## 1. 状态持久化规则（防止 context 压缩丢细节）

- 项目根目录维护 `.claude/STATE.md`，记录当前任务的完整状态；清单类任务额外维护 `.claude/BUGS.md`。
- 每完成一个有意义的子任务（修复一个 bug、完成一个功能点），**立即**在 `.claude/STATE.md` 追加记录，格式：

  ```
  ## [日期] 任务名
  - 状态：已完成 / 进行中 / 待处理
  - 优先级：P0-P5
  - 描述：xxx
  - 涉及文件：path/to/file
  - 备注：xxx
  ```

- 每次开始新任务前，**先读取 `.claude/STATE.md`**，以此为准，不要仅依赖对话历史中的记忆。
- 一次性列出多个待办事项（bug 清单、任务清单）时，**先写入 `.claude/BUGS.md`**，再逐条处理；每处理完一项就同步更新状态，不要等全部做完再一次性更新。
- `.claude/BUGS.md` 只记录**当前待处理 / 进行中**的条目，格式极简（状态、优先级、一句话描述）。
- 某一项处理完成后，**从 BUGS.md 中直接移除该条目**（而不是标记"已完成"后继续保留），完整细节（涉及文件、备注、决策过程）**只写入 STATE.md 一次**。
- BUGS.md 相当于"进行时"的待办板，STATE.md 相当于"过去时"的历史记录，同一个事项在任一时刻只应该出现在其中一个文件里，不重复维护状态字段。

---

## 2. 上下文管理

- 感知到对话已经很长（用户提到"上下文快满了""context 80%"等）时，主动提醒：先把当前状态写入文件，再考虑开新 session 或执行 `/compact`。
- 不假设自己"记得"很久之前对话中的细节；涉及关键决策、清单类信息时，优先以文件内容为准，而不是对话记忆。
- `.claude/STATE.md` 和 `.claude/BUGS.md` 只保留**当前进行中 + 最近已完成**的任务记录；超过一定量（如已完成任务累计超过 20 条，或体积明显偏大）时，将较早的"已完成"记录整体移入 `.claude/STATE-ARCHIVE.md`，仅在需要追溯历史时读取归档文件。

---

## 3. 代码修改规范

- 修改代码前，先说明将要做的改动范围，避免大范围无关改动。
- 每次修改后，运行**相关**测试确认没有引入新问题（不必每次跑全量测试套件，除非用户要求或改动涉及核心模块）。
- 不确定的地方，先提问或说明假设，不要臆造需求。
- 涉及不确定的技术判断时，在回复中明确说明"这是推测"还是"这是确认过的事实"，尤其是对"某个方案能不能解决问题"这类结论性判断，先去读实际代码验证，不要凭经验直接下结论。

---

## 4. 运行进程与系统资源释放

- **禁止主动启动任何长期运行的后台进程或机器人实例**（如后台运行 `go run .`、带 `&` 的常驻进程等），无论是否为了"验证改动"。
- 修改代码后，直接说明改了什么、建议手动测试哪些命令或流程即可；由用户自行在终端中启动运行。
- 需要验证逻辑但不涉及启动长期服务的场景（如运行单元测试 `go test`、编译检查 `go build`、代码格式化 `gofmt`、静态检查 `go vet`）不受此条限制。
- **浏览器自动化工具（Chrome DevTools MCP / Playwright 等）验证后必须立即关闭标签页（资源释放铁律）**：
  - 在用户明确要求或允许使用浏览器工具进行网页检查或网络探测时，**验证结束后必须立即主动关闭相关测试标签页**，严禁留置运行中页面在后台持续消耗 CPU、GPU 与网络带宽。
  - 若测试页面为当前唯一标签页、浏览器不允许关闭最后一个标签页，需先打开一个空白页再关闭业务测试页面，确保浏览器回到干净、零占用的待机状态。

---

## 5. Git 提交与版本管理规范

- commit message 使用简洁的祈使句，格式：`<类型>: <说明>`，类型包括 `feat` / `fix` / `refactor` / `docs` / `test` / `chore` / `ci`。
- **版本规范（针对本 Go 项目 Release 机制）**：
  - 本项目通过 GitHub Actions Tag 触发多平台编译打包（`v*`，如 `v2.1.1`）。
  - **纯文档/规则修改豁免**：若本次修改仅涉及文档、项目规范、清单文件（如 `CLAUDE.md`、`README.md`、`.claude/` 等）而没有实际 Go 代码改动，**严禁递增版本号或建议打新 Tag**，明确说明"纯文档修改，版本保持不变"。
  - **代码修改版本递增建议**：包含实际功能或缺陷修复代码改动时，根据修改幅度评估语义化版本建议：
    - **Patch（修订号）**：缺陷修复（`fix`）、内部小重构、非破坏性微调；
    - **Minor（次版本号）**：新增功能模块（`feat`）、重大架构优化；
    - **Major（主版本号）**：破坏性变更、底层核心库/协议颠覆性升级。
    在回复中指出建议的 Release Tag 升级方向，供用户发布时参考。
- **不自动执行 `git commit`**，除非用户在当前任务中明确要求（如"提交一下""commit 一下"）。默认只汇报修改内容与建议 commit message。
- **不自动执行 `git push`**，除非用户明确要求。
- **严禁自动执行涉及历史重写的操作**（`rebase -i`、`reset --hard`、`push --force`），除非用户明确要求。

---

## 6. 沟通风格

- 回复中直接给出结论和改动内容，避免不必要的寒暄。
- 涉及不确定的技术判断时，明确说明"这是推测"还是"这是确认过的事实"。

---

## 7. 领域专项规则（本项目核心架构红线）

- **Telebot 事件循环与异步分发铁律**：
  - Telebot 的 `Handle` 回调函数中**严禁执行任何阻塞式 I/O、外部网络请求或长时任务**。必须通过 `go svc.Method(...)` 开启独立 goroutine 处理业务并立即 `return nil`，保持 Telebot 事件轮询循环畅通。
  - 异步任务结果统一打包为 `models.Message` 发送至全局缓冲通道 `r.msgCh`（`chan models.Message`），由 `internal/bot/dispatcher.go` 集中异步推送到 Telegram，严禁在业务 goroutine 中直接调用 bot 阻塞发送。
- **命令鉴权与静默安全规范**：
  - 新增命令必须在 `internal/router/router.go` 注册并通过 `authorizedOnly` 中间件包裹。
  - 除公共命令 `/chatid` 无条件放行外，一旦配置了 `TELEGRAM_ADMIN_IDS`，非白名单用户的请求必须在中间件层记录 Warn 日志后**直接静默丢弃（return nil）**，严禁向未授权用户发送任何回复或报错信息。
- **Crypto Provider 多源解耦规范**：
  - 业务层（`coin`、`crocodile`）必须统一依赖 `internal/crypto/provider.Manager`，严禁在业务 Handler 中绕过 Manager 硬编码任何单一第三方 API 客户端。
  - **智能地址路由规则**：
    - EVM 合约（`0x[0-9a-fA-F]{40}`）、Solana Base58 地址或带 `chain:` 前缀的查询，必须直接路由至 DEX 引擎（GeckoTerminal）。
    - 标准 Ticker / Token Symbol 优先查询 CEX（CoinGecko），未命中时自动平滑回退 DEX。
    - 搜索操作并发检索 CEX 与 DEX 并合并输出。
    - `Crocodile` K 线扫描必须优先获取 GeckoTerminal 1d OHLCV 真实成交量，接口失效或无数据时降级至 CoinGecko。
- **并发安全与存储隔离（`internal/store/`）**：
  - 内存共享数据（如 RSS 过滤集、监控活跃状态、LRU 缓存）必须严格使用 `sync.RWMutex` 保护，防止后台定时轮询与用户命令并发读写产生数据竞争。
  - `FileStore` 在远程 URL 模式下写操作会重定向到 `/tmp/aio/data`，新增持久化数据结构必须确保 JSON 序列化兼容性。

---

## 8. 功能实现索引维护规则

- 项目维护 `.claude/FEATURE-MAP.md`，记录"功能模块 → 核心文件路径"的快速索引，用于减少大范围代码搜索。此文件**仅供 Claude 导航使用**，不作为团队产品文档维护，保持精简、路径为主。
- 索引只做索引的事：只记录"模块 → 文件路径"和一句话功能描述，不记录任务状态、处理进度、修改历史；进度一律写入 `.claude/STATE.md`。
- **触发更新的时机**：
  - 新增功能模块、迁移/重命名关键文件时，同步更新对应条目。
  - 完成一个子任务、写 `.claude/STATE.md` 时，检查是否涉及索引变更并一并同步。
- 每次接手涉及"定位某功能实现在哪"的任务时，**先读取 `.claude/FEATURE-MAP.md`**；缺失或对不上时再全项目搜索并顺手补全。

---

## 待办 / 索引类文件一览

- 当前 bug 清单：`.claude/BUGS.md`
- 项目状态快照：`.claude/STATE.md`
- 状态归档（历史已完成任务）：`.claude/STATE-ARCHIVE.md`
- 功能实现索引（仅供 Claude 导航）：`.claude/FEATURE-MAP.md`

---

## 常用开发命令

- 构建所有包：`go build ./...`
- 编译可执行二进制：`go build -o all-in-one-bot .`
- 启动 Bot（纯环境变量）：`go run .`
- 启动 Bot（指定 YAML 配置）：`go run . -config config.yaml`
- 下载并整理依赖：`go mod download && go mod tidy`
- 格式化 Go 代码：`gofmt -w $(git ls-files '*.go')`
- 静态分析：`go vet ./...`
- 运行全部测试：`go test ./...`
- 运行指定包测试：`go test -v ./internal/handler/bbs/bitcointalk`
- 运行单个测试函数：`go test -v -run TestSearch ./internal/crypto/provider/...`
- 竞态检测（需 CGO 及 GCC 工具链）：`CGO_ENABLED=1 go test -race ./...`

---

## 核心架构概览

### 1. 入口与依赖装配 (`main.go`, `internal/app/run.go`)
- `main.go` 极简调用 `app.Run()`。
- `Run()` 负责：
  1. 初始化结构化日志 `logger.NewLogger()` (`log/slog`)。
  2. 加载配置 `config.LoadConfig(*configPath)`。
  3. 建立 Telebot 实例 `telegram.NewBot(cfg.Telegram)`。
  4. 初始化持久化存储 `store.NewStore(cfg.Database, log)`。
  5. 组装 `router.Dependencies`（Config, Logger, Store, AdminIDs 白名单 map）。
  6. 创建容量为 20 的缓冲消息通道 `c := make(chan models.Message, 20)` 并启动后台 `bot.Dispatcher`。
  7. 注册所有 Handler 并包裹鉴权中间件。
  8. 调用 `b.Start()` 阻塞轮询。

### 2. 配置加载层级 (`internal/config/`)
- 优先级：`环境变量 (ENV / .env via godotenv) > YAML 文件 (-config 参数) > 代码内置默认值`。
- `TELEGRAM_ADMIN_IDS`（逗号分隔或 YAML 列表）由 `router.AdminIDsSet` 解析为 `map[int64]bool`。

### 3. 命令路由与鉴权 (`internal/router/`)
- 命令接口：`Handler` (`Cmd() string`, `Handle(c tb.Context) error`)。
- 所有 Handler 在注册时统一包裹 `authorizedOnly` 中间件：
  - `AdminIDs` 为空时全量放行。
  - `AdminIDs` 非空时，仅在白名单内的 senderID 可执行，其他静默丢弃。
  - `/chatid` 始终放行。

### 4. 异步消息模型 (`internal/bot/dispatcher.go`, `internal/models/`)
- Handler 开 goroutine 处理逻辑，将回复发送至 `chan models.Message`。
- `Dispatcher` 监听通道并向 Telegram 调用 `bot.Send`（支持 `KindText` 与 `KindMarkdown`，关闭链接预览）。

### 5. 加密货币行情架构 (`internal/crypto/provider/`)
- 接口抽象：`PriceProvider`, `CEXSearchProvider`, `DEXSearchProvider`, `KlineProvider`, `TrendingProvider`。
- 实现引擎：`coingecko`（CEX 现货）与 `geckoterminal`（链上 DEX 池子）。
- `Manager` 智能调度：自动识别 EVM / Solana 合约地址走 DEX，普通代码走 CEX，并发搜索，Crocodile K 线优先取 GeckoTerminal DEX 成交量。

### 6. 领域业务模块 (`internal/handler/`)
- `crypto/coin`: 单币/池子查询 (`/coin`)、多源搜索 (`/coin_search`)、持仓估值 (`/coin_price`)、DEX 热门榜 (`/coin_trending`)、池子详情 (`/coin_pool`)、持仓播报监控 (`/coin_monitor`, `/coin_stop`)。
- `crypto/crocodile`: 每日 UTC 00:05 量能异动扫描 (`/crocodile_check`, `/crocodile_monitor`, `/crocodile_stop`, `/crocodile_list`, `/crocodile_add`, `/crocodile_rule`)。
- `polymarket`: 预测市场持仓查询 (`/polymarket_holdings`)、L2 凭据检测 (`/polymarket_l2_check`)。
- `bbs/bitcointalk` & `bbs/nodeseek`: 论坛新帖关键词/RSS 监控。
- `telegram`: 工具命令 (`/chatid`)。

### 7. 数据持久化 (`internal/store/`)
- `Store` 接口 (`Load`, `Save`, `Set`)。
- `FileStore` 支持读取本地文件路径与 GitHub Raw 等网络 URL，自动 fallback `.json` 与 `.dat`；URL 模式下写入重定向至 `/tmp/aio/data`。
