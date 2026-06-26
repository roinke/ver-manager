# ============================================================
# VerMan Docker 镜像 — 多阶段构建
# ============================================================

# ── 阶段 1：前端构建 ──
FROM node:22-alpine AS frontend-builder
WORKDIR /build/frontend

# 安装依赖（利用 Docker 层缓存）
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

# 复制源码并构建
COPY frontend/ ./
RUN npm run build

# ── 阶段 2：后端构建 ──
FROM golang:1.25-alpine AS backend-builder
WORKDIR /build

# 配置 Go 国内代理（解决 proxy.golang.org 不可达）
ENV GOPROXY=https://goproxy.cn,direct

# 安装 Go 依赖（利用 Docker 层缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制后端源码
COPY *.go ./
COPY model/     model/
COPY db/        db/
COPY repo/      repo/
COPY handler/   handler/

# 从前端构建阶段复制 dist
COPY --from=frontend-builder /build/frontend/dist/ frontend/dist/

# 编译静态链接二进制
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o verman .

# ── 阶段 3：运行镜像 ──
FROM alpine:3.21

# 安全加固：非 root 运行
RUN addgroup -g 1000 verman && \
    adduser -u 1000 -G verman -s /bin/sh -D verman

# 时间同步
RUN apk add --no-cache tzdata ca-certificates

# 数据目录
RUN mkdir -p /data && chown verman:verman /data

COPY --from=backend-builder /build/verman /usr/local/bin/verman

# 默认从 /data 目录读写数据库
ENV VERMAN_DB=/data/verman.db
ENV PORT=8080

EXPOSE 8080

USER verman
WORKDIR /data

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O- http://localhost:${PORT} || exit 1

ENTRYPOINT ["verman"]
