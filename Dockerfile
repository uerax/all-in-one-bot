FROM alpine:latest

# 引入 TARGETARCH 变量
ARG TARGETARCH

RUN apk add --no-cache curl ca-certificates bash libc6-compat libgcc libstdc++

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
  curl -L "https://github.com/uerax/all-in-one-bot/releases/download/$LATEST/$AIO_BIN" -o /usr/local/bin/aio; \
  chmod +x /usr/local/bin/aio; \ 
  curl -L "https://raw.githubusercontent.com/uerax/all-in-one-bot/refs/heads/master/all-in-one-bot.yml" -o /etc/aio/all-in-one-bot.yml; 

VOLUME ["/var/log/aio"]
VOLUME ["/etc/aio"]

ENV TZ=Asia/Shanghai

CMD ["/usr/local/bin/aio", "-c", "/etc/aio/all-in-one-bot.yml"]
