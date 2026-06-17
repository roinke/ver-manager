# 版本管理系统 — 需求记录

日期：2026-06-16（初始）→ 2026-06-17（时间轴重写 + UI 安全加固）

## 一、已完成并保留的需求

### 1. 新建分支时基于版本自动填充拉取时间
- **文件**：`frontend/src/views/BranchList.vue`
- **功能**：新建/编辑分支时，选择父分支后出现"基于版本"下拉框，列出该父分支的全部版本。选中版本后自动将版本的 `build_time` 填入分支的 `pulled_at`，用户仍可手动修改。
- **实现**：`watch(form.parent_branch_id)` 触发 `getVersions({ branch_id })` 加载版本列表；`onVersionSelect` 将 `version.build_time` 赋给 `form.pulled_at`。

### 2. 分支详情弹窗中版本列表分页
- **文件**：`frontend/src/views/BranchList.vue`、`repo/version.go`、`handler/version.go`、`handler/dashboard.go`
- **功能**：分支详情弹窗中的版本表格增加分页，默认每页 5 条，可选 5/10/20。
- **后端修复**：`CountVersions` 函数签名改为接受 `VersionQuery`，支持 `branch_id`/`product_name`/`status` 组合过滤，之前总数不随筛选条件变化。

### 3. 版本管理筛选栏分支下拉框宽度修复
- **文件**：`frontend/src/views/VersionList.vue`
- **功能**：版本管理页筛选栏的分支 `<el-select>` 增加 `style="width:200px"`，之前默认宽度太短导致选中分支名显示不全。

---

## 二、时间轴重写（2026-06-17）

> 原有 9 项时间轴增强（缩放平移、斜排标签、一年默认视图等）在 commit `3ad5de9` 后整体回退。2026-06-17 完全重写 `ClientView.vue`，以更清晰的架构重新实现核心功能。

### 1. 分支名固定在左侧
- **实现**：双层面板布局。左侧 150px `div` 面板（`flex-shrink:0`），分支名 `absolute` 定位与 SVG 分支线 Y 对齐；右侧 SVG 容器 `overflow-x:auto` 独立水平滚动。外层 `overflow-y:auto` 实现两者同步垂直滚动。
- **效果**：水平滚动时间轴时分支名始终可见。

### 2. 默认展示最近一年数据
- **实现**：`scrollToDefaultView()` 计算可视窗口 `[latestTime - 365天, latestTime + 30天]`，数据加载后 `nextTick` + 设置 `scrollLeft`。更早数据通过向左滚动或缩小查看。
- **关键变量**：`YEAR_MS = 365 * 86400000`，`PADDING_MS = 30 * 86400000`

### 3. 鼠标缩放，视觉元素大小不变
- **实现**：`zoom` ref（0.3–8.0），滚轮 `wheel` 事件以鼠标 X 为中心缩放（步长 15%）。zoom 仅乘入 X 坐标（`xPos = leftMargin + ratio × chartWidth(zoom)`），**font-size / r / stroke-width 全部常量**，不使用 SVG `transform:scale()`。
- **辅助**：缩放后 `nextTick` 调整 `scrollLeft` 保持鼠标下方时间点不动。标题栏显示百分比 + "重置视图"按钮。

### 4. 时间轴铺满整个屏幕
- **实现**：`ResizeObserver` 监听主区域高度 → `containerHeight`。`branchSpacing = clamp(64, (containerHeight - 70) / branchCount, 120)`。`svgHeight = max(containerHeight, topMargin + count × spacing + 20)`。
- **效果**：分支少时分距拉大填满屏幕，分支多时保持 64px 最小间距并触发垂直滚动。

### 5. 版本标签斜排显示
- **实现**：版本名 `rotate(-35)` 绕版本点中心，`text-anchor="start"`，"左下→右上"方向，减少水平重叠。

### 6. 鼠标拖拽平移
- **实现**：`pointerdown/move/up` 事件，仅左键，调整 `scrollLeft`。`cursor: grabbing` 视觉反馈。

### 7. 悬浮提示优化
- **实现**：仅展示产品名 + 版本号、构建时间、版本描述。120ms 延迟隐藏防抖。
- **不展示**：status、commit_hash 等非关键字段。

### 8. 时间刻度自适应
- **实现**：`tickCount = chartWidth / 120`，从 `[1h, 1d, 7d, 30d, 90d, 180d, 365d]` 中选择最接近的间隔。宽间距 "年-月"，密间距 "月/日"。

---

## 三、UI 安全加固（2026-06-17）

| 改动 | 文件 | 说明 |
|------|------|------|
| 移除主表格 ID 列 | `BranchList.vue`, `VersionList.vue` | 分支和版本列表表格不再展示数据库 ID |
| 移除详情弹窗 ID 行 | `BranchList.vue`, `VersionList.vue` | 详情弹窗不再展示 `detail.id` |
| 移除关联版本表 ID 列 | `BranchList.vue` | 分支详情中版本子表不再展示版本 ID |
| 移除创建/更新时间 | `BranchList.vue`, `VersionList.vue` | 详情弹窗不再展示 `created_at` / `updated_at` |
| 移除分支详情中的版本列表 | `BranchList.vue` | 分支详情弹窗不再包含"该分支的版本"区块（含表格+分页+空状态），相关变量 `detailVersions`/`detailLoading`/`detailPage`/`detailPageSize`/`detailTotal`/`loadDetailVersions` 一并清理 |

**原则**：前端仅展示业务字段，不暴露数据库元数据（ID、时间戳）。

---

## 四、后端改进

| 改动 | 文件 |
|------|------|
| `CountVersions` 支持 branch_id/status 过滤 | `repo/version.go` |
| `ListVersions` handler 传入完整 VersionQuery | `handler/version.go` |
| Dashboard 适配新 CountVersions 签名 | `handler/dashboard.go` |

---

## 五、App.vue 改动

| 改动 | 说明 |
|------|------|
| 时间轴容器增加 `overflow:hidden` | 防止 `/timeline` 路由下 body 出现双滚动条 |

---

## 六、未来可考虑的需求

1. 时间轴进一步的交互优化（触屏支持、动画过渡）
2. 分支/版本的批量导入导出
3. 版本对比功能
4. 搜索/全文检索
5. 用户认证与权限
