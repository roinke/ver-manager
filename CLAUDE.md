# CLAUDE.md — VerMan 项目结构说明书

## 项目概述

单文件部署的版本管理系统。管理端（分支/版本 CRUD，侧边栏布局）+ 用户端（水平时间轴，全屏独立页面）。

`App.vue` 通过 `isTimeline` 计算属性检测路由，`/timeline` 时跳过侧边栏直接全屏渲染 `router-view`。

- 后端：Go + Gin + SQLite（modernc.org/sqlite，纯 Go）
- 前端：Vue 3 + Element Plus（编译后 go:embed 嵌入二进制）
- 产物：单个 `verman` 可执行文件

## 架构分层

```
HTTP 请求 → Gin Router
  ├── /api/*     → handler/   (JSON API)
  │     └── handler 调用 repo/ → repo 用 db.DB (database/sql)
  └── 其他路径    → serveSPA 中间件 (从 embed.FS 读文件返回)
```

## 目录与职责

```
main.go                 DB 初始化 → 注册 API → serveSPA 中间件 → http.Listen
model/model.go          Branch, Version, DateTime, VersionQuery 结构体
db/database.go          Init/Close/migrate（建表+ALTER TABLE兼容旧库）
repo/branch.go          分支仓库：Create, GetByID, GetByName, ListBranches(分页), Update, Deactivate, Count
repo/version.go         版本仓库：Create, GetByID, ListVersions(筛选+分页), GetLatestBy*, Update, Count
handler/dashboard.go    GET /api/dashboard → 聚合统计
handler/branch.go       /api/branches CRUD + parseTime 辅助
handler/version.go      /api/versions CRUD + GetLatestVersions（兼容 page/page_size 和 limit 两种参数）
frontend/src/router/    Hash 模式，4 条路由：/ /branches /versions /timeline
frontend/src/api/       index.js(axios实例+拦截器), branch.js, version.js
frontend/src/views/
  Dashboard.vue         统计卡片 + 分支树(BranchTreeNode) + 最近版本表
  BranchList.vue        分页表格 + 新建/编辑/详情 el-dialog 弹窗
  VersionList.vue       筛选栏 + 分页表格 + 新建/编辑/详情 el-dialog 弹窗
  ClientView.vue        全屏暗色 SVG 水平时间轴（独立页面，无管理框架）
frontend/src/components/
  BranchTreeNode.vue    递归分支树组件
```

## 数据模型

### Branch
```
ID             int64      主键
Name           string     分支名（UNIQUE）
ParentBranchID *int64     父分支ID（nil=根，自引用形成树）
BranchType     string     main|release|feature|hotfix|custom
Description    string
IsActive       bool       软删除标记
PulledAt       *DateTime  用户手动设置的拉取时间（可为nil）
CreatedAt      DateTime   系统录入时间
UpdatedAt      DateTime
```

### Version
```
ID            int64    主键
BranchID      int64    外键→branches.id
ProductName   string   产品名称（不建表）
VersionNumber string   自由格式版本号（不拆解）
Description   string
ReleaseNotes  string
BuildTime     DateTime 用户手动设置的构建时间
CommitHash    string   Git SHA
ArtifactURL   string   构建产物链接
Status        string   draft|released|deprecated|revoked
CreatedAt     DateTime
BranchName    string   JOIN填充，非DB字段
```

### VersionQuery
```go
type VersionQuery struct {
    BranchID    *int64
    ProductName string
    Status      string
    Limit       int
    Offset      int
}
```

## DateTime 类型

`model/model.go`：`type DateTime time.Time`

实现接口：`sql.Scanner` | `driver.Valuer` | `json.Marshaler` | `json.Unmarshaler`

- 存储格式：`"2006-01-02 15:04:05"`
- 可空字段用 `*DateTime`，nil → JSON null / DB NULL
- `Now()` 返回当前时间

## API 规范

### 响应格式
```json
// 单条/创建/更新
{"data": {...}}

// 列表（分页）
{"data": [...], "total": 10, "page": 1, "page_size": 20}

// 错误
{"msg": "错误描述"}
```

### 分页参数
- `page` + `page_size`：标准分页（默认 20，上限 100）
- 版本接口兼容旧 `?limit=999` 参数（ClientView 使用）
- 分支接口 `getAllBranches()` 用 `page_size=9999` 获取全量（下拉框用）

### 必填字段
- 创建分支：`name`
- 创建版本：`branch_id`, `product_name`, `version_number`, `build_time`
- `parent_branch_id=0` 或 null 均视为"无父分支"

## 前端设计要点

### 双布局设计
- `App.vue` 用 `computed(() => route.path === '/timeline')` 判断当前模式
- **管理端**（`/`、`/branches`、`/versions`）：`el-container` → `el-aside`（侧边栏）+ `el-main`（内容）
- **用户端**（`/timeline`）：直接 `<div style="100vw;100vh"><router-view/></div>`，无任何管理框架
- 侧边栏不包含时间轴入口（对管理用户隐藏）

