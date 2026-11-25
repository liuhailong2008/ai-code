#!/usr/bin/env bash
#
# 在可联网的机器执行，把内网所需的毕昇 JDK 安装包下载到本脚本所在目录。
# 下载完成后，将整个 jdk_install 目录上传到内网文件服务器即可。
#
# 用法：
#   ./download.sh                    # 仅下载首选版 25.0.5（x64 + aarch64）
#   ./download.sh --with-rollback    # 额外下载回滚基线 25.0.4
#
set -euo pipefail

BASE="${BASE:-https://mirrors.huaweicloud.com/kunpeng/archive/compiler/bisheng_jdk}"
VERSIONS=("bisheng-jdk-25.0.5-b16")
ARCHES=("x64" "aarch64")

[[ "${1:-}" == "--with-rollback" ]] && VERSIONS+=("bisheng-jdk-25.0.4-b15")

cd "$(dirname "$0")"
log() { printf '\033[32m[INFO ]\033[0m %s\n' "$*"; }
die() { printf '\033[31m[ERROR]\033[0m %s\n' "$*" >&2; exit 1; }

for ver in "${VERSIONS[@]}"; do
    for arch in "${ARCHES[@]}"; do
        pkg="${ver}-linux-${arch}.tar.gz"
        if [[ -f "$pkg" && -f "$pkg.sha256" ]] && sha256sum -c "$pkg.sha256" >/dev/null 2>&1; then
            log "已存在且校验通过，跳过：$pkg"
            continue
        fi
        log "下载：$pkg"
        curl -fL --progress-bar -o "$pkg" "${BASE}/${pkg}" || die "下载失败：${BASE}/${pkg}"
        curl -fsSL -o "$pkg.sha256" "${BASE}/${pkg}.sha256" || die "校验文件下载失败：$pkg.sha256"
        sha256sum -c "$pkg.sha256" || die "SHA256 校验失败：$pkg"
        log "完成：$pkg"
    done
done

# 额外的所需文件放在 extra/ 下（如应用 jar、配置文件），一并打包上传
if [[ -d extra ]]; then
    log "extra/ 目录内容："
    ls -lh extra | sed 's/^/    /'
fi

log "全部文件就绪："
ls -lh . | sed 's/^/    /'
