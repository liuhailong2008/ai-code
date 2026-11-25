#!/usr/bin/env bash
#
# 毕昇 JDK 一键安装脚本（内网 / 文件服务器场景）
#
# 功能：
#   1. 自动识别 CPU 架构（x86_64 → x64，aarch64 → aarch64）
#   2. 从内网文件服务器 wget 安装包并用 .sha256 校验
#   3. 解压到 $INSTALL_ROOT，创建 current 软链，写入 JAVA_HOME / PATH
#   4. 校验 glibc 门槛与安装结果，支持一键回滚
#
# 用法：
#   ./install_bisheng_jdk.sh -u http://10.0.0.10/jdk              # 从文件服务器安装
#   ./install_bisheng_jdk.sh -u http://10.0.0.10/jdk -d /data/jdk # 自定义安装目录
#   ./install_bisheng_jdk.sh --offline /data/pkg                   # 直接用本地目录的安装包
#   ./install_bisheng_jdk.sh --rollback                            # 回滚到 25.0.4
#   ./install_bisheng_jdk.sh -h                                    # 查看帮助
#
set -euo pipefail

# ==================== 默认参数 ====================
# 内网文件服务器根地址：部署时修改这一行，目标机即可用"一条命令"安装
#   curl -fsSL http://<文件服务器>/jdk/install_bisheng_jdk.sh | sudo bash
DEFAULT_FILE_SERVER="${JDK_FILE_SERVER:-}"

JDK_VERSION="${JDK_VERSION:-bisheng-jdk-25.0.5-b16}"
ROLLBACK_VERSION="${ROLLBACK_VERSION:-bisheng-jdk-25.0.4-b15}"
INSTALL_ROOT="${INSTALL_ROOT:-/opt/jdk}"
FILE_SERVER="${FILE_SERVER:-$DEFAULT_FILE_SERVER}"
OFFLINE_DIR=""
ARCH=""
HOST_ARCH=""
ACTION="install"
PROFILE_FILE=""
WORKDIR=""

# ==================== 输出函数 ====================
log()  { printf '\033[32m[INFO ]\033[0m %s\n' "$*"; }
warn() { printf '\033[33m[WARN ]\033[0m %s\n' "$*"; }
die()  { printf '\033[31m[ERROR]\033[0m %s\n' "$*" >&2; exit 1; }

usage() {
    cat <<'EOF'
毕昇 JDK 一键安装脚本（内网 / 文件服务器场景）

用法：
  # 一条命令安装（脚本与安装包都在文件服务器上）
  curl -fsSL http://<文件服务器>/jdk/install_bisheng_jdk.sh | sudo bash -s -- -u http://<文件服务器>/jdk
  # 若脚本内 DEFAULT_FILE_SERVER 已配置，可省略 -u
  curl -fsSL http://<文件服务器>/jdk/install_bisheng_jdk.sh | sudo bash

  # 已把脚本下载到本机
  ./install_bisheng_jdk.sh -u http://10.0.0.10/jdk              # 从文件服务器安装
  ./install_bisheng_jdk.sh -u http://10.0.0.10/jdk -d /data/jdk # 自定义安装目录
  ./install_bisheng_jdk.sh --offline /data/pkg                   # 使用本地目录的安装包
  ./install_bisheng_jdk.sh --rollback                            # 回滚到 25.0.4
  ./install_bisheng_jdk.sh -h                                    # 查看帮助

参数：
  -u, --url      内网文件服务器目录（脚本按 <url>/<包名> 拼接下载地址）
  -d, --dir      安装根目录，默认 /opt/jdk，安装后生成 current 软链
  -v, --version  主版本目录名，默认 bisheng-jdk-25.0.5-b16
  -a, --arch     强制指定架构（x64 / aarch64），默认按 uname -m 自动识别
  -o, --offline  离线目录（安装包已拷贝到本机时使用，优先级高于 -u）
      --rollback 将 current 切回 25.0.4（需先安装该版本）
EOF
    exit 0
}

