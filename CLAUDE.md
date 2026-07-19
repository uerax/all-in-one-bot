# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- `go run . -c all-in-one-bot.yml` - start the Telegram bot with the local YAML config.
- `go run . -v` - print the current version string.
- `go build -o aio main.go` - build a local binary.
- `go test ./...` - run all tests.
- `go test ./crypto/crocodile -run '^TestParseRuleConfig$' -count=1` - run one test in one package; replace the package and test name as needed.
- `go test ./... -run 'TestName' -count=1` - run matching tests across all packages.
- `go test ./... -run '^$'` - compile all packages and tests without running test functions.
- `gofmt -w <files>` - format edited Go files.
- `go vet ./...` - run the standard Go static checks.
- `docker compose up -d` - run the published Docker image with `./config` mounted to `/etc/aio` and `./logs` mounted to `/var/log/aio`.
- `bash build.sh` - interactive release helper that builds Linux/Windows/macOS binaries and can create/push a tag.

## Runtime configuration

- The Go module is `github.com/uerax/all-in-one-bot`; `go.mod` declares Go 1.22 with toolchain `go1.23.1`.
- `main.go` accepts `-c` for the config path, defaults to `all-in-one-bot.yml`, loads it through `config.Load()`/`goconf`, then starts `tg.Server()`.
- Telegram token and chat ID come from `telegram.token` and `telegram.chatId`; `TG_TOKEN` and `TG_CHATID` environment variables override those values. Docker Compose reads them from `.env`/`.env.example`.
- `main.go` also starts `net/http/pprof` on `localhost:7777`.
- Runtime state and generated files are controlled by config paths, notably `/usr/local/share/aio/` for crypto tracking dumps and `/tmp/aio-tgbot/...` for video/photo/sticker/gif temporary files.
- The Dockerfile does not build from the checked-out source; it downloads the latest GitHub release binary and the default config into the image.
- Do not commit real tokens, chat IDs, or API keys. Keep secrets in `.env` or an untracked local config.

## High-level architecture

This is a Go Telegram bot with domain packages behind a central Telegram adapter.

- `main.go` is intentionally thin: set logging/flags, load config, start the Telegram update loop.
- `tg/` is the orchestration layer. `tg.Server()` reads updates from `go-telegram-bot-api`, filters by `ChatId` when configured, and routes commands.
- Command routing uses two pieces of global state in `tg/server.go`: `Cmd` stores the currently selected multi-step command, and `api` is the process-wide `*Aio`. Command messages either execute immediate actions or set `Cmd` and send a usage tip; the following non-command message is dispatched by the `Cmd` switch.
- `tg/cmd.go` contains thin wrappers between Telegram command names and domain services. When adding or changing a command, usually update both `tg/server.go` (routing/tips) and `tg/cmd.go` (argument parsing and service call), plus the command lists in `all-in-one-bot.yml`/`README.md` if user-visible.
- `tg/aio.go` constructs all domain services in `Aio.NewBot()`. Services share a `chan common.AioEvent`; `Aio.WaitToSend()` consumes events and `dispatch()` sends Telegram text, markdown, photos, videos, documents, and audio.
- `common/event.go` defines the event contract used by feature packages. New services should prefer accepting an optional `chan<- common.AioEvent` (often as a variadic `ch ...chan<- common.AioEvent`) and sending via `common.Send()` instead of depending directly on Telegram APIs.

### Adding a feature (typical path)

1. Implement domain logic in its package (or extend an existing one under `crypto/`, `video/`, etc.).
2. Construct the service in `Aio.NewBot()` if it needs process lifetime / the shared event channel.
3. Wire Telegram entry points in `tg/cmd.go` and route them in `tg/server.go` (immediate command and/or multi-step `Cmd` + tip).
4. Emit user-facing results through `common.AioEvent` helpers (`Text`, `Markdown`, `PhotoTo`, …).
5. If the command is user-visible, update `command:` sections in `all-in-one-bot.yml` and the BotFather/command docs in `README.md`.

## Domain packages

- `crypto/` holds most market and wallet functionality. `NewCrypto()` is a singleton for external crypto APIs and caches pair data. `CryptoMonitor`, `Probe`, `Track`, and `Coingecko` layer polling/monitoring behavior on top of it.
- `crypto/track.go` and `crypto/probe.go` manage long-running goroutines with `context.CancelFunc` maps. They recover/dump state from the configured etherscan path (`tracking.json`, `kline.json`, smart address dumps, pair cache) and use `PollingKeyV2` for rotating Etherscan API keys.
- `crypto/crocodile/` is the volume-spike monitor. It loads a remote list, local `crypto/crocodile/list.json`, or embedded list; rules are configurable at runtime through `crocodile_rule`; chart rendering emits photo events. Prefer `NewCrocodileWithSource` in tests.
- `video/`, `photo/`, `tg/sticker.go`, and `tg/gif.go` handle media downloads/conversions and depend on configured temp directories. README notes that cut/download features may require third-party API keys and FFmpeg.
- `bbs/`, `cron/`, `chatgpt/`, `utils/`, `vps/`, and `lists/` are independent feature domains constructed by `Aio.NewBot()` and called through `tg/cmd.go`.
- Package-local data files are part of behavior: `crypto/list/list.json`, `crypto/crocodile/list.json`, and `bbs/bitcointalk/filter.json` are read by their packages.
- `internal/crypto/` exists but is not the active domain path for current features; prefer `crypto/` unless intentionally working under `internal/`.

## Testing notes

- Existing tests live next to code as `*_test.go` in `tg/`, `crypto/`, and `crypto/crocodile/`.
- Use table-driven or fake-source tests for command formatting, parsing, config-sensitive logic, and monitor rules. Crocodile tests inject a fake `klineSource` and read events from a buffered `common.AioEvent` channel.
- For HTTP-facing code, prefer `httptest.Server` or pure helper tests like the existing CoinGecko tests rather than depending on live third-party APIs.
- Run `go test ./...` before opening a PR.

## Style and repo docs

- Follow normal Go conventions: `gofmt`, tabs, `CamelCase` exports, descriptive feature-oriented file names (`wallet.go`, `monitor.go`, `youtube.go`).
- Recent commit history and `AGENTS.md` use short conventional subjects such as `feat: ...`, `fix: ...`, and `chore: ...`.
- The README is primarily user-facing and Chinese-language; it documents install/update scripts, Docker setup, BotFather command lists, and feature-specific Telegram commands.
- `AGENTS.md` is the contributor-oriented counterpart for package boundaries and local commands. Keep `CLAUDE.md` and `AGENTS.md` in sync when changing command workflows, config shape, Docker behavior, major package responsibilities, or test expectations.
- CI (`.github/workflows/go.yml`) builds multi-platform release binaries on version tags (`v*`); local development relies on the Go toolchain commands above rather than that workflow.
