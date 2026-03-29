# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development commands

- Build all packages: `go build ./...`
- Run the bot with the default config file in repo root: `go run .`
- Run the bot with an explicit config path: `go run . -c /absolute/path/to/all-in-one-bot.yml`
- Print the version string: `go run . -v`
- Download dependencies: `go mod download`
- Run formatting on tracked Go files: `gofmt -w $(git ls-files '*.go')`
- Run tests/compile check: `go test ./...`
- Run tests for a single package: `go test ./tg` or `go test ./crypto`
- Run a single test if tests are added: `go test ./crypto -run TestName`
- Start the packaged container setup: `docker compose up -d`
- Stop the packaged container setup: `docker compose down`
- Follow container logs: `docker compose logs -f aio`
- Build release binaries interactively: `bash build.sh`

## Runtime and config notes

- The binary defaults to `all-in-one-bot.yml`; local development usually runs against that file or a copied variant.
- `TG_TOKEN` and `TG_CHATID` override the Telegram values from config at runtime.
- `docker-compose.yml` runs the published image `uerax/aio:latest` and mounts `./config` to `/etc/aio` plus `./logs` to `/var/log/aio`.
- Wallet tracking and smart-address analysis depend on configured Etherscan keys.
- Media workflows depend on external tools/services: FFmpeg for cutting/downloading media, and optionally a local Telegram Bot API server for larger uploads.
- `main.go` also starts a pprof server on `localhost:7777`.

## Architecture overview

### Entry and wiring

- `main.go` is minimal: parse flags, load YAML config through `goconf`, start pprof, then call `tg.Server()`.
- `config/config.go` is intentionally thin. Most packages read configuration directly with `goconf.Var*` calls instead of using a shared typed config object.
- `tg/aio.go` is the central wiring point. It constructs the `Aio` container and initializes all subsystems: crypto monitors, wallet tracking, ChatGPT, cron, media download, image processing, utility commands, RSS/BBS monitors, and list/help responders.

### Telegram control flow

- `tg/server.go` is the core update loop. It receives Telegram updates, enforces the configured `ChatId` gate, and routes either command messages or follow-up plain-text payloads.
- Interactive command flow is stateful via the package-level `tg.Cmd` variable. A command often sets `Cmd`, sends a usage hint, then interprets the next non-command message as that command’s arguments.
- `tg/cmd.go` contains the adapter layer between Telegram commands and domain services. Most functions just parse/default arguments and launch subsystem work asynchronously with goroutines.

### Async delivery model

- Subsystems communicate back to Telegram primarily through channels rather than returning values directly.
- `Aio.WaitToSend()` in `tg/aio.go` is the fan-in loop that listens to all subsystem channels and turns them into Telegram messages, files, images, audio, or videos.
- This design means most features are long-lived background services hanging off one bot process rather than request/response handlers.

### Main domain areas

- `crypto/` is the main business domain. It mixes:
  - exchange/token data clients (`crypto.go`, `coingecko.go`, `kline.go`, `meme.go`)
  - monitoring and probing (`monitor.go`, `probe.go`, `polling.go`)
  - wallet/smart-address tracking (`track.go`, `trackV2.go`, `txs.go`)
- `crypto/crypto.go` is the shared external-data client for Binance, Dexscreener, GoPlus, Honeypot, DexTools, Etherscan-derived endpoints, and pair-cache persistence.
- `crypto/monitor.go` manages user-specific high/low price alerts in memory and periodically polls prices, emitting notifications through a channel.
- The wallet-analysis and “smart money” workflows described in `README.md` are implemented across `tg/server.go`, `tg/cmd.go`, and the tracking/probe files in `crypto/`.

### Other subsystems

- `video/`, `photo/`, `bbs/`, `vps/`, `utils/`, `lists/`, plus the GIF/sticker handlers under `tg/`, all follow the same pattern: initialize a subsystem in `Aio`, expose a small API, then emit results through channels.
- The repo is organized more around bot capabilities than around layered MVC-style boundaries.

### Persistence and state assumptions

- Most runtime state is in memory. Restarting the bot can clear active monitoring/tracking state unless a subsystem explicitly dumps it.
- One notable persisted cache is the crypto pairs map: `crypto/crypto.go` periodically writes `pairs_dump.json` under the configured `crypto.etherscan.path` (default `/usr/local/share/aio/`).
- Because `tg.Cmd` is global rather than per-chat, command flow changes should assume the bot is effectively operating for one configured chat at a time.

## Repo-specific guidance from existing docs

- `README.md` is important for operational context: Docker Compose is the recommended deployment path, and many user-facing bot commands and setup requirements are documented there.
- The checked-in `all-in-one-bot.yml` is the best source of truth for config shape, required keys, temp-path expectations, and optional integrations.
- There is no existing `CLAUDE.md`, no `.cursorrules`, no `.cursor/rules/`, and no `.github/copilot-instructions.md` in this repository.
- `build.sh` is release-oriented rather than a normal dev build script: it cross-compiles binaries for multiple OS/arch targets and can optionally create and push a git tag.