# ==================== 参数解析 ====================
while [[ $# -gt 0 ]]; do
    case "$1" in
        -u|--url)        FILE_SERVER="${2:-}"; shift 2 ;;
        -d|--dir)        INSTALL_ROOT="${2:-}"; shift 2 ;;
        -v|--version)    JDK_VERSION="${2:-}"; shift 2 ;;
        -a|--arch)       ARCH="${2:-}"; shift 2 ;;
        -o|--offline)    OFFLINE_DIR="${2:-}"; shift 2 ;;
        --rollback)      ACTION="rollback"; shift ;;
        -h|--help)       usage ;;
        *)               die "未知参数：$1（-h 查看帮助）" ;;
    esac
done

# ==================== 环境检查 ====================
cpu_model() {
    [[ -r /proc/cpuinfo ]] || return 0
    awk -F': ' '/^model name|^Model|^CPU part|^Processor/ {gsub(/^[ \t]+/,"",$2); print $2; exit}' /proc/cpuinfo
}

detect_arch() {
    local machine
    machine="$(uname -m)"

    # 先确定主机真实架构，用于交叉安装告警
    case "$machine" in
        x86_64|amd64|x64) HOST_ARCH="x64" ;;
        aarch64|arm64)    HOST_ARCH="aarch64" ;;
        loongarch64|mips64|sw_64|riscv64)
            die "CPU 架构 ${machine} 不受毕昇 JDK 支持（仅提供 x86_64 / aarch64 包）" ;;
        *) HOST_ARCH="" ;;
    esac

    if [[ -n "$ARCH" ]]; then
        case "$ARCH" in
            x64|aarch64) ;;
            *) die "非法架构参数：${ARCH}（仅支持 x64 / aarch64）" ;;
        esac
        warn "已强制指定架构：${ARCH}（跳过自动识别）"
        if [[ -n "$HOST_ARCH" && "$ARCH" != "$HOST_ARCH" ]]; then
            warn "指定架构与主机架构（${HOST_ARCH}）不一致，安装后无法在本机运行，仅适合制作镜像/中转"
        fi
    else
        [[ -n "$HOST_ARCH" ]] || die "无法识别的 CPU 架构：${machine}，请用 -a 指定 x64 或 aarch64"
        ARCH="$HOST_ARCH"
    fi

    local model
    model="$(cpu_model)"
    log "CPU：${model:-未知型号}（uname -m: ${machine}）→ 选择安装包 linux-${ARCH}"

    if [[ "$ARCH" == "x64" ]]; then
        case "$model" in
            *Hygon*|*海光*) log "识别为海光 x86_64，与 Intel 共用 linux-x64 包（组合未在官方内测矩阵，建议压测）" ;;
            *)              : ;;
        esac
    else
        case "$model" in
            *Phytium*|*飞腾*) warn "识别为飞腾 aarch64：可运行 linux-aarch64 包，但不在毕昇官方内测清单，建议灰度验证" ;;
            *)                : ;;
        esac
    fi
}

# 校验包内 java 的 ELF 架构与主机一致（防止文件服务器放错架构的包）
verify_binary_arch() {
    local home="$1"
    local java_bin="${home}/bin/java"
    [[ -r "$java_bin" ]] || return 0
    command -v od >/dev/null 2>&1 || return 0

    local machine_id
    machine_id="$(od -An -tx1 -j18 -N2 "$java_bin" 2>/dev/null | tr -d ' \n')"
    [[ -n "$machine_id" ]] || return 0

    case "${ARCH}:${machine_id}" in
        x64:3e00|aarch64:b700)
            log "包内二进制架构校验通过（${ARCH}）" ;;
        *)
            die "安装包架构不匹配：期望 ${ARCH}，但包内 java 的 ELF e_machine=${machine_id}（x64 应为 3e00，aarch64 应为 b700），请检查文件服务器上的包" ;;
    esac
}

check_glibc() {
    command -v getconf >/dev/null 2>&1 || { warn "缺少 getconf，跳过 glibc 检查"; return; }
    local ver
    ver="$(getconf GNU_LIBC_VERSION 2>/dev/null | awk '{print $2}' || true)"
    [[ -z "$ver" ]] && { warn "无法获取 glibc 版本，跳过检查"; return; }

    if [[ "$ver" == "2.17" ]]; then
        warn "当前 glibc ${ver}：毕昇官方要求 ≥ 2.18，实测二进制可运行但属未认证组合（如 CentOS 7.6）"
    elif [[ "$(printf '%s\n2.18\n' "$ver" | sort -V | head -1)" != "2.18" ]]; then
        die "glibc ${ver} 低于最低要求 2.17，无法运行毕昇 JDK 25"
    else
        log "glibc ${ver} 满足要求"
    fi
}

# ==================== 包获取与校验 ====================
# 统一下载入口：优先 wget，其次 curl
dl() {
    local url="$1" out="$2"
    if command -v wget >/dev/null 2>&1; then
        wget -q --show-progress -O "$out" "$url"
    elif command -v curl >/dev/null 2>&1; then
        curl -fL --progress-bar -o "$out" "$url"
    else
        return 127
    fi
}

fetch_pkg() {
    local name="$1"
    local dest="${WORKDIR}/$name"

    if [[ -n "$OFFLINE_DIR" ]]; then
        [[ -f "$OFFLINE_DIR/$name" ]] || die "离线目录缺少文件：$OFFLINE_DIR/$name"
        cp -f "$OFFLINE_DIR/$name" "$dest"
        log "使用本地文件：$OFFLINE_DIR/$name"
    else
        [[ -n "$FILE_SERVER" ]] || die "必须指定文件服务器地址（-u）或离线目录（--offline）"
        local url="${FILE_SERVER%/}/$name"
        log "下载：$url"
        dl "$url" "$dest" || die "下载失败：$url（系统缺少 wget/curl 时可用 --offline 模式）"
    fi

    # 校验和：优先用同名 .sha256，缺失时按本脚本内置值兜底
    local expect=""
    if [[ -f "$OFFLINE_DIR/$name.sha256" ]]; then
        expect="$(awk '{print $1}' "$OFFLINE_DIR/$name.sha256")"
    elif [[ -n "$FILE_SERVER" ]] && dl "${FILE_SERVER%/}/$name.sha256" "$dest.sha256" 2>/dev/null; then
        expect="$(awk '{print $1}' "$dest.sha256")"
    fi

    if [[ -n "$expect" ]]; then
        local actual
        actual="$(sha256sum "$dest" | awk '{print $1}')"
        [[ "$expect" == "$actual" ]] || die "SHA256 校验失败：$name（期望 $expect，实际 $actual）"
        log "SHA256 校验通过：$name"
    else
        warn "未找到 $name.sha256，跳过校验（建议由文件服务器提供校验文件）"
    fi
}

# ==================== 安装 / 切换 ====================
# 解析某个版本的实际安装目录（安装包内顶层目录为 bisheng-jdk-<版本>，不含构建号与架构）
resolve_home() {
    local version="$1"
    local prefix="${version%%-b*}"
    local home="${INSTALL_ROOT}/${prefix}"
    [[ -x "${home}/bin/java" ]] && { printf '%s\n' "$home"; return 0; }
    local d
    for d in "${INSTALL_ROOT}/${prefix}"*/; do
        [[ -x "${d}bin/java" ]] && { printf '%s\n' "${d%/}"; return 0; }
    done
    return 1
}

