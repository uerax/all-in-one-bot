#!/bin/bash

#fonts color
Green="\033[32m"
Red="\033[31m"
Yellow="\033[33m"
Blue='\033[0;34m'         # Blue
Purple='\033[0;35m'       # Purple
Cyan='\033[0;36m'         # Cyan
White='\033[0;37m'

GreenBG="\033[42;37m"
RedBG="\033[41;37m"
Font="\033[0m"

version=v0.0.1
prj_name="aio-bot"
project_dir="/usr/local/bin"
prj_url="https://api.github.com/repos/uerax/all-in-one-bot/releases/latest"
cfg_url=""

env() {
    apt update
    apt install -y curl
    apt install -y skill
    apt install -y tail
}

install() {

    # 下载链接
    download_url=$(curl -sL $prj_url | grep "browser_download_url" | cut -d '"' -f 4)
    v=$(curl -sL $prj_url | grep "tag_name" | cut -d '"' -f 4)
    cfg_url="https://raw.githubusercontent.com/uerax/all-in-one-bot/$v/all-in-one-bot.yml"

    # 创建项目目录
    mkdir -p "$project_dir/$prj_name/log"

    curl -L "$download_url" -o "$project_dir/$prj_name/$prj_name"
    curl -L "$cfg_url" -o "$project_dir/$prj_name/all-in-one-bot.yml"

    chmod +x $project_dir/$prj_name/$prj_name

    echo "最新版本已下载并放置在 $project_dir 目录中"
}

start() {
    stop
    cd $project_dir/$prj_name

    ./$prj_name -c $project_dir/$prj_name/all-in-one-bot.yml  > ./log/aio.log 2>&1 &
}

stop() {
    skill $prj_name
}

log() {
    tail -f $project_dir/$prj_name/log/aio.log
}

write() {
    vim $project_dir/$prj_name/all-in-one-bot.yml
}


menu() {
    echo -e "${Cyan}——————————————— 脚本信息 ———————————————${Font}"
    echo -e "\t\t${Yellow}aio-bot 操作脚本${Font}"
    echo -e "\t${Yellow}---authored by uerax---${Font}"
    echo -e "\t${Yellow}https://github.com/uerax${Font}"
    echo -e "\t\t${Yellow}版本号：${version}${Font}"
    echo -e "${Cyan}——————————————— 操作向导 ———————————————${Font}"
    echo -e "${Green}1)   一键安装${Font}"
    echo -e "${Blue}2)   启动服务${Font}"
    echo -e "${Blue}3)   关闭服务${Font}"
    echo -e "${Blue}4)   查看日志${Font}"
    echo -e "${Blue}5)   编辑配置文件${Font}"
    echo -e "${Red}q)   退出${Font}"
    echo -e "${Cyan}————————————————————————————————————————${Font}\n"

    read -rp "输入数字(回车确认)：" menu_num
    echo -e ""
    case $menu_num in
    1)
    install
    ;;
    2)
    start
    ;;
    3)
    stop
    ;;
    4)
    log
    ;;
    5)
    write
    ;;
    q)
    ;;
    *)
    ;;
    esac
}


menu