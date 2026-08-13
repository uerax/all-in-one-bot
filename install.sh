#!/bin/bash

# ==============================================================================
# all-in-one-bot (lite) 一键安装与管理脚本
# ==============================================================================

# 颜色定义
Green="\033[32m"
Red="\033[31m"
Yellow="\033[33m"
Blue="\033[34m"
Cyan="\033[36m"
Font="\033[0m"

PRJ_NAME="aio"
BIN_NAME="all-in-one-bot"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/aio"
LOG_DIR="/var/log/aio"
SYSTEMD_PATH="/etc/systemd/system/aio.service"

REPO_OWNER="uerax"
REPO_NAME="all-in-one-bot"
RELEASE_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
CONFIG_URL="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/refs/heads/lite/config.example.yaml"
ENV_URL="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/refs/heads/lite/.env.example"

# 检查 root 权限
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        echo -e "${Red}错误：请使用 root 权限执行此脚本 (例如: sudo bash install.sh)${Font}"
        exit 1
    fi
}

# 自动检测架构
detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64 | amd64)
            ARCH="linux-amd64"
            ;;
        aarch64 | arm64)
            ARCH="linux-arm64"
            ;;
        *)
            echo -e "${Red}错误：暂不支持的系统架构: ${arch}${Font}"
            exit 1
            ;;
    esac
}

# 安装依赖
install_dependencies() {
    echo -e "${Green}正在检查并安装基础依赖 (curl, wget, tar)...${Font}"
    if command -v apt-get &>/dev/null; then
        apt-get update -y && apt-get install -y curl wget ca-certificates
    elif command -v yum &>/dev/null; then
        yum install -y curl wget ca-certificates
    elif command -v dnf &>/dev/null; then
        dnf install -y curl wget ca-certificates
    fi
}

# 获取最新 Tag
get_latest_version() {
    echo -e "${Cyan}正在从 GitHub 获取最新发布版本号...${Font}"
    LATEST_TAG=$(curl -sL "${RELEASE_URL}" | grep '"tag_name":' | cut -d '"' -f 4)
    if [ -z "$LATEST_TAG" ]; then
        echo -e "${Yellow}警告：无法从 GitHub API 获取 Tag，默认尝试 latest 关联资源${Font}"
        LATEST_TAG="latest"
    else
        echo -e "${Green}检测到最新版本: ${LATEST_TAG}${Font}"
    fi
}

# 检查旧版 (v2) 部署
check_legacy_v2() {
    local legacy_cfg="/usr/local/etc/aio/all-in-one-bot.yml"
    local legacy_bin="/usr/local/bin/aio"

    if [ -f "$legacy_cfg" ] || [ -f "$legacy_bin" ]; then
        echo -e "\n${Yellow}====================================================${Font}"
        echo -e "${Yellow}提示：检测到当前服务器已安装旧版 (v2) 项目！${Font}"
        echo -e "${Yellow}旧版二进制路径: ${legacy_bin}${Font}"
        echo -e "${Yellow}旧版配置文件: ${legacy_cfg}${Font}"
        echo -e "${Yellow}注意：一键安装会将 Systemd 服务 aio.service 升级指向 lite 版本。${Font}"
        echo -e "${Yellow}你的旧版配置文件和二进制会被完整保留在原目录，不会被删除。${Font}"
        echo -e "${Yellow}====================================================${Font}\n"
        read -rp "是否继续升级安装 lite 版本？[Y/n]: " continue_install
        if [[ "$continue_install" =~ ^[Nn]$ ]]; then
            echo -e "${Yellow}已取消安装。${Font}"
            exit 0
        fi
    fi
}

# 下载并安装二进制文件及默认配置
install_aio() {
    check_root
    check_legacy_v2
    detect_arch
    install_dependencies
    get_latest_version

    # 创建必要的目录
    mkdir -p "${INSTALL_DIR}"
    mkdir -p "${CONFIG_DIR}"
    mkdir -p "${LOG_DIR}"

    local download_url
    if [ "$LATEST_TAG" = "latest" ]; then
        download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/Aio-${ARCH}"
    else
        download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_TAG}/Aio-${ARCH}"
    fi

    echo -e "${Cyan}正在下载二进制文件: ${download_url} ...${Font}"
    if ! curl -L "$download_url" -o "${INSTALL_DIR}/${BIN_NAME}"; then
        echo -e "${Red}错误：二进制文件下载失败！请检查网络或 GitHub Release${Font}"
        exit 1
    fi
    chmod +x "${INSTALL_DIR}/${BIN_NAME}"

    # 下载配置文件模板（若不存在）
    if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
        echo -e "${Cyan}正在下载默认配置文件模板 config.yaml ...${Font}"
        curl -sL "${CONFIG_URL}" -o "${CONFIG_DIR}/config.yaml"
    fi

    if [ ! -f "${CONFIG_DIR}/.env" ]; then
        echo -e "${Cyan}正在下载默认环境变量模板 .env ...${Font}"
        curl -sL "${ENV_URL}" -o "${CONFIG_DIR}/.env"
    fi

    # 创建 systemd 服务
    cat > "${SYSTEMD_PATH}" << EOF
