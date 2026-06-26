# VerMan 运维操作手册

> 基于 2026-06-26 实际操作验证编写，涵盖从零到生产部署的全部步骤。

## 目录

- [一、产物总览](#一产物总览)
- [二、环境准备](#二环境准备)
- [三、本地开发运行](#三本地开发运行)
- [四、Docker 镜像构建](#四docker-镜像构建)
- [五、Docker 容器部署](#五docker-容器部署)
- [六、裸二进制部署](#六裸二进制部署)
- [七、镜像导出与跨服务器迁移](#七镜像导出与跨服务器迁移)
- [八、Docker Compose 部署](#八docker-compose-部署)
- [九、Nginx 反向代理](#九nginx-反向代理)
- [十、数据备份与恢复](#十数据备份与恢复)
- [十一、升级流程](#十一升级流程)
- [十二、故障排查](#十二故障排查)
- [十三、关键文件清单](#十三关键文件清单)

---

## 一、产物总览

| 产物 | 文件 | 大小 | 适用场景 |
|------|------|------|---------|
| **静态二进制** | `verman-static` | 26 MB | 任意 Linux amd64，零依赖直接跑 |
| **Docker tar 包** | `verman-latest.tar` | 14 MB | 拷贝到其他服务器 `docker load` 导入 |
| **Docker 镜像** | `verman:latest` | 52.2 MB (解压后) | 本地 Docker 运行 |
| 开发版二进制 | `verman` | 37 MB | 动态链接，仅限开发机 |

---

## 二、环境准备

### 2.1 Docker 安装（snap 版）— 本次验证环境

```bash
# 确认安装方式
which docker
# /snap/bin/docker  → snap 版

snap list | grep docker
```

### 2.2 配置 Docker 镜像加速（国内必备）

**snap 版 Docker**（配置文件路径不同）：

```bash
# 查看当前配置
sudo cat /var/snap/docker/current/config/daemon.json

# 写入镜像加速
sudo tee /var/snap/docker/current/config/daemon.json <<'EOF'
{
    "log-level":        "error",
    "registry-mirrors": [
        "https://mirror.ccs.tencentyun.com"
    ]
}
EOF

# 重启生效
sudo snap restart docker

# 验证
sudo docker info 2>&1 | grep -A 3 "Registry Mirrors"
# Registry Mirrors:
#   https://mirror.ccs.tencentyun.com/
```

**systemd 版 Docker**（`apt/yum` 安装）：

```bash
sudo tee /etc/docker/daemon.json <<'EOF'
{
    "registry-mirrors": [
        "https://mirror.ccs.tencentyun.com"
    ]
}
EOF

sudo systemctl daemon-reload
sudo systemctl restart docker
```

### 2.3 Go 代理配置（本地开发时用）

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

Dockerfile 中已内置（见第四节），无需额外操作。

---

## 三、本地开发运行

```bash
# 1. 前端构建（产物写入 frontend/dist/）
cd frontend && npm install && npm run build && cd ..

# 2. Go 编译
go build -o verman .

# 3. 启动
./verman
# ✅ 已创建默认 master 分支
# 🚀 VerMan 已启动: http://localhost:8080

# 4. 验证
curl http://localhost:8080/api/dashboard
```

> `go run .` 也可直接启动，省去编译步骤。

---

## 四、Docker 镜像构建

### 4.1 Dockerfile 说明（三阶段构建）

```
┌─ 阶段1: node:22-alpine ─────────────────────┐
│  npm ci → npm run build                      │
│  产物: frontend/dist/                        │
└──────────────────┬───────────────────────────┘
                   │ COPY --from=frontend-builder
┌─ 阶段2: golang:1.25-alpine ─────────────────┐
│  ENV GOPROXY=https://goproxy.cn,direct       │ ← 国内必需
│  go mod download → CGO_ENABLED=0 go build    │
│  产物: verman (静态链接, 无libc依赖)           │
└──────────────────┬───────────────────────────┘
                   │ COPY --from=backend-builder
┌─ 阶段3: alpine:3.21 ────────────────────────┐
│  + tzdata + ca-certificates                 │
│  用户: verman (uid 1000, 非root)             │
│  最终镜像: ~52MB (压缩后 14MB tar)            │
└─────────────────────────────────────────────┘
```

### 4.2 执行构建

```bash
# 方式 A：直接 docker build
sudo docker build -t verman:latest .

# 方式 B：用构建脚本（支持打标签和推送）
chmod +x build.sh
./build.sh                    # 仅构建
./build.sh -t v1.0.0          # 构建 + 版本标签
```

构建耗时约 50 秒（首次），后续利用 Docker 层缓存仅重编译变更部分。

### 4.3 验证镜像

```bash
sudo docker images verman
# REPOSITORY   TAG       IMAGE ID       SIZE
# verman       latest    b64096fc52e2   52.2MB

sudo docker history verman:latest
# 26.6MB  COPY /build/verman          ← 二进制
# 3.0MB   apk add tzdata ca-certificates
# 8.5MB   alpine:3.21 base
```

---

## 五、Docker 容器部署

### 5.1 创建数据目录

```bash
mkdir -p /opt/verman-data
chmod 777 /opt/verman-data          # ⚠️ 关键：容器内 uid 1000 需要写权限
```

### 5.2 启动容器

```bash
sudo docker run -d \
  --name verman \
  -p 8080:8080 \
  -v /opt/verman-data:/data \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  verman:latest
```

**参数解释：**

| 参数 | 说明 |
|------|------|
| `-d` | 后台运行 |
| `--name verman` | 容器名 |
| `-p 8080:8080` | 宿主机:容器 端口映射 |
| `-v /opt/verman-data:/data` | 数据库持久化目录（宿主机:容器） |
| `-e TZ=Asia/Shanghai` | 时区 |
| `--restart unless-stopped` | Docker 启动时自动启动容器 |

### 5.3 自定义数据库路径

```bash
# 数据库文件存到宿主机其他位置
docker run -d --name verman \
  -v /mnt/data/verman:/app/data \
  -e VERMAN_DB=/app/data/prod.db \
  -p 8080:8080 \
  verman:latest
```

环境变量 `VERMAN_DB` 和卷挂载配合使用，路径自由组合。

### 5.4 验证

```bash
sudo docker ps --filter name=verman
# STATUS 列应显示 "Up" 和 "(healthy)"

curl http://localhost:8080/api/dashboard
# {"branch_count":1,"branches":[...]...}

sudo docker logs verman
# ✅ 已创建默认 master 分支
# 🚀 VerMan 已启动: http://localhost:8080
```

---

## 六、裸二进制部署

不需要 Docker，直接把编译好的静态二进制拷贝到服务器。

### 6.1 编译静态二进制

```bash
# 在开发机上编译（纯静态，零依赖）
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o verman-static .

# 验证
file verman-static
# ELF 64-bit LSB executable, statically linked, stripped

ldd verman-static
# not a dynamic executable  ← 确认无动态库依赖

ls -lh verman-static
# 26 MB
```

### 6.2 部署

```bash
# 拷贝到目标服务器
scp verman-static user@目标IP:/usr/local/bin/verman

# 在目标服务器上
chmod +x /usr/local/bin/verman
mkdir -p /opt/verman-data

# 直接启动（数据库路径通过环境变量指定）
VERMAN_DB=/opt/verman-data/verman.db ./verman &

# 或写入 systemd 服务（推荐）
```

### 6.3 systemd 服务配置（可选）

```ini
# /etc/systemd/system/verman.service
[Unit]
Description=VerMan Version Manager
After=network.target

[Service]
Type=simple
User=verman
Environment=VERMAN_DB=/opt/verman-data/verman.db
Environment=PORT=8080
ExecStart=/usr/local/bin/verman
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now verman
sudo systemctl status verman
```

如需 ARM 架构（树莓派等），加 `GOARCH=arm64` 重新编译即可。

---

## 七、镜像导出与跨服务器迁移

目标服务器无需网络、无需镜像源，直接导入运行。

### 7.1 导出镜像

```bash
# 导出为 tar 文件
sudo docker save verman:latest -o /tmp/verman-latest.tar

# 查看大小
ls -lh /tmp/verman-latest.tar
# 14 MB
```

### 7.2 拷贝到目标服务器

```bash
scp /tmp/verman-latest.tar user@目标IP:/tmp/
```

### 7.3 目标服务器导入并启动

```bash
# 1. 导入（不联网）
docker load -i /tmp/verman-latest.tar

# 2. 确认
docker images verman
# REPOSITORY   TAG       SIZE
# verman       latest    52.2MB

# 3. 准备数据目录并启动
mkdir -p /opt/verman-data && chmod 777 /opt/verman-data

docker run -d \
  --name verman \
  -p 8080:8080 \
  -v /opt/verman-data:/data \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  verman:latest
```

**全程不需要拉取任何外部镜像。**

### 7.4 导出历史镜像（清理用）

```bash
# 删除旧镜像保留 tar 包
docker rmi verman:latest
# 需要时重新导入
docker load -i verman-latest.tar
```

---

## 八、Docker Compose 部署

适合本地测试或单机生产，配置写入文件便于管理。

### 8.1 启动

```bash
docker compose up -d
```

### 8.2 日常操作

```bash
docker compose ps           # 查看状态
docker compose logs -f      # 实时日志
docker compose restart      # 重启
docker compose down         # 停止并删除容器
docker compose down -v      # ⚠️ 同时删除数据卷（数据库丢失）
```

### 8.3 docker-compose.yml 关键配置

```yaml
services:
  verman:
    build: .
    image: verman:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - verman-data:/data           # Docker Volume（自动管理路径）
      # - ./data:/data              # 或绑定宿主机目录
    environment:
      - PORT=8080
      - VERMAN_DB=/data/verman.db
      - TZ=Asia/Shanghai

volumes:
  verman-data:

  # 如需指定宿主机路径：
  # verman-data:
  #   driver: local
  #   driver_opts:
  #     device: /opt/verman-data
  #     type: none
  #     o: bind
```

---

## 九、Nginx 反向代理

生产环境建议 Nginx 前置，处理静态资源缓存、SSL 终止、限流等。

```nginx
# /etc/nginx/conf.d/verman.conf
upstream verman {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name ver.example.com;

    client_max_body_size 10m;

    # 可选：全站 HTTPS
    # return 301 https://$host$request_uri;

    location / {
        proxy_pass http://verman;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Docker Compose 中集成 Nginx 见 `docker-compose.yml + nginx/` 组合部署（DEPLOY.md 方案 C）。

---

## 十、数据备份与恢复

SQLite 单文件，备份就是 `cp`。

### 10.1 手动备份

```bash
# Docker 部署
cp /opt/verman-data/verman.db /backup/verman-$(date +%Y%m%d).db

# 裸二进制部署
cp /opt/verman-data/verman.db /backup/verman-$(date +%Y%m%d).db
```

### 10.2 从容器内备份

```bash
docker exec verman cp /data/verman.db /data/verman.db.bak
docker cp verman:/data/verman.db.bak ./verman-backup.db
```

### 10.3 定时备份（cron）

```bash
# 每天凌晨 3 点
0 3 * * * cp /opt/verman-data/verman.db /backup/verman-$(date +\%Y\%m\%d).db

# 保留最近 30 天
0 3 * * * find /backup/ -name "verman-*.db" -mtime +30 -delete
```

### 10.4 恢复

```bash
# Docker
docker stop verman
cp /backup/verman-20260626.db /opt/verman-data/verman.db
docker start verman

# 裸二进制
pkill verman
cp /backup/verman-20260626.db /opt/verman-data/verman.db
VERMAN_DB=/opt/verman-data/verman.db ./verman &
```

---

## 十一、升级流程

```bash
# === Docker 部署升级 ===

# 1. 备份数据库
cp /opt/verman-data/verman.db /opt/verman-data/verman.db.bak.$(date +%Y%m%d)

# 2. 拉取/构建新镜像
docker pull verman:latest          # 从仓库拉取
# 或 docker build -t verman:latest . # 本地构建

# 3. 滚动替换
docker stop verman
docker rm verman
docker run -d --name verman \
  -p 8080:8080 \
  -v /opt/verman-data:/data \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  verman:latest

# 4. 验证
curl http://localhost:8080/api/dashboard

# === 裸二进制部署升级 ===

pkill verman
cp verman-static /usr/local/bin/verman
VERMAN_DB=/opt/verman-data/verman.db /usr/local/bin/verman &
```

---

## 十二、故障排查

### 容器不断重启

```bash
docker logs verman 2>&1 | tail -20
```

**常见原因和解决：**

| 错误信息 | 原因 | 解决 |
|----------|------|------|
| `unable to open database file (14)` | 宿主机挂载目录权限不对 | `chmod 777 /opt/verman-data` |
| `bind: address already in use` | 端口被占用 | 换端口 `-p 9090:8080` |
| `database disk image is malformed` | SQLite 文件损坏 | 从备份恢复 |

### 镜像拉取/模块下载超时

| 问题 | 原因 | 解决 |
|------|------|------|
| Docker Hub 超时 | `registry-1.docker.io` 不可达 | 配置镜像加速（见 2.2） |
| `go mod download` 超时 | `proxy.golang.org` 不可达 | Dockerfile 已设 `GOPROXY=https://goproxy.cn` |

### 健康检查失败

```bash
docker inspect verman | grep -A 10 Health
# 检查容器内 wget 是否可用
docker exec verman wget -q -O- http://localhost:8080
```

### 其他常用调试命令

```bash
docker logs -f verman                         # 实时日志
docker exec -it verman sh                     # 进入容器
docker exec verman ls -la /data/              # 检查数据库文件
docker inspect verman | grep -A 5 Mounts      # 查看卷挂载
lsof -i :8080                                 # 端口占用
```

---

## 十三、关键文件清单

| 文件 | 用途 |
|------|------|
| `Dockerfile` | 三阶段构建（内含 `GOPROXY` 国内代理） |
| `docker-compose.yml` | Compose 一键部署 |
| `.dockerignore` | 排除 node_modules、*.db 等 |
| `build.sh` | 构建脚本（支持 `-t` 标签、`-p` 推送） |
| `OPS.md` | 本文档 |

---

## 速查卡片

```bash
# 构建
sudo docker build -t verman:latest .
CGO_ENABLED=0 go build -ldflags="-s -w" -o verman-static .

# Docker 启动
mkdir -p /opt/verman-data && chmod 777 /opt/verman-data
sudo docker run -d --name verman -p 8080:8080 \
  -v /opt/verman-data:/data -e TZ=Asia/Shanghai \
  --restart unless-stopped verman:latest

# 导出/导入
sudo docker save verman:latest -o verman-latest.tar
docker load -i verman-latest.tar

# 裸机部署
scp verman-static user@server:/usr/local/bin/verman
ssh user@server "VERMAN_DB=/data/verman.db nohup /usr/local/bin/verman &"

# 每日备份
cp /opt/verman-data/verman.db /backup/verman-$(date +%Y%m%d).db

# 健康检查
curl http://localhost:8080/api/dashboard
```
