FROM alpine:3.18

# 引入 TARGETARCH 变量
ARG TARGETARCH

RUN apk add --no-cache curl ca-certificates bash

WORKDIR /

# 根据架构下载二进制文件、创建配置文件和日志目录
RUN set -e; \
  if [ "$TARGETARCH" = "x86_64" ] || [ "$TARGETARCH" = "amd64" ]; then \
    AIO_BIN="Aio-linux-amd64"; \
  elif [ "$TARGETARCH" = "aarch64" ] || [ "$TARGETARCH" = "arm64" ]; then \
    AIO_BIN="Aio-linux-arm64"; \
  else \
    echo "Unsupported architecture: $ARCH"; exit 1; \
  fi; \
  mkdir -p /var/log/aio /etc/aio; \
  chmod 0755 /var/log/aio /etc/aio; \
  LATEST=$(curl -sL https://api.github.com/repos/uerax/all-in-one-bot/releases/latest | grep "tag_name" | cut -d '"' -f 4); \
  curl -L "https://github.com/uerax/all-in-one-bot/releases/download/$LATEST/$AIO_BIN" -o /var/lib/aio; \
  chmod +x /var/lib/aio; \
  if [ ! -f /etc/aio/all-in-one-bot.yml ]; then \
    curl -L "https://raw.githubusercontent.com/uerax/all-in-one-bot/master/all-in-one-bot.yml" -o /etc/aio/all-in-one-bot.yml; \
    echo "Configuration downloaded. Please edit /etc/aio/all-in-one-bot.yml and restart the container."; \
    exit 0; \
  fi

VOLUME ["/var/log/aio"]

ENV TZ=Asia/Shanghai

CMD ["/var/lib/aio", "-c", "/etc/aio/all-in-one-bot.yml"]
