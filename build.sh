#!/usr/bin/env bash
# ============================================================
# VerMan Docker 镜像构建脚本
#
# 用法:
#   ./build.sh              # 构建镜像
#   ./build.sh -t v1.0      # 构建并打版本标签
#   ./build.sh -p           # 构建并推送到仓库
#   ./build.sh -t v1.0 -p   # 构建 + 打标签 + 推送
# ============================================================

set -euo pipefail

# ── 默认配置 ──
IMAGE_NAME="verman"
REGISTRY="${DOCKER_REGISTRY:-}"              # 如 ghcr.io/username 或 docker.io/username
VERSION_TAG=""
PUSH=false

# ── 颜色输出 ──
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# ── 解析参数 ──
while [[ $# -gt 0 ]]; do
    case "$1" in
        -t|--tag)
            VERSION_TAG="$2"
            shift 2
            ;;
        -p|--push)
            PUSH=true
            shift
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -h|--help)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  -t, --tag TAG      版本标签（如 v1.0.0）"
            echo "  -p, --push         构建后推送到镜像仓库"
            echo "  -r, --registry URL 镜像仓库地址"
            echo "  -h, --help         显示帮助"
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            exit 1
            ;;
    esac
done

# ── 构建镜像名 ──
LOCAL_TAG="${IMAGE_NAME}:latest"
if [[ -n "$REGISTRY" ]]; then
    REMOTE_TAG="${REGISTRY}/${IMAGE_NAME}:latest"
fi

log_info "开始构建 Docker 镜像..."

# ── 构建 ──
docker build \
    --platform linux/amd64 \
    -t "$LOCAL_TAG" \
    -f Dockerfile \
    .

log_info "构建完成: $LOCAL_TAG"

# ── 打版本标签 ──
if [[ -n "$VERSION_TAG" ]]; then
    docker tag "$LOCAL_TAG" "${IMAGE_NAME}:${VERSION_TAG}"
    log_info "已打标签: ${IMAGE_NAME}:${VERSION_TAG}"

    if [[ -n "$REGISTRY" ]]; then
        docker tag "$LOCAL_TAG" "${REGISTRY}/${IMAGE_NAME}:${VERSION_TAG}"
        log_info "已打远程标签: ${REGISTRY}/${IMAGE_NAME}:${VERSION_TAG}"
    fi
fi

if [[ -n "$REGISTRY" ]]; then
    docker tag "$LOCAL_TAG" "$REMOTE_TAG"
    log_info "已打远程标签: $REMOTE_TAG"
fi

# ── 推送 ──
if $PUSH; then
    if [[ -z "$REGISTRY" ]]; then
        log_error "推送前请先设置 REGISTRY 环境变量或使用 -r 参数指定仓库地址"
        log_error "示例: DOCKER_REGISTRY=ghcr.io/myuser ./build.sh -p"
        exit 1
    fi

    log_info "推送镜像到 $REGISTRY ..."
    docker push "$REMOTE_TAG"
    log_info "已推送: $REMOTE_TAG"

    if [[ -n "$VERSION_TAG" ]]; then
        docker push "${REGISTRY}/${IMAGE_NAME}:${VERSION_TAG}"
        log_info "已推送: ${REGISTRY}/${IMAGE_NAME}:${VERSION_TAG}"
    fi
fi

# ── 汇总 ──
echo ""
log_info "══════ 构建汇总 ══════"
echo "  本地镜像:  $LOCAL_TAG"
docker images "$IMAGE_NAME" --format "  大小:      {{.Size}}"
if [[ -n "$REGISTRY" ]]; then
    echo "  远程仓库:  $REGISTRY"
fi