### 弹窗式 CRUD
- 主页面仅展示表格 + 顶部操作按钮
- 新建/编辑/详情统一使用 `el-dialog`
- BranchList：表格 + 3 个弹窗（新建/edit共用表单 + 详情，不含关联版本表）
- VersionList：筛选栏 + 表格 + 3 个弹窗 + el-pagination
- **UI 安全规范**：前端不展示数据库 ID（id）、创建时间（created_at）、更新时间（updated_at），仅展示业务字段

### 分页
- 列表接口返回 `total/page/page_size`
- `el-pagination` 组件，支持 10/20/50 条/页
- 筛选条件变化时 `reload()` 重置到第 1 页

### ClientView 时间轴（全屏独立页面，2026-06-17 重写）
- `App.vue` 检测 `/timeline` 路由，`overflow:hidden` 防止双滚动条，全屏渲染
- 暗色主题：背景 `#1a1b2e`，标题栏 `#22243a`，文本 `#c0c4e0`/`#606380`

**双层面板布局**：
```
┌─ 标题栏 ──────────────────────────────────────┐
│  标题 + 统计          │  缩放% + 重置视图按钮   │
├──────────┬────────────────────────────────────┤
│ 分支名面板 │  SVG 时间轴区域                    │
│ (150px)   │  overflow-x:auto（水平滚动）        │
│ flex-     │  overflow-y:hidden                 │
│ shrink:0  │                                    │
├──────────┴────────────────────────────────────┤
└─ 图例栏 ──────────────────────────────────────┘
```
- 左侧分支名面板与 SVG 区域共享垂直滚动（外层 `overflow-y:auto`），分支名不随水平滚动
- 分支名使用绝对定位的 `<div>`，Y 坐标与 SVG 分支线对齐，颜色一致

**缩放机制**：
- `zoom` ref（0.3–8.0），滚轮以鼠标位置为中心缩放（每次 15%），不依赖 `pending tasks```
- 核心思路：zoom 仅乘入 X 坐标计算（`xPos = leftMargin + ratio × chartWidth × zoom`），**所有视觉属性常量**（font-size、r、stroke-width 不除以 zoom）
- 标题栏显示 `Math.round(zoom * 100)%` + "重置视图"按钮（zoom=1 + 回到一年视图）

**默认一年视图**：
- 找最新 `build_time` → `latestTime`，默认窗口 `[latestTime - 365天, latestTime + 30天]`
- 加载后 `scrollToDefaultView()` 设定 scrollLeft
- 更早数据向左滚动或缩小查看

**铺满屏幕**：
- `ResizeObserver` 监听主区域容器 → `containerHeight`
- `branchSpacing = clamp(64, (containerHeight - 70) / branchCount, 120)`
- `svgHeight = max(containerHeight, topMargin + branchCount × branchSpacing + 20)`

**版本标签**：35° 斜排（左下→右上），`rotate(-35)` 绕版本点中心，`text-anchor="start"`

**鼠标拖拽平移**：pointerdown/move/up 事件，仅左键

**悬浮提示**（120ms 延迟隐藏防抖）：
- 显示：产品名 + 版本号、构建时间、版本描述
- 不显示：status、commit_hash 等非关键字段

- 纯 SVG，无第三方图表库
- X 轴 = 时间（左→右），Y 轴 = 分支（上→下，父在子上方）
- 分支排列：根分支 → 子分支（递归，确保父在子前）
- 版本点为 `r=6` 圆，颜色映射：released=绿 draft=橙 deprecated=灰 revoked=红
- 派生关系为竖虚线，拉取点为 `r=4` 小圆
- 颜色循环：8 色预定义调色板
- 时间刻度自适应：`count = chartWidth / 120`，宽间距 "年-月"，密间距 "月/日"
- 数据源：`getAllBranches()` + `getVersions({limit:9999})`

### SPA 服务
`main.go` 的 `serveSPA` 中间件：
```
请求 → /api 前缀 → c.Next()
     → 其他 → fs.ReadFile(spaFS, path)
         → 找到 → c.Data(contentType, data)
         → 未找到 → c.Data("text/html", index.html)  // SPA fallback
```
Content-Type 由 `mime.TypeByExtension(ext)` 判断。

## 修改指南

### 新增数据库字段
1. `db/database.go`：CREATE TABLE 加列 + ALTER TABLE 兼容迁移
2. `model/model.go`：结构体加字段（时间用 `DateTime` 或 `*DateTime`）
3. `repo/`：INSERT/SELECT/UPDATE 加列 + scan 加参数
4. `handler/`：请求体加字段 + 解析
5. 前端表单/表格：加对应表单项和列

### 新增 API
1. `repo/` 添加数据函数
2. `handler/` 添加 Gin handler
3. `main.go` 的 api 路由组注册

### 新增前端页面
1. `views/` 创建 `.vue`
2. `router/index.js` 注册
3. `App.vue` 侧边栏加 `el-menu-item`
4. 需要 API 的在 `api/` 添加封装
5. `npm run build` → `go build`

## AI 接手清单

1. 先读 README.md（产品）→ CLAUDE.md（架构）
2. 数据流：前端 → handler(JSON) → repo(SQL) → db
3. 改模型要同步 5 层：db 建表 / model 结构体 / repo SQL / handler 请求体 / 前端表单
4. 前端 SPA，路由由 Vue Router 管理，后端仅 `/api` + 静态文件
5. 必须先 `npm run build` 再 `go build`（内嵌 dist）
