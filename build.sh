#!/bin/bash

echo "Building Linux versions..."
GOOS=linux GOARCH=amd64 go build -o Aio-linux-64 main.go
GOOS=linux GOARCH=arm64 go build -o Aio-linux-arm64 main.go

# 打包 Windows 版本
echo "Building Windows versions..."
GOOS=windows GOARCH=amd64 go build -o Aio-windows.exe main.go
GOOS=windows GOARCH=arm64 go build -o Aio-windows-arm64.exe main.go

# 打包 macOS 版本
echo "Building macOS versions..."
GOOS=darwin GOARCH=arm64 go build -o Aio-macos-arm64 main.go