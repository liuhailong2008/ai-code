#!/usr/bin/env bash
#
# 合并 parts/ 下的安装包分片，并用同名 .sha256 校验还原结果。
#
# 用法：
#   ./merge_parts.sh
#
# 行为：
#   1. 检查分片数量、每片大小（除末片外必须为 20 MiB）
#   2. 按 part.00 / part.01 … 顺序拼接为原始安装包
#   3. 用 <包名>.sha256 校验，校验通过才落盘；已存在且校验通过的包会跳过
set -euo pipefail

cd "$(dirname "$0")"

PKGS=(
    "bisheng-jdk-25.0.5-b16-linux-x64.tar.gz"
    "bisheng-jdk-25.0.5-b16-linux-aarch64.tar.gz"
)
CHUNK_SIZE=20971520   # 20 MiB，分片大小上限
PARTS_DIR="parts"

log()  { printf '\033[32m[INFO ]\033[0m %s\n' "$*"; }
die()  { printf '\033[31m[ERROR]\033[0m %s\n' "$*" >&2; exit 1; }

sha256_of() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | awk '{print $1}'
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | awk '{print $1}'
    else
        die "缺少 sha256sum / shasum，无法校验"
    fi
}

merge_one() {
    local pkg="$1"
    local sha_file="${pkg}.sha256"
    local parts=()
    local p i size expect actual

    for p in "${PARTS_DIR}/${pkg}.part."*; do
        [[ -f "$p" ]] || die "分片缺失：${p}（请确认已完整拉取 parts/ 目录）"
        parts+=("$p")
    done
    [[ ${#parts[@]} -gt 0 ]] || die "分片缺失：${PARTS_DIR}/${pkg}.part.* 不存在"
    [[ -f "$sha_file" ]] || die "校验文件缺失：$sha_file"

    # 除最后一片外，每片都应正好是 CHUNK_SIZE，可借此发现漏片/截断
    for ((i = 0; i < ${#parts[@]} - 1; i++)); do
        size="$(wc -c < "${parts[$i]}" | tr -d ' ')"
        [[ "$size" -eq "$CHUNK_SIZE" ]] || \
            die "分片大小异常：${parts[$i]} 为 ${size} 字节，期望 ${CHUNK_SIZE}"
    done

    expect="$(awk 'NR==1{print $1}' "$sha_file")"

    if [[ -f "$pkg" ]]; then
        actual="$(sha256_of "$pkg")"
        [[ "$actual" == "$expect" ]] || \
            die "已存在的 $pkg 校验失败（期望 ${expect}，实际 ${actual}），请删除后重试"
        log "已存在且校验通过，跳过：$pkg"
        return 0
    fi

    log "合并 ${#parts[@]} 个分片 → ${pkg}"
    cat "${parts[@]}" > "${pkg}.tmp"
    actual="$(sha256_of "${pkg}.tmp")"
    if [[ "$actual" != "$expect" ]]; then
        rm -f "${pkg}.tmp"
        die "SHA256 校验失败：${pkg}（期望 ${expect}，实际 ${actual}）"
    fi
    mv "${pkg}.tmp" "$pkg"
    log "SHA256 校验通过：${pkg}（${actual}）"
}

for pkg in "${PKGS[@]}"; do
    merge_one "$pkg"
done

log "合并完成。离线安装：bash install_bisheng_jdk.sh --offline ."
