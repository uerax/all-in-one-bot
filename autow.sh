#!/bin/bash

echo -e ""
read -rp "命令: " cmd
read -rp "描述: " tip
echo -e ""

readme="$cmd - $tip"
cfg="    - \"$cmd - $tip\""
l="- [x] $cmd $tip"

echo -e "1)  crypto"
echo -e "2)  video"
echo -e "3)  image"
echo -e "4)  utils"
echo -e "5)  list"
read -rp "分类: " type
case $type in
    1)
    t="- \"list_wallet_tracking - 列出正在追踪的聪明钱包地址\""
    list="### 1. 加密货币监控功能清单"
    sed -i "/$list/a\\$l" README.md
    ;;
    2)
    t="- \"douyin_download - 下载douyin的视频\""
    list="### 7. 视频下载"
    sed -i "/$list/a\\$l" README.md
    ;;
    3)
    t="- \"cutout - 抠图功能\""
    list="### 8. 贴纸和GIF下载"
    sed -i "/$list/a\\$l" README.md
    ;;
    4)
    t="- \"chatid - 查询chatid\""
    list="### 9. 工具箱"
    sed -i "/$list/a\\$l" README.md
    ;;
    5)
    t="- \"cmd_list - 列出全部功能\""
    ;;
    *)
    list="### 1. 加密货币监控功能清单"
    t="- \"list_wallet_tracking - 列出正在追踪的聪明钱包地址\""
    sed -i "/$list/a\\$l" README.md
    ;;
esac

sed -i "/chatid - 查询chatid/a\\$readme" README.md

sed -i "/$t/a\\$cfg" all-in-one-bot.yml

echo -e "case \"$cmd\":"
echo -e "   Cmd = \"$cmd\""
echo -e "   tips(update.Message.Chat.ID, \"$tip\")"
