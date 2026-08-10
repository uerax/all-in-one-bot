# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

The `lite` branch is a lightweight refactored version of the Telegram bot using `gopkg.in/telebot.v4`. It replaces the monolithic `tg/` switch-based router and YAML configuration with a clean `Handler` interface registry and environment variable (`godotenv`) configuration.

## Development commands

- Build all packages: `go build ./...`
- Run the bot: `go run .`
- Download dependencies: `go mod download`
- Run formatting on tracked Go files: `gofmt -w $(git ls-files '*.go')`
- Run tests: `go test ./...`
- Run race detector: `go test -race ./...`
- Run tests for a single package: `go test ./internal/handler/bbs/bitcointalk`

## Architecture overview

### Entry and wiring

- `main.go` is minimal: calls `app.Run()`.
- `internal/app/run.go`: Loads environment variables via `.env` / `godotenv`, initializes logger (`slog`), bot instance, stores, dependency container, message dispatcher, and registers router handlers before starting `b.Start()`.

### Configuration

- `internal/config`: Loads settings from environment variables with fail-safe defaults (`strOrDefault`, `intOrDefault`, `float64OrDefault`).
- Key env variables include Telegram tokens, Polymarket API keys, Bitcointalk/Nodeseek RSS limits and intervals, CoinGecko API keys, and Crocodile volume-spike scanner thresholds.

### Storage

- `internal/store`: Defines the unified `Store` interface (`Set`, `Load`, `Save`) and `FileStore` implementation, responsible for loading and saving persistence data (such as Nodeseek keywords, Bitcointalk filters, CoinGecko holdings, and Crocodile watchlists) via local files or remote data sources.

### Message Delivery Model

- `internal/bot/dispatcher.go`: Consumes messages from a bounded `models.Message` channel and sends them via Telebot (`KindText` or `KindMarkdown`).

### Router & Handlers

- `internal/router/router.go`: Defines the `Handler` interface (`Cmd() string`, `Handle(c tb.Context) error`) and `Dependencies` container.
- Handlers are organized by domain under `internal/handler/`:
  - `bbs/bitcointalk`: Bitcointalk topic monitoring & notification.
  - `bbs/nodeseek`: NodeSeek RSS topic monitoring & notification.
  - `polymarket`: Polymarket holdings and L2 credential checks.
  - `crypto/coingecko`: CoinGecko price querying, searching, and 24h price reporting.
  - `crypto/crocodile`: Daily UTC 00:05 volume-spike alert engine for crypto assets (Markdown output).
  - `telegram`: Basic utility commands (e.g. `/chatid`).
