#!/bin/bash

RELEASE_NOTES="Release notes"  # 发布说明
prj_url="https://api.github.com/repos/uerax/all-in-one-bot/releases/latest"
version=$(curl -sL $prj_url | grep "tag_name" | cut -d '"' -f 4)

read -rp "当前版本为: $version, 请输入新的版本号: " input

read -rp "新的版本号为: $input, 请确认(y/n):" confirm
if [ "$confirm" != "y" ]; then
    echo "已取消操作，退出脚本"
    exit
fi

echo "Building Linux versions..."
GOOS=linux GOARCH=amd64 go build -o build/$input/Aio-linux-64 main.go
GOOS=linux GOARCH=arm64 go build -o build/$input/Aio-linux-arm64 main.go

# 打包 Windows 版本
echo "Building Windows versions..."
GOOS=windows GOARCH=amd64 go build -o build/$input/Aio-windows.exe main.go
GOOS=windows GOARCH=arm64 go build -o build/$input/Aio-windows-arm64.exe main.go

# 打包 macOS 版本
echo "Building macOS versions..."
GOOS=darwin GOARCH=arm64 go build -o build/$input/Aio-macos-arm64 main.go

echo ""
echo "打包完成"

#gh release create "$version" --title "$version" --notes "$RELEASE_NOTES"

# for file in build/$version/*; do
#   gh release upload "$version" "$file" --clobber
# done
