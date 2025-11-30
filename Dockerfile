FROM alpine:3.18

# Minimal runtime image. At container start the entrypoint downloads the
# latest release binary (Aio-linux-amd64 by default) and runs it. The image
# also includes a default `all-in-one-bot.yml` so the container can run
# without host-provided config; users can override by bind-mounting
# ./config -> /usr/local/etc/aio.

RUN apk add --no-cache curl ca-certificates bash

WORKDIR /

# Copy entrypoint script that will download the release asset at runtime
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Provide a default config baked into the image so 'docker compose up'
# works out of the box. Users can override by mounting a host directory.
RUN mkdir -p /usr/local/etc/aio && chmod 0755 /usr/local/etc/aio
COPY all-in-one-bot.yml /usr/local/etc/aio/all-in-one-bot.yml

# Keep logs as a volume so logs can be persisted by the host if desired
VOLUME ["/var/log/aio"]

ENV TZ=Asia/Shanghai

ENTRYPOINT ["/entrypoint.sh"]
