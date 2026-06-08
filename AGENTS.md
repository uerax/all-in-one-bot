# Repository Guidelines

## Project Structure & Module Organization
`main.go` is the entrypoint and loads `all-in-one-bot.yml` via `-c`. Feature code is split by domain: `tg/` for Telegram command handling, `crypto/` for market and wallet tracking, `video/` for download helpers, `photo/` for cutout tools, `bbs/` for forum watchers, `chatgpt/`, `cron/`, `vps/`, `utils/`, plus shared code in `common/` and `config/`. Keep package-specific assets close to their package, for example `crypto/list/list.json` and `bbs/bitcointalk/filter.json`.

## Build, Test, and Development Commands
Use the standard Go toolchain first:

- `go run . -c all-in-one-bot.yml` starts the bot with the local YAML config.
- `go build -o aio main.go` builds a local binary.
- `go test ./...` runs all package tests; add tests for new logic even though the repository currently has no committed `*_test.go` files.
- `docker compose up -d` runs the published container and mounts `./config` and `./logs`.
- `bash build.sh` creates release binaries for multiple platforms and can tag a release.

## Coding Style & Naming Conventions
Follow normal Go conventions: format with `gofmt -w`, keep imports clean, and use tabs for indentation. Preserve the existing package layout and keep files focused on one domain. Exported identifiers use `CamelCase`; unexported helpers use `camelCase`. Prefer descriptive file names such as `wallet.go`, `monitor.go`, or `youtube.go` that match the feature area.

## Testing Guidelines
Place tests next to the code they cover and name them `*_test.go`. Prefer table-driven tests for command parsing, config loading, and utility conversions. For integrations that depend on Telegram or third-party APIs, isolate pure logic and test that directly. Run `go test ./...` before opening a PR.

## Commit & Pull Request Guidelines
Recent history uses short conventional subjects like `feat: ...`, `fix: ...`, and `chore: ...`; keep that format and write imperative summaries. PRs should describe behavior changes, note config or API-key impacts, and list the commands you ran to verify the change. Include screenshots or chat transcripts when changing Telegram command flows or other user-visible outputs.

## Configuration & Secrets
Do not commit real tokens, chat IDs, or API keys. Use `.env.example` for `TG_TOKEN` and `TG_CHATID`, and keep local secrets in `.env` or an untracked config file. If you change config shape, update both `all-in-one-bot.yml` and the relevant README or Docker usage notes.
