#!/bin/bash

# LightCall 安装脚本
# 用于下载、配置和安装 LightCall 呼叫中心系统

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
DOWNLOAD_URL="https://github.com/tcmzzz/lightcall/archive/refs/heads/master.zip"
INSTALL_DIR="./lightcall"
TEMP_DIR=""
ZIP_FILE=""
DOMAIN=""
FAKE_GATEWAY=""
ADMIN_EMAIL=""
ICE_USERNAME=""
ICE_CREDENTIAL=""
PUBLIC_IP=""
ENVIRONMENT=""
SELF_SIGNED_CERT=""

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 清理函数
cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        log_info "正在清理临时目录: $TEMP_DIR"
        rm -rf "$TEMP_DIR"
    fi

    if [ -n "$ZIP_FILE" ] && [ -f "$ZIP_FILE" ]; then
        log_info "正在清理下载文件: $ZIP_FILE"
        rm -f "$ZIP_FILE"
    fi
}

# 错误处理
trap 'log_error "安装过程中发生错误，正在清理..."; cleanup; exit 1' ERR

# 生成随机字符串函数
generate_random_username() {
    # 生成5位随机字母
    local chars="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    local result=""
    for i in {1..5}; do
        result+="${chars:$((RANDOM % ${#chars})):1}"
    done
    echo "$result"
}

generate_random_credential() {
    # 生成6位随机字母+数字+特殊字符
    local chars="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
    local result=""
    for i in {1..6}; do
        result+="${chars:$((RANDOM % ${#chars})):1}"
    done
    echo "$result"
}

# 检查系统依赖
check_dependencies() {
    log_info "检查系统依赖..."

    local missing_deps=()

    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi

    if ! command -v unzip &> /dev/null; then
        missing_deps+=("unzip")
    fi

    if ! command -v rsync &> /dev/null; then
        missing_deps+=("rsync")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "缺少以下依赖: ${missing_deps[*]}"
        log_info "请先安装缺少的依赖，然后重新运行脚本"
        exit 1
    fi

    log_success "系统依赖检查通过"
}

# 下载和解压功能
download_and_extract() {
    log_info "正在下载 LightCall 源码..."

    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    ZIP_FILE="$TEMP_DIR/master.zip"

    # 下载文件
    if ! curl -L -o "$ZIP_FILE" "$DOWNLOAD_URL"; then
        log_error "下载失败，请检查网络连接"
        exit 1
    fi

    log_success "下载完成"

    # 解压文件
    log_info "正在解压文件..."

    if ! unzip -q "$ZIP_FILE" -d "$TEMP_DIR"; then
        log_error "解压失败"
        exit 1
    fi

    log_success "解压完成"
}

