# VerMan — 轻量级版本管理系统

基于 **Go + Gin + SQLite + Vue 3 + Element Plus** 的一体化版本管理工具。

- **管理端**：分支与版本的增删改查（弹窗式交互 + 分页列表）
- **用户端**：水平时间轴视图（SVG 绘制，展示分支派生关系与版本发布节点）

## 快速启动

```bash
go build -o verman .
./verman
# 浏览器打开 http://localhost:8080
```

数据库文件 `verman.db` 自动创建于当前目录，环境变量 `VERMAN_DB` 可自定义路径。

## 技术栈

| 层 | 技术 | 说明 |
|----|------|------|
| 后端框架 | Gin v1 | HTTP 路由 + 中间件 |
| 数据库 | SQLite (modernc.org/sqlite) | 纯 Go，零依赖，内嵌 |
| 前端框架 | Vue 3 + Vite | SPA |
| UI 库 | Element Plus | 弹窗、表格、分页、日期选择等 |
| 图表 | 纯 SVG | 时间轴视图，无第三方图表依赖 |
| 构建 | Go embed | 前端 dist/ 编译进二进制，单文件部署 |

## 页面结构

| 页面 | 路由 | 说明 |
|------|------|------|
| 仪表盘 | `/#/` | 统计卡片 + 分支树 + 最近 10 个版本 |
| 分支管理 | `/#/branches` | 分页表格 + 弹窗式新建/编辑/详情 |
| 版本管理 | `/#/versions` | 筛选 + 分页表格 + 弹窗式新建/编辑/详情 |
| 时间轴视图 | `/#/timeline` | 用户端：水平 git-graph，SVG 绘制 |

## 核心概念

### 分支 (Branch)
代码分支，通过 `parent_branch_id` 自引用形成树形派生关系。`pulled_at` 记录实际拉取时间（用户手动设置，区别于系统录入时间 `created_at`）。

### 版本 (Version)
每次发布产出一条版本记录，绑定到某个分支。版本号为**自由格式字符串**（`v1.2.3`、`V2.0-Release`、`2024Q1-SP1` 均可）。`build_time` 由用户手动设置。

### 产品 (Product)
产品名称存储在版本的 `product_name` 字段，不单独建表。同一分支可为不同产品出版本。

## 项目结构

```
ver-manager/
├── main.go                     # 入口：DB 初始化 → 注册 API 路由 → serveSPA 中间件
├── model/model.go              # Branch / Version / DateTime / VersionQuery
├── db/database.go              # SQLite 连接 + 建表 + 兼容旧表 ALTER TABLE
├── repo/                       # 数据仓库层
│   ├── branch.go               # 分支 CRUD + 分页列表 + 统计
│   └── version.go              # 版本 CRUD + 条件筛选 + 分页 + 统计
├── handler/                    # REST API（JSON 入/出）
│   ├── dashboard.go            # GET /api/dashboard
│   ├── branch.go               # CRUD /api/branches
│   └── version.go              # CRUD /api/versions + /latest
└── frontend/                   # Vue 3 SPA
    ├── src/
    │   ├── App.vue             # 侧边栏 + 路由出口
    │   ├── router/index.js     # 4 条路由（Hash 模式）
    │   ├── api/                # Axios 封装
    │   │   ├── index.js        # 实例 + 响应拦截
    │   │   ├── branch.js       # 分支 API（含分页 + 全量两个版本）
    │   │   └── version.js      # 版本 API + dashboard
    │   ├── views/
    │   │   ├── Dashboard.vue   # 仪表盘
    │   │   ├── BranchList.vue  # 分支管理（弹窗 CRUD + 分页）
    │   │   ├── VersionList.vue # 版本管理（弹窗 CRUD + 筛选 + 分页）
    │   │   └── ClientView.vue  # 时间轴视图（SVG）
    │   └── components/
    │       └── BranchTreeNode.vue  # 分支树递归组件
    └── dist/                   # Vite 构建产物（go:embed）
```

## 数据库

### branches
```sql
CREATE TABLE branches (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    name             TEXT    NOT NULL UNIQUE,           -- 全局唯一
    parent_branch_id INTEGER REFERENCES branches(id),  -- NULL = 根分支
    branch_type      TEXT    NOT NULL DEFAULT 'custom'  -- main|release|feature|hotfix|custom
                        CHECK(branch_type IN ('main','release','feature','hotfix','custom')),
    description      TEXT    DEFAULT '',
    is_active        INTEGER DEFAULT 1,                 -- 0=软删除
    pulled_at        TEXT    DEFAULT NULL,              -- 实际拉取时间（用户手动设置，可空）
    created_at       TEXT    DEFAULT (datetime('now','localtime')),
    updated_at       TEXT    DEFAULT (datetime('now','localtime'))
);
```