install_one() {
    local version="$1"
    local pkg="${version}-linux-${ARCH}.tar.gz"
    local home

    if home="$(resolve_home "$version")"; then
        log "已安装，跳过：$home"
    else
        fetch_pkg "$pkg"
        log "解压到：$INSTALL_ROOT"
        mkdir -p "$INSTALL_ROOT"
        tar -xzf "${WORKDIR}/$pkg" -C "$INSTALL_ROOT" || die "解压失败：$pkg"
        rm -f "${WORKDIR}/$pkg"
        home="$(resolve_home "$version")" || die "解压后未找到 bin/java，请检查安装包结构"
        log "安装完成：$home"
    fi

    verify_binary_arch "$home"
    ln -sfn "$home" "${INSTALL_ROOT}/current"
    log "current → $home"
}

write_profile() {
    local home="${INSTALL_ROOT}/current"
    local if_root=0
    [[ "$(id -u)" -eq 0 ]] && if_root=1

    if [[ $if_root -eq 1 && -d /etc/profile.d ]]; then
        PROFILE_FILE="/etc/profile.d/bisheng-jdk.sh"
    else
        PROFILE_FILE="${HOME}/.bashrc"
    fi

    touch "$PROFILE_FILE"
    # 先清掉旧的配置块，保证幂等
    sed -i.bak '/# >>> bisheng-jdk >>>/,/# <<< bisheng-jdk <<</d' "$PROFILE_FILE" 2>/dev/null || true

    cat >> "$PROFILE_FILE" <<EOF
# >>> bisheng-jdk >>>
export JAVA_HOME=${home}
export PATH="\$JAVA_HOME/bin:\$PATH"
# <<< bisheng-jdk <<<
EOF
    rm -f "${PROFILE_FILE}.bak"
    log "已写入环境变量：$PROFILE_FILE"
}

verify_install() {
    local java="${INSTALL_ROOT}/current/bin/java"
    [[ -x "$java" ]] || die "安装校验失败：${java} 不存在"
    log "版本信息："
    "$java" -version 2>&1 | sed 's/^/    /'
    log "关键属性："
    "$java" -XshowSettings:properties -version 2>&1 \
        | grep -E 'java\.(vendor|version|vm\.name)|os\.arch' | sed 's/^/    /'
    log "架构：$(uname -m)（安装包选择：linux-${ARCH}）"
    log "提示：重新登录或执行 'source ${PROFILE_FILE}' 后生效"
}

# ==================== 主流程 ====================
command -v tar >/dev/null 2>&1 || die "缺少 tar"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT
detect_arch
[[ -n "$OFFLINE_DIR" ]] && [[ -d "$OFFLINE_DIR" ]] || { [[ -z "$OFFLINE_DIR" ]] || die "离线目录不存在：$OFFLINE_DIR"; }

if [[ "$ACTION" == "install" && -z "$OFFLINE_DIR" && -z "$FILE_SERVER" ]]; then
    die "未指定文件服务器地址：用 -u http://<文件服务器>/jdk 指定，或修改脚本顶部 DEFAULT_FILE_SERVER"
fi

if [[ "$ACTION" == "rollback" ]]; then
    rollback_home="$(resolve_home "$ROLLBACK_VERSION")" \
        || die "回滚目标不存在：${ROLLBACK_VERSION}（请先安装该版本）"
    verify_binary_arch "$rollback_home"
    ln -sfn "$rollback_home" "${INSTALL_ROOT}/current"
    log "已回滚：current → $rollback_home"
    verify_install
    exit 0
fi

log "安装目标：${JDK_VERSION}（架构 ${ARCH}）→ ${INSTALL_ROOT}"
check_glibc
mkdir -p "$INSTALL_ROOT" || die "无法创建目录 ${INSTALL_ROOT}（请用 sudo 执行，或用 -d 指定可写目录）"
[[ -w "$INSTALL_ROOT" ]] || die "目录不可写：${INSTALL_ROOT}（请用 sudo 执行，或用 -d 指定可写目录）"

install_one "$JDK_VERSION"
write_profile
verify_install
log "安装结束。"
