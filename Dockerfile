# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 go build -o /app/all-in-one-bot -trimpath -ldflags "-s -w" main.go

# Final release stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled binary
COPY --from=builder /app/all-in-one-bot /app/all-in-one-bot

ENV TZ=Asia/Shanghai

ENTRYPOINT ["/app/all-in-one-bot"]