### versions
```sql
CREATE TABLE versions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    branch_id       INTEGER NOT NULL REFERENCES branches(id),
    product_name    TEXT    NOT NULL,                    -- 产品名称
    version_number  TEXT    NOT NULL,                    -- 自由格式版本号
    description     TEXT    DEFAULT '',
    release_notes   TEXT    DEFAULT '',
    build_time      TEXT    DEFAULT (datetime('now','localtime')),  -- 用户手动设置
    commit_hash     TEXT    DEFAULT '',                  -- Git SHA
    artifact_url    TEXT    DEFAULT '',                  -- 构建产物链接
    status          TEXT    DEFAULT 'draft'              -- draft|released|deprecated|revoked
                        CHECK(status IN ('draft','released','deprecated','revoked')),
    created_at      TEXT    DEFAULT (datetime('now','localtime')),
    UNIQUE(branch_id, version_number)                   -- 同一分支版本号唯一
);
```

## API 接口

全部前缀 `/api`。列表接口返回 `{"data":[...], "total":N, "page":P, "page_size":S}`。

### 仪表盘
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/dashboard` | 统计 + 全部分支 + 最近 10 个版本 |

### 分支
| 方法 | 路径 | 请求体 | 说明 |
|------|------|--------|------|
| GET | `/api/branches?page=1&page_size=20` | - | 分页列表 |
| POST | `/api/branches` | `{name*, parent_branch_id?, branch_type?, description?, pulled_at?}` | 创建 |
| GET | `/api/branches/:id` | - | 详情 |
| PUT | `/api/branches/:id` | `{name?, parent_branch_id?, ...}` | 更新 |
| DELETE | `/api/branches/:id` | - | 停用 |

### 版本
| 方法 | 路径 | 请求体 | 说明 |
|------|------|--------|------|
| GET | `/api/versions?page=1&page_size=20&branch_id=&product=&status=` | - | 筛选+分页 |
| GET | `/api/versions/latest?product=` | - | 各分支最新版本 |
| POST | `/api/versions` | `{branch_id*, product_name*, version_number*, build_time*, status?, ...}` | 创建 |
| GET | `/api/versions/:id` | - | 详情 |
| PUT | `/api/versions/:id` | `{branch_id?, product_name?, version_number?, build_time?, ...}` | 更新 |

\* = 必填。版本接口兼容旧 `?limit=999` 参数。

## 关键设计

### DateTime 自定义类型
SQLite 时间存为 TEXT。`model.DateTime` 实现 `sql.Scanner` / `driver.Valuer` / `json.Marshaler` / `json.Unmarshaler`，统一 `"2006-01-02 15:04:05"` 格式。可为 `*DateTime`（nullable）。

### 版本号：自由格式字符串
不做语义化拆解，用户任意输入（`v1.2.3`、`V2.0-Release`、`2024Q1-SP1`）。

### 前端内嵌
`//go:embed frontend/dist` + 中间件 `serveSPA` 用 `fs.ReadFile` 直接读文件（而非 `http.FileServer`，避免 301 重定向）。

### 弹窗式交互
新建/编辑/详情全部 `el-dialog`，主页面只保留列表。减少页面跳转。

### 时间轴：SVG 绘制
X 轴=时间，Y 轴=分支（父在上）。版本为彩色圆点（绿=released / 橙=draft / 灰=deprecated / 红=revoked），分支派生为竖虚线。

### 分页
分支和版本列表均支持 `el-pagination`（10/20/50 条/页），筛选条件变化时自动回到第 1 页。

### 时间由用户设置
`pulled_at`（分支拉取时间）和 `build_time`（版本构建时间）均为用户手动输入，前端使用 `el-date-picker`。

## 数据库迁移策略

`db/database.go` 的 `migrate()`：
1. 执行 CREATE TABLE IF NOT EXISTS（新安装）
2. 执行 ALTER TABLE ADD COLUMN（旧库升级，忽略"列已存在"错误）
3. SQL 用 `strings.Split(schema, ";")` 拆分逐条执行

## 开发与构建

```bash
# 开发
go run .                              # 后端 :8080
cd frontend && npm run dev            # 前端 :5173（代理 /api → :8080）

# 生产
cd frontend && npm run build          # 前端 → dist/
cd .. && go build -o verman .         # 后端（内嵌 dist）
./verman                              # 单文件运行
```