[Unit]
Description=all-in-one-bot (lite) Telegram Bot Service
Documentation=https://github.com/uerax/all-in-one-bot
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=${CONFIG_DIR}
ExecStart=${INSTALL_DIR}/${BIN_NAME} -config ${CONFIG_DIR}/config.yaml
Restart=on-failure
RestartSec=5s
LimitNOFILE=65535
LimitNPROC=65535

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable aio

    echo -e "\n${Green}====================================================${Font}"
    echo -e "${Green}恭喜！all-in-one-bot 安装成功！${Font}"
    echo -e "${Green}二进制文件路径 : ${INSTALL_DIR}/${BIN_NAME}${Font}"
    echo -e "${Green}配置文件目录   : ${CONFIG_DIR}/config.yaml${Font}"
    echo -e "${Green}Systemd 服务名  : aio.service${Font}"
    echo -e "${Yellow}提示：请编辑 ${CONFIG_DIR}/config.yaml 或 ${CONFIG_DIR}/.env 填入你的 TELEGRAM_TOKEN 后启动服务。${Font}"
    echo -e "${Green}====================================================${Font}\n"
}

# 启动服务
start_aio() {
    check_root
    systemctl start aio
    echo -e "${Green}服务已启动！运行状态：${Font}"
    systemctl status aio --no-pager
}

# 停止服务
stop_aio() {
    check_root
    systemctl stop aio
    echo -e "${Yellow}服务已停止。${Font}"
}

# 重启服务
restart_aio() {
    check_root
    systemctl restart aio
    echo -e "${Green}服务已重启！运行状态：${Font}"
    systemctl status aio --no-pager
}

# 查看日志
logs_aio() {
    journalctl -u aio -f -n 100
}

# 更新二进制
update_aio() {
    check_root
    detect_arch
    get_latest_version

    echo -e "${Yellow}正在停止当前运行中的服务...${Font}"
    systemctl stop aio

    local download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_TAG}/Aio-${ARCH}"
    echo -e "${Cyan}正在下载最新的二进制文件...${Font}"
    if curl -L "$download_url" -o "${INSTALL_DIR}/${BIN_NAME}"; then
        chmod +x "${INSTALL_DIR}/${BIN_NAME}"
        systemctl start aio
        echo -e "${Green}更新成功并已重启服务！版本号：${LATEST_TAG}${Font}"
    else
        echo -e "${Red}更新下载失败，重新尝试启动原服务...${Font}"
        systemctl start aio
    fi
}

# 完全卸载
uninstall_aio() {
    check_root
    read -rp "确定要完全卸载 all-in-one-bot 吗？(包括配置文件和日志) [y/N]: " confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        systemctl stop aio &>/dev/null
        systemctl disable aio &>/dev/null
        rm -f "${SYSTEMD_PATH}"
        systemctl daemon-reload

        rm -f "${INSTALL_DIR}/${BIN_NAME}"
        rm -rf "${CONFIG_DIR}"
        rm -rf "${LOG_DIR}"

        echo -e "${Green}已完全卸载 all-in-one-bot。${Font}"
    else
        echo -e "${Yellow}已取消卸载操作。${Font}"
    fi
}

# 菜单向导
menu() {
    clear
    echo -e "${Cyan}====================================================${Font}"
    echo -e "${Green}        all-in-one-bot (lite) 一键管理脚本         ${Font}"
    echo -e "${Cyan}====================================================${Font}"
    echo -e "${Green} 1.${Font} 安装服务 (Install)"
    echo -e "${Green} 2.${Font} 启动服务 (Start)"
    echo -e "${Green} 3.${Font} 停止服务 (Stop)"
    echo -e "${Green} 4.${Font} 重启服务 (Restart)"
    echo -e "${Green} 5.${Font} 查看运行日志 (Realtime Logs)"
    echo -e "${Green} 6.${Font} 更新到最新版本 (Update Binary)"
    echo -e "${Red} 7.${Font} 完全卸载 (Uninstall)"
    echo -e "${Cyan} 0.${Font} 退出脚本 (Exit)"
    echo -e "${Cyan}====================================================${Font}"

    read -rp "请输入数字 [0-7]: " num
    case "$num" in
        1) install_aio ;;
        2) start_aio ;;
        3) stop_aio ;;
        4) restart_aio ;;
        5) logs_aio ;;
        6) update_aio ;;
        7) uninstall_aio ;;
        0) exit 0 ;;
        *) echo -e "${Red}请输入正确的数字！${Font}" && sleep 1 && menu ;;
    esac
}

# 命令行命令行参数直接调度
case "$1" in
    install) install_aio ;;
    start) start_aio ;;
    stop) stop_aio ;;
    restart) restart_aio ;;
    logs) logs_aio ;;
    update) update_aio ;;
    uninstall) uninstall_aio ;;
    *) menu ;;
esac
