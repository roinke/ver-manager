# VerMan 部署文档

## 快速部署（三选一）

### 方案 A：Docker Compose（推荐，本地/测试）

前置条件：Docker + Docker Compose 已安装。

```bash
# 1. 构建并启动
docker compose up -d

# 2. 检查状态
docker compose ps

# 3. 访问
# http://localhost:8080
```

数据持久化在 Docker Volume `verman-data` 中，重启不丢失。

```bash
# 查看日志
docker compose logs -f

# 停止
docker compose down

# 停止并删除数据卷（⚠️ 数据库将丢失）
docker compose down -v
```

### 方案 B：直接构建镜像

```bash
# 构建
docker build -t verman:latest .

# 运行
docker run -d \
  --name verman \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  verman:latest

# 访问 http://localhost:8080
```

数据库文件 `verman.db` 将保存在宿主机的 `./data/` 目录下。

### 方案 C：Docker Compose + Nginx 反向代理（生产推荐）

目录结构：
```
verman-deploy/
├── docker-compose.yml
├── nginx/
│   └── default.conf
└── data/               # 数据库目录（自动创建）
```

**docker-compose.yml**：

```yaml
version: "3.8"

services:
  verman:
    image: verman:latest
    container_name: verman
    restart: unless-stopped
    volumes:
      - ./data:/data
    environment:
      - PORT=8080
      - VERMAN_DB=/data/verman.db
      - TZ=Asia/Shanghai
    networks:
      - verman-net
    # 不暴露端口到宿主机，仅 nginx 可访问

  nginx:
    image: nginx:1.27-alpine
    container_name: verman-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/default.conf:/etc/nginx/conf.d/default.conf:ro
      - ./ssl:/etc/nginx/ssl:ro          # 可选：SSL 证书
      - nginx-logs:/var/log/nginx
    depends_on:
      - verman
    networks:
      - verman-net

networks:
  verman-net:

volumes:
  nginx-logs:
```

**nginx/default.conf**：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 可选：重定向到 HTTPS
    # return 301 https://$host$request_uri;

    client_max_body_size 10m;

    location / {
        proxy_pass http://verman:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# 可选：HTTPS
# server {
#     listen 443 ssl http2;
#     server_name your-domain.com;
#
#     ssl_certificate     /etc/nginx/ssl/fullchain.pem;
#     ssl_certificate_key /etc/nginx/ssl/privkey.pem;
#
#     client_max_body_size 10m;
#
#     location / {
#         proxy_pass http://verman:8080;
#         proxy_set_header Host $host;
#         proxy_set_header X-Real-IP $remote_addr;
#         proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
#         proxy_set_header X-Forwarded-Proto $scheme;
#     }
# }
```

## 配置说明

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | 服务监听端口 |
| `VERMAN_DB` | `verman.db` | SQLite 数据库文件路径 |
| `TZ` | `UTC` | 时区（容器内用），如 `Asia/Shanghai` |

### 端口

| 端口 | 用途 |
|------|------|
| `8080` | VerMan HTTP 服务（API + 前端 SPA） |

### 卷挂载

| 容器路径 | 说明 |
|----------|------|
| `/data` | SQLite 数据库文件所在目录 |

## 镜像仓库推送

```bash
# 1. 构建
./build.sh

# 2. 构建并打版本号
./build.sh -t v1.0.0

# 3. 构建并推送到仓库
DOCKER_REGISTRY=ghcr.io/youruser ./build.sh -t v1.0.0 -p
```

`build.sh` 完整参数：

```
用法: ./build.sh [选项]

选项:
  -t, --tag TAG      版本标签（如 v1.0.0）
  -p, --push         构建后推送到镜像仓库
  -r, --registry URL 镜像仓库地址（也可用 DOCKER_REGISTRY 环境变量）
  -h, --help         显示帮助
```

## 多架构构建（可选）

如需支持 ARM64（如 Apple Silicon / 树莓派），用 `docker buildx`：

```bash
# 创建 builder
docker buildx create --name multiarch --use

# 多架构构建 + 推送
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ghcr.io/youruser/verman:latest \
  -t ghcr.io/youruser/verman:v1.0.0 \
  --push \
  .
```

## 数据备份

SQLite 是单文件数据库，备份极其简单：

```bash
# Docker Compose 部署
docker compose exec verman cp /data/verman.db /data/verman.db.bak
docker compose cp verman:/data/verman.db.bak ./verman-backup-$(date +%Y%m%d).db

# 或直接复制宿主机上的文件
cp ./data/verman.db ./verman-backup-$(date +%Y%m%d).db
```

恢复：

```bash
docker compose down
cp ./verman-backup-YYYYMMDD.db ./data/verman.db
docker compose up -d
```

建议加入 cron 定时任务：

```bash
# 每天凌晨 3 点备份
0 3 * * * cp /path/to/data/verman.db /path/to/backups/verman-$(date +\%Y\%m\%d).db
```

## 升级指南

```bash
# 1. 拉取新镜像
docker compose pull

# 2. 备份数据库（安全起见）
cp ./data/verman.db ./data/verman.db.bak.$(date +%Y%m%d)

# 3. 滚动重启
docker compose up -d --no-deps verman

# 4. 验证
curl http://localhost:8080/api/dashboard
```

## 资源限制

Docker Compose 中可以加资源约束：

```yaml
services:
  verman:
    # ...
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: "1.0"
```

典型资源消耗：内存 ~30MB，CPU 负载极低。

## 故障排查

```bash
# 查看日志
docker compose logs -f verman

# 进入容器
docker compose exec verman sh

# 检查数据库
docker compose exec verman ls -la /data/

# 端口冲突
lsof -i :8080

# 强制重建
docker compose down
docker compose build --no-cache
docker compose up -d
```
