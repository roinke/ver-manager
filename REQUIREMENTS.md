# 版本管理系统 — 需求记录

日期：2026-06-16

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

## 二、已完成但已回退的需求（时间轴相关）

> 以下需求在 `ClientView.vue` 和 `App.vue` 上实现过，后续因体验问题整体回退到最初版本。保留作为后续参考。

### 1. 时间轴缩放与平移
- 鼠标滚轮缩放（以鼠标位置为中心），拖拽平移，缩放范围 30%~800%
- 标题栏显示缩放百分比 + 重置视图按钮

### 2. 版本标签斜排显示
- 版本名由原来的水平居中改为 35° 斜排（左下→右上），减少重叠

### 3. 默认展示最近一年
- 打开时间轴默认只显示最近一年的版本，更早数据通过向左拖拽查看
- 通过 `fitToRecentYear()` 计算初始 `zoom` 和 `panX`

### 4. 缩放时文字和几何元素保持屏幕大小不变
- 所有 `font-size`、`stroke-width`、`r` 使用 `基准值 / zoom` 动态计算
- 缩放只拉开间距，不改变元素视觉大小

### 5. 时间刻度自适应缩放
- 刻度数量 = `chartWidth × zoom / 120`，缩放越大刻度越密
- 标签格式自动切换：宽间距显示"年-月"，密间距显示"月/日"

### 6. 分支名固定在左侧
- 分支名独立于缩放分组，随纵向平移但不随水平平移滚动
- 加半透明背景条防止线条穿透

### 7. 纵向间距固定不随缩放变化
- 所有 Y 坐标除以 zoom，配合 `scale(zoom)` 后屏幕纵向位置恒定

### 8. SVG 铺满容器
- `svgHeight` 通过 `ResizeObserver` 动态跟踪容器高度，确保不留底部空白

### 9. 时间轴数据刷新
- `App.vue` 改为单一 `<router-view>` + `v-show` 控制侧边栏，`timelineKey` 强制重建组件
- `ClientView.vue` 添加 `watch(route.path)` 在每次导航到 `/timeline` 时重新加载数据

---

## 三、时间轴当前状态

`ClientView.vue` 和 `App.vue` 已恢复为原始版本（commit `3ad5de9`）：

- 全屏暗色 SVG 水平时间轴，无缩放平移
- 原始 `v-if`/`v-else` 双 `<router-view>` 布局
- 版本名水平居中显示
- 时间刻度每约 45 天一个
- 分支间距 64px，topMargin 50px
- 原生 `overflow:auto` 滚动

---

## 四、后端改进

| 改动 | 文件 |
|------|------|
| `CountVersions` 支持 branch_id/status 过滤 | `repo/version.go` |
| `ListVersions` handler 传入完整 VersionQuery | `handler/version.go` |
| Dashboard 适配新 CountVersions 签名 | `handler/dashboard.go` |

---

## 五、未来可考虑的需求

1. 时间轴整体重新设计（当前版本体验不佳）
2. 分支/版本的批量导入导出
3. 版本对比功能
4. 搜索/全文检索
5. 用户认证与权限