# 安装目录配置
configure_install_directory() {
    echo
    read -p "请输入安装目录 (默认: $INSTALL_DIR): " user_input

    if [ -n "$user_input" ]; then
        INSTALL_DIR="$user_input"
    fi

    # 转换为绝对路径
    if [[ "$INSTALL_DIR" != /* ]]; then
        INSTALL_DIR="$(pwd)/$INSTALL_DIR"
    fi

    log_info "安装目录: $INSTALL_DIR"

    # 检查目录是否已存在
    if [ -d "$INSTALL_DIR" ]; then
        log_warning "目录 $INSTALL_DIR 已存在"
        read -p "是否继续？将覆盖现有文件 (y/N): " confirm
        if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
            log_info "安装已取消"
            exit 0
        fi
    fi
}

# 运行环境选择
choose_environment() {
    echo
    echo "请选择运行环境:"
    echo "1) 内网访问"
    echo "2) 公网访问"

    while true; do
        read -p "请输入选择 (1-2, 默认: 2): " env_choice

        # 如果用户直接回车，使用默认值
        if [ -z "$env_choice" ]; then
            env_choice="2"
        fi

        case $env_choice in
            1)
                ENVIRONMENT="internal"
                log_info "已选择: 内网访问"
                break
                ;;
            2)
                ENVIRONMENT="public"
                log_info "已选择: 公网访问"
                break
                ;;
            *)
                log_error "无效选择，请输入 1 或 2"
                ;;
        esac
    done
}

# 公网IP检测和配置
configure_public_ip() {
    PUBLIC_IP=""

    if [ "$ENVIRONMENT" = "public" ]; then
        echo
        log_info "正在检测公网IP..."

        # 尝试多个IP检测服务
        local ip_services=(
            "ifconfig.me"
            "ipinfo.io/ip"
            "icanhazip.com"
            "ident.me"
        )

        for service in "${ip_services[@]}"; do
            if IP=$(curl -s --connect-timeout 5 "$service"); then
                if [[ $IP =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
                    PUBLIC_IP="$IP"
                    break
                fi
            fi
        done

        if [ -n "$PUBLIC_IP" ]; then
            log_success "检测到公网IP: $PUBLIC_IP"
            read -p "请输入公网IP地址 (默认: $PUBLIC_IP): " user_ip

            if [ -n "$user_ip" ]; then
                PUBLIC_IP="$user_ip"
            fi
        else
            log_warning "无法自动检测公网IP"
            while true; do
                read -p "请手动输入公网IP地址: " user_ip
                if [[ $user_ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
                    PUBLIC_IP="$user_ip"
                    break
                else
                    log_error "IP地址格式不正确，请重新输入"
                fi
            done
        fi

        log_info "使用公网IP: $PUBLIC_IP"
    fi
}

# 配置访问域名
configure_domain() {
    echo
    read -p "请输入访问域名 (默认: your.domain.com): " user_domain

    if [ -n "$user_domain" ]; then
        DOMAIN="$user_domain"
    else
        DOMAIN="your.domain.com"
    fi

    log_info "使用域名: $DOMAIN"
}

# 配置自签证书
configure_self_signed_cert() {
    echo
    echo "是否需要自签证书？"
    echo "说明：使用实际域名会自动申请免费证书，使用内部域名或hosts方式访问才需要自签证书"
    echo "1) 是 - 需要自签证书"
    echo "2) 否 - 不需要自签证书"

    while true; do
        read -p "请输入选择 (1-2, 默认: 1): " cert_choice

        # 如果用户直接回车，使用默认值
        if [ -z "$cert_choice" ]; then
            cert_choice="1"
        fi

        case $cert_choice in
            1)
                SELF_SIGNED_CERT="yes"
                log_info "已选择: 需要自签证书"
                break
                ;;
            2)
                SELF_SIGNED_CERT="no"
                log_info "已选择: 不需要自签证书"
                break
                ;;
            *)
                log_error "无效选择，请输入 1 或 2"
                ;;
        esac
    done
}

# 配置模拟网关
configure_fake_gateway() {
    echo
    echo "是否需要模拟网关？"
    echo "说明：模拟网关用于测试环境，生产环境通常不需要"
    echo "1) 是 - 需要模拟网关"
    echo "2) 否 - 不需要模拟网关"

    while true; do
        read -p "请输入选择 (1-2, 默认: 1): " gateway_choice

        # 如果用户直接回车，使用默认值
        if [ -z "$gateway_choice" ]; then
            gateway_choice="1"
        fi

        case $gateway_choice in
            1)
                FAKE_GATEWAY="yes"
                log_info "已选择: 需要模拟网关"
                break
                ;;
            2)
                FAKE_GATEWAY="no"
                log_info "已选择: 不需要模拟网关"
                break
                ;;
            *)
                log_error "无效选择，请输入 1 或 2"
                ;;
        esac
    done
}

# 配置管理员邮箱
configure_admin_email() {
    echo
    local default_email="admin@test.com"
    read -p "请输入管理员邮箱 (默认: $default_email): " user_email

    if [ -n "$user_email" ]; then
        # 验证邮箱格式
        if [[ $user_email =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
            ADMIN_EMAIL="$user_email"
        else
            log_warning "邮箱格式不正确，使用默认邮箱"
            ADMIN_EMAIL="$default_email"
        fi
    else
        ADMIN_EMAIL="$default_email"
    fi

    log_info "管理员邮箱: $ADMIN_EMAIL"
}

# 文件复制和安装
install_files() {
    log_info "正在逐个处理文件..."

    # 创建安装目录
    mkdir -p "$INSTALL_DIR"

    # 查找解压后的源码目录
    local source_dir=""
    for dir in "$TEMP_DIR"/*; do
        if [ -d "$dir" ] && [[ "$(basename "$dir")" == lightcall-* ]]; then
            source_dir="$dir"
            break
        fi
    done

    if [ -z "$source_dir" ]; then
        log_error "找不到解压后的源码目录"
        exit 1
    fi

    # 1. 处理 fake-provider（如果需要）
    if [ "$FAKE_GATEWAY" = "yes" ]; then
        log_info "正在处理模拟网关文件..."
        local fake_dir="$source_dir/example/fake-provider"
        if [ -d "$fake_dir" ]; then
            mkdir -p "$INSTALL_DIR/fake-provider"
            cp -r "$fake_dir"/* "$INSTALL_DIR/fake-provider/"
            log_success "模拟网关文件处理完成"
        else
            log_warning "未找到模拟网关目录 example/fake-provider"
        fi
    fi

    # 2. 处理 Caddyfile
    log_info "正在处理 Caddyfile..."
    local caddyfile_source="$source_dir/example/prod/Caddyfile"
    if [ -f "$caddyfile_source" ]; then
        cp "$caddyfile_source" "$INSTALL_DIR/Caddyfile"
        process_caddyfile
        log_success "Caddyfile 处理完成"
    else
        log_warning "未找到 Caddyfile"
    fi

    # 3. 处理 docker-compose.yaml
    log_info "正在处理 docker-compose.yaml..."
    local compose_source="$source_dir/example/prod/docker-compose.yaml"
    if [ -f "$compose_source" ]; then
        cp "$compose_source" "$INSTALL_DIR/docker-compose.yaml"
        process_docker_compose
        log_success "docker-compose.yaml 处理完成"
    else
        log_warning "未找到 docker-compose.yaml"
    fi

    # 4. 处理 turnserver.conf（仅公网环境）
    if [ "$ENVIRONMENT" = "public" ]; then
        log_info "正在处理 turnserver.conf..."
        local turnserver_source="$source_dir/example/prod/turnserver.conf"
        if [ -f "$turnserver_source" ]; then
            cp "$turnserver_source" "$INSTALL_DIR/turnserver.conf"
            log_success "turnserver.conf 处理完成"
        else
            log_warning "未找到 turnserver.conf"
        fi
    fi

    # 5. 处理其他 prod 目录文件（config.yaml）
    log_info "正在处理其他配置文件..."
    local config_source="$source_dir/example/prod/config.yaml"
    if [ -f "$config_source" ]; then
        cp "$config_source" "$INSTALL_DIR/config.yaml"
        log_success "config.yaml 处理完成"
    fi

    # 6. 创建 .env 配置文件
    create_config_files

    log_success "所有文件处理完成"
}

# 创建配置文件
create_config_files() {
    log_info "正在生成 .env 配置文件..."

    # 生成 ICE 凭据（公网环境需要）
    if [ "$ENVIRONMENT" = "public" ]; then
        ICE_USERNAME=$(generate_random_username)
        ICE_CREDENTIAL=$(generate_random_credential)
    fi

    # 创建 .env 文件
    cat > "$INSTALL_DIR/.env" << EOF
DOMAIN=$DOMAIN
ADMIN_EMAIL=$ADMIN_EMAIL
EOF

    if [ "$ENVIRONMENT" = "public" ]; then
        cat >> "$INSTALL_DIR/.env" << EOF
EXT_IP=$PUBLIC_IP
ICE_URLS=turn:$PUBLIC_IP:3478
ICE_USERNAME=$ICE_USERNAME
ICE_CREDENTIAL=$ICE_CREDENTIAL
EOF
    fi

    log_success ".env 配置文件生成完成"
}

# 处理 Caddyfile
process_caddyfile() {
    log_info "正在处理 Caddyfile..."

    local caddyfile="$INSTALL_DIR/Caddyfile"

    if [ -f "$caddyfile" ]; then
        # 如果不需要自签证书，删除 tls internal 行
        if [ "$SELF_SIGNED_CERT" = "no" ]; then
            sed -i '/tls internal/d' "$caddyfile"
        fi

        log_success "Caddyfile 处理完成"
    else
        log_warning "未找到 Caddyfile，跳过处理"
    fi
}

# 处理 docker-compose.yaml
process_docker_compose() {
    log_info "正在处理 docker-compose.yaml..."

    local compose_file="$INSTALL_DIR/docker-compose.yaml"

    if [ -f "$compose_file" ]; then

        # 内网环境：删除 coturn 服务
        if [ "$ENVIRONMENT" = "internal" ]; then
            # 使用 awk 更精确地处理，删除 coturn 服务块
            awk '
            /^  coturn:/ { in_coturn=1 }
            in_coturn && /^  [^ ]/ && !/^  coturn:/ { in_coturn=0 }
            !in_coturn { print }
            ' "$compose_file" > "$compose_file.tmp" && mv "$compose_file.tmp" "$compose_file"
        fi

        # 不需要模拟网关：从 fakeprovider 服务开始删除到文件末尾
        if [ "$FAKE_GATEWAY" = "no" ]; then
            # 找到 fakeprovider 服务的行号
            local fakeprovider_line=$(grep -n "^  fakeprovider:" "$compose_file" | cut -d: -f1)
            if [ -n "$fakeprovider_line" ]; then
                # 删除从 fakeprovider 行开始到文件末尾的所有内容
                sed -i "${fakeprovider_line},\$d" "$compose_file"
            fi
        fi

        log_success "docker-compose.yaml 处理完成"
    else
        log_warning "未找到 docker-compose.yaml，跳过处理"
    fi
}


# 显示安装完成信息
show_completion_info() {
    echo
    echo "==============================================="
    log_success "LightCall 安装完成！"
    echo "==============================================="
    echo
    echo "安装目录: $INSTALL_DIR"
    echo "运行环境: $([ "$ENVIRONMENT" = "public" ] && echo "公网访问 ($PUBLIC_IP)" || echo "内网访问")"
    echo "访问域名: $DOMAIN"
    echo "自签证书: $([ "$SELF_SIGNED_CERT" = "yes" ] && echo "是" || echo "否")"
    echo "模拟网关: $([ "$FAKE_GATEWAY" = "yes" ] && echo "是" || echo "否")"
    echo "管理员邮箱: $ADMIN_EMAIL"
    if [ "$ENVIRONMENT" = "public" ] && [ -n "$ICE_USERNAME" ]; then
        echo "ICE 配置: $ICE_USERNAME / $ICE_CREDENTIAL"
    fi
    echo
    echo "下一步操作:"
    echo "1. cd $INSTALL_DIR"
    echo "2. 检查并修改配置文件（如需要）"
    echo "3. 启动服务: docker-compose up -d"
    echo
    if [ "$SELF_SIGNED_CERT" = "yes" ]; then
        echo "注意：使用自签证书，首次访问时需要在浏览器中信任证书"
    fi
    echo
    echo "帮助文档: https://github.com/tcmzzz/lightcall"
    echo "==============================================="
}

# 主函数
main() {
    echo "==============================================="
    echo "           LightCall 安装脚本"
    echo "==============================================="
    echo

    # 检查依赖
    check_dependencies

    # 下载和解压
    download_and_extract

    # 配置安装目录
    configure_install_directory

    # 选择运行环境
    choose_environment

    # 配置公网IP（如果需要）
    configure_public_ip

    # 配置访问域名
    configure_domain

    # 配置自签证书
    configure_self_signed_cert

    # 配置模拟网关
    configure_fake_gateway

    # 配置管理员邮箱
    configure_admin_email

    # 安装文件
    install_files

    # 清理临时文件
    cleanup

    # 显示完成信息
    show_completion_info
}

# 运行主函数
main "$@"
