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

version=v0.0.7
prj_name="aio"
project_dir="/usr/local/bin"
prj_url="https://api.github.com/repos/uerax/all-in-one-bot/releases/latest"
prj_pre_url="https://api.github.com/repos/uerax/all-in-one-bot/releases"
cfg_path="/usr/local/etc"
log_url="/var/log/"
assets="Aio-linux-64"

env() {
    apt install -y curl
    apt install -y wget
}

is_root() {
    if [ $(id -u) == 0 ]; then
        echo -e "进入安装流程"
        sleep 3
    else
        echo -e "请切使用root用户执行脚本"
        echo -e "切换root用户命令: sudo su"
        exit 1
    fi
}

install() {

    # 下载链接
    
    cfg_url="https://raw.githubusercontent.com/uerax/all-in-one-bot/master/all-in-one-bot.yml"
    v=$(curl -sL $prj_url | grep "tag_name" | cut -d '"' -f 4)
    download_url="https://github.com/uerax/all-in-one-bot/releases/download/$v/$assets"

    # 创建项目目录
    mkdir -p "$project_dir"
    mkdir -p "$log_url/$prj_name"
    mkdir -p "$cfg_path/$prj_name"

    curl -L "$download_url" -o "$project_dir/$prj_name"
    wget --no-check-certificate ${cfg_url} -O ${cfg_path}/${prj_name}/all-in-one-bot.yml

    chmod +x ${project_dir}/${prj_name}

    echo -e "安装完成,版本:$v"

}

start() {
    systemctl restart aio
    
}

stop() {
    systemctl stop aio
}

elog() {
    tail -f $log_url/$prj_name/error.log
}

olog() {
    tail -f $log_url/$prj_name/access.log
}

write() {
    vim $cfg_path/$prj_name/all-in-one-bot.yml
}

add_service() {
    is_root
    env
    install
    set_cfg

cat > /etc/systemd/system/aio.service << EOF
[Unit]
Description=Aio Service
Documentation=https://github.com/uerax/all-in-one-bot
After=network.target nss-lookup.target

[Service]
User=root
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_BIND_SERVICE
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE
NoNewPrivileges=true
ExecStart=$project_dir/$prj_name -c $cfg_path/$prj_name/all-in-one-bot.yml
StandardOutput=append:$log_url$prj_name/access.log
StandardError=append:$log_url$prj_name/error.log
Restart=on-failure
RestartPreventExitStatus=23
LimitNPROC=10000
LimitNOFILE=1000000

[Install]
WantedBy=multi-user.target
EOF

ln -s /etc/systemd/system/aio.service /etc/systemd/system/multi-user.target.wants/aio.service

}

update_aio() {
    env
    systemctl stop aio
    v=$(curl -sL $prj_url | grep "tag_name" | cut -d '"' -f 4)
    url="https://github.com/uerax/all-in-one-bot/releases/download/$v/$assets"
    wget -q $url -O $project_dir/$prj_name
    chmod +x ${project_dir}/${prj_name}
    systemctl start aio
    echo -e "更新完成,版本:$v"
}

pre_update_aio() {
    env
    systemctl stop aio
    v=$(curl -sL $prj_pre_url | grep "tag_name" | head -n 1 | cut -d '"' -f 4)
    url="https://github.com/uerax/all-in-one-bot/releases/download/$v/$assets"
    wget -q $url -O $project_dir/$prj_name
    chmod +x ${project_dir}/${prj_name}
    systemctl start aio
    echo -e "更新完成,版本:$v"
}

uninstall() {
    systemctl stop aio
    systemctl disable aio
    rm /etc/systemd/system/aio.service
    systemctl daemon-reload

    rm -f $project_dir/$prj_name
    rm -f $cfg_path/$prj_name/all-in-one-bot.yml
    rm -rf $log_url/$prj_name
    
}

set_cfg() {
    echo -e "${Green}1)  设置token${Font}"
    echo -e "${Green}2)  设置chatid${Font}"
    echo -e "${Green}3)  设置chatgpt key${Font}"
    echo -e "${Green}4)  设置pixian key${Font}"
    echo -e "${Green}q)  不设置${Font}"


    read -rp "输入数字(回车确认)：" num
    echo -e ""

    case $num in
    1)
    field_name=token
    ;;
    2)
    field_name=chatId
    ;;
    3)
    field_name=key
    ;;
    4)
    field_name=authorization
    ;;
    q)
    return
    ;;
    *)
    return
    ;;
    esac

    read -rp "输入要设置的值：" val

    read -rp "新的值为: $val, 请确认(y/n):" confirm
    if [ "$confirm" != "y" ]; then
        echo "已取消操作，退出脚本"
        exit
    fi

    sed -i "s/\b\($field_name: \).*/\1$val/" ${cfg_path}/${prj_name}/all-in-one-bot.yml

}

menu() {
    echo -e "${Cyan}——————————————— 脚本信息 ———————————————${Font}"
    echo -e "\t${Yellow}    aio-bot 操作脚本${Font}"
    echo -e "\t${Yellow}---authored by uerax---${Font}"
    echo -e "\t${Yellow}https://github.com/uerax${Font}"
    echo -e "\t${Yellow}    版本号：${version}${Font}"
    echo -e "${Cyan}——————————————— 操作向导 ———————————————${Font}"
    echo -e "${Green}1)   一键安装${Font}"
    echo -e "${Blue}2)   编辑配置文件${Font}"
    echo -e "${Blue}3)   启动服务${Font}"
    echo -e "${Blue}4)   关闭服务${Font}"
    echo -e "${Blue}5)   查看 error 日志${Font}"
    echo -e "${Blue}6)   查看 output 日志${Font}"
    echo -e "${Blue}7)   更新 AIO${Font}"
    echo -e "${Blue}8)   添加配置${Font}"

    echo -e "${Blue}10)   完全卸载${Font}"
    echo -e "${Red}q)   退出${Font}"
    echo -e "${Cyan}————————————————————————————————————————${Font}\n"

    read -rp "输入数字(回车确认)：" menu_num
    echo -e ""
    case $menu_num in
    1)
    add_service
    ;;
    2)
    write
    ;;
    3)
    start
    ;;
    4)
    stop
    ;;
    5)
    elog
    ;;
    6)
    olog
    ;;
    7)
    update_aio
    ;;
    8)
    set_cfg
    ;;
    10)
    uninstall
    ;;
    q)
    ;;
    *)
    ;;
    esac
}

case $1 in
    install)
        add_service
        ;;
    uninstall)
        uninstall
        ;;
    update)
        update_aio
        ;;
    pre)
        pre_update_aio
        ;;
    *)
        menu
        ;;
esac