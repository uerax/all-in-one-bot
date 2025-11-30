#!/bin/sh
set -eu

# Configurable variables:
# ASSET_NAME - name of the release asset to download (default Aio-linux-amd64)
# REPO - owner/repo (default uerax/all-in-one-bot)
# RELEASE_URL - optional full URL to download (overrides REPO/ASSET_NAME)
# CONFIG_PATH - path to config file passed to binary

ASSET_NAME=${ASSET_NAME:-Aio-linux-amd64}
REPO=${REPO:-uerax/all-in-one-bot}
RELEASE_URL=${RELEASE_URL:-}
CONFIG_PATH=${CONFIG_PATH:-/usr/local/etc/aio/all-in-one-bot.yml}

BIN=/usr/local/bin/aio

download() {
    url=${RELEASE_URL}
    if [ -z "$url" ]; then
        url="https://github.com/${REPO}/releases/latest/download/${ASSET_NAME}"
    fi
    echo "Downloading $url ..."
    # retry a few times
    if ! curl -fSL --retry 3 --retry-delay 2 "$url" -o "$BIN"; then
        echo "Failed to download release asset: $url" >&2
        return 1
    fi
    chmod +x "$BIN"
}

# Ensure config dir and logs exist
mkdir -p "$(dirname "$CONFIG_PATH")" /var/log/aio

# On first startup, if host config doesn't exist, copy default from image
if [ ! -f "$CONFIG_PATH" ]; then
    DEFAULT_CONFIG="/usr/local/etc/aio/all-in-one-bot.yml.default"
    if [ -f "$DEFAULT_CONFIG" ]; then
        echo "First startup: Creating default config at $CONFIG_PATH"
        cp "$DEFAULT_CONFIG" "$CONFIG_PATH"
        echo "Config created from image default. Please edit and restart container."
    else
        echo "WARNING: No config file found and no default available." >&2
    fi
fi

if [ ! -x "$BIN" ]; then
    if ! download; then
        echo "Download failed and no local binary present. Exiting." >&2
        exit 2
    fi
fi

echo "Starting aio with config: $CONFIG_PATH"
exec "$BIN" -c "$CONFIG_PATH"
