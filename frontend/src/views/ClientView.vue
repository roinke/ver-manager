<template>
  <div style="display:flex;flex-direction:column;height:100vh;background:#1a1b2e;color:#e0e0e0;overflow:hidden">
    <!-- 顶部标题栏 -->
    <div style="padding:10px 24px;background:#22243a;display:flex;justify-content:space-between;align-items:center;flex-shrink:0">
      <div style="display:flex;align-items:center;gap:12px">
        <h3 style="margin:0;color:#fff;font-size:18px">版本发布时间轴</h3>
        <span style="font-size:13px;color:#909399">{{ branches.length }} 个分支 · {{ totalVersions }} 个版本</span>
      </div>
      <div style="display:flex;align-items:center;gap:12px">
        <span style="font-size:12px;color:#909399">{{ Math.round(zoom * 100) }}%</span>
        <button
          @click="resetView"
          style="background:#409eff;color:#fff;border:none;padding:4px 12px;border-radius:4px;cursor:pointer;font-size:12px"
        >重置视图</button>
      </div>
    </div>

    <!-- 主区域：分支名固定面板 + SVG 滚动区 -->
    <div ref="mainArea" style="flex:1;overflow-y:auto;overflow-x:hidden;background:#1a1b2e">
      <div :style="{ display:'flex', height: svgHeight+'px', minHeight: '100%' }">
        <!-- 左侧固定分支名面板 -->
        <div
          ref="branchPanel"
          style="width:150px;flex-shrink:0;position:relative;background:#1a1b2e;z-index:1;border-right:1px solid #2a2d45"
        >
          <div
            v-for="b in layoutBranches" :key="b.id"
            :style="{
              position:'absolute', left:'12px', top: b.y+'px', transform:'translateY(-8px)',
              color: b.color, fontSize:'13px', fontWeight:'600',
              whiteSpace:'nowrap', overflow:'hidden', textOverflow:'ellipsis', maxWidth:'130px'
            }"
          >{{ b.name }}</div>
        </div>

        <!-- SVG 时间轴（仅水平滚动） -->
        <div
          ref="svgArea"
          style="flex:1;overflow-x:auto;overflow-y:hidden"
          @wheel.prevent="onWheel"
          @pointerdown="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointerleave="onPointerUp"
        >
          <svg :width="svgWidth" :height="svgHeight" style="display:block">
            <!-- 背景纵向网格线 -->
            <line
              v-for="(tick, i) in timeTicks" :key="'g'+i"
              :x1="xPos(tick.ratio)" :y1="topMargin"
              :x2="xPos(tick.ratio)" :y2="svgHeight - 10"
              stroke="#2a2d45" stroke-width="1"
            />

            <!-- 时间轴横线 -->
            <line
              :x1="leftMargin" :y1="topMargin"
              :x2="leftMargin + chartWidth" :y2="topMargin"
              stroke="#4a4d65" stroke-width="1"
            />

            <!-- 时间刻度标签 -->
            <text
              v-for="(tick, i) in timeTicks" :key="'t'+i"
              :x="xPos(tick.ratio)" :y="topMargin - 8"
              text-anchor="middle" fill="#606380" font-size="12"
            >{{ tick.label }}</text>

            <!-- 分支线 + 版本点 -->
            <template v-for="b in layoutBranches" :key="b.id">
              <!-- 分支水平线 -->
              <line
                :x1="xPos(b.xStart)" :y1="b.y"
                :x2="xPos(1) + 60" :y2="b.y"
                :stroke="b.color" stroke-width="2.5"
              />

              <!-- 分支拉取点（小圆） -->
              <circle
                v-if="b.xStart > 0"
                :cx="xPos(b.xStart)" :cy="b.y" r="4"
                :fill="b.color" stroke="#fff" stroke-width="1.5"
              />

              <!-- 派生连线（父→子竖虚线） -->
              <line
                v-if="b.parent_y != null"
                :x1="xPos(b.xStart)" :y1="b.parent_y"
                :x2="xPos(b.xStart)" :y2="b.y"
                :stroke="b.color" stroke-width="1.5" stroke-dasharray="4,3"
              />

              <!-- 版本点 -->
              <g v-for="v in b.versions" :key="v.id">
                <circle
                  :cx="xPos(v.xRatio)" :cy="b.y" r="6"
                  :fill="statusColor(v.status)" stroke="#fff" stroke-width="2"
                  style="cursor:pointer"
                  @mouseenter="showTooltip(v, b)"
                  @mouseleave="hideTooltip"
                />
                <text
                  :x="xPos(v.xRatio) + 7" :y="b.y - 2"
                  text-anchor="start" fill="#a0a4c0" font-size="11"
                  :transform="`rotate(-35, ${xPos(v.xRatio)}, ${b.y})`"
                >{{ v.version_number }}</text>
              </g>
            </template>

            <!-- 悬浮提示框 -->
            <g v-if="hoverVersion" style="pointer-events:none">
              <rect
                :x="hoverX + 12" :y="hoverY - 46"
                :width="hoverLabelW" height="56" rx="4"
                fill="#303133" opacity="0.93"
              />
              <text :x="hoverX + 20" :y="hoverY - 26" fill="#fff" font-size="12">
                <tspan :x="hoverX + 20" dy="0">{{ hoverVersion.product_name }} {{ hoverVersion.version_number }}</tspan>
                <tspan :x="hoverX + 20" dy="16" fill="#a0a4c0" font-size="11">{{ hoverVersion.build_time }}</tspan>
                <tspan :x="hoverX + 20" dy="14" fill="#a0a4c0" font-size="11">{{ hoverVersion.description || '暂无描述' }}</tspan>
              </text>
            </g>
          </svg>
        </div>
      </div>
    </div>

    <!-- 图例栏 -->
    <div style="padding:8px 24px;border-top:1px solid #2a2d45;display:flex;gap:20px;font-size:12px;color:#606380;flex-shrink:0;background:#1a1b2e">
      <span>● 版本点</span>
      <span style="color:#67c23a">● released</span>
      <span style="color:#e6a23c">● draft</span>
      <span style="color:#909399">● deprecated</span>
      <span style="color:#f56c6c">● revoked</span>
      <span>┆ 拉取点</span>
      <span>- - 派生关系</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { getAllBranches } from '../api/branch'
import { getVersions } from '../api/version'

// ── 响应式状态 ──
const branches = ref([])
const allVersions = ref([])
const totalVersions = ref(0)
const zoom = ref(1)
const containerWidth = ref(1200)
const containerHeight = ref(800)
const hoverVersion = ref(null)
const hoverX = ref(0)
const hoverY = ref(0)
const hoverLabelW = ref(240)

const mainArea = ref(null)
const branchPanel = ref(null)
const svgArea = ref(null)

// ── 常量 ──
const COLORS = ['#409eff','#67c23a','#e6a23c','#f56c6c','#909399','#8e71c7','#3cb4c4','#e085b4']
const leftMargin = 20
const topMargin = 50
const YEAR_MS = 365 * 86400000
const PADDING_MS = 30 * 86400000
const ZOOM_MIN = 0.3
const ZOOM_MAX = 8
const ZOOM_STEP = 1.15

// ── 时间范围计算 ──
const timeRange = computed(() => {
  let minT = Infinity, maxT = -Infinity

  allVersions.value.forEach(v => {
    if (v.build_time) {
      const t = new Date(v.build_time).getTime()
      if (t < minT) minT = t
      if (t > maxT) maxT = t
    }
  })
  branches.value.forEach(b => {
    const t = b.pulled_at ? new Date(b.pulled_at).getTime() : new Date(b.created_at).getTime()
    if (t < minT) minT = t
    if (t > maxT) maxT = t
  })

  const now = Date.now()
  if (minT === Infinity) {
    minT = now - YEAR_MS
    maxT = now
  }
  const span = maxT - minT || YEAR_MS
  const fullMin = minT - span * 0.1
  const fullMax = maxT + span * 0.1
  const fullSpan = fullMax - fullMin

  return {
    fullMin, fullMax, fullSpan,
    min: fullMin,
    max: fullMax,
    span: fullSpan,
    dataMin: minT,
    dataMax: maxT,
    latestTime: maxT,
    defaultViewMin: maxT - YEAR_MS,
    defaultViewMax: maxT + PADDING_MS,
  }
})

// ── 像素/缩放计算 ──
const yearWidth = computed(() => Math.max(600, containerWidth.value - 150))
const pixelsPerMs = computed(() => yearWidth.value / YEAR_MS * zoom.value)
const chartWidth = computed(() => pixelsPerMs.value * timeRange.value.fullSpan)
const svgWidth = computed(() => leftMargin + chartWidth.value + 40)

// ── 分支间距（铺满屏幕） ──
const branchSpacing = computed(() => {
  const n = branches.value.length || 1
  const available = containerHeight.value - topMargin - 20
  const ideal = available / n
  return Math.min(120, Math.max(64, ideal))
})

const svgHeight = computed(() => {
  const h = topMargin + layoutBranches.value.length * branchSpacing.value + 20
  return Math.max(containerHeight.value, h)
})

// ── 时间刻痕 ──
const timeTicks = computed(() => {
  const range = timeRange.value
  const targetInterval = range.fullSpan / Math.max(4, Math.floor(chartWidth.value / 120))

  const niceIntervals = [
    3600000, 86400000, 7 * 86400000, 30 * 86400000,
    90 * 86400000, 180 * 86400000, 365 * 86400000,
  ]
  let interval = niceIntervals[niceIntervals.length - 1]
  for (const ni of niceIntervals) {
    if (ni >= targetInterval * 0.7) { interval = ni; break }
  }

  const ticks = []
  const start = Math.ceil(range.fullMin / interval) * interval
  for (let t = start; t <= range.fullMax; t += interval) {
    const d = new Date(t)
    const label = interval >= 30 * 86400000
      ? d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0')
      : String(d.getMonth() + 1) + '/' + String(d.getDate())
    const ratio = (t - range.fullMin) / range.fullSpan
    ticks.push({ ratio, label })
  }
  return ticks
})

// ── 分支 Y 轴布局 ──
const layoutBranches = computed(() => {
  const range = timeRange.value
  const placed = new Set()
  const rows = []

  function place(branch) {
    if (placed.has(branch.id)) return
    placed.add(branch.id)

    const y = topMargin + 10 + rows.length * branchSpacing.value
    const start = branch.pulled_at || branch.created_at
    const xS = start ? (new Date(start).getTime() - range.fullMin) / range.fullSpan : 0

    const versions = []
    if (branch._versions) {
      branch._versions.forEach(v => {
        if (v.build_time) {
          const xR = (new Date(v.build_time).getTime() - range.fullMin) / range.fullSpan
          if (xR >= -0.05 && xR <= 1.05) {
            versions.push({ ...v, xRatio: xR })
          }
        }
      })
    }

    const parentRow = rows.find(r => r.id === branch.parent_branch_id)
    rows.push({
      id: branch.id, name: branch.name, y,
      xStart: xS,
      color: COLORS[rows.length % COLORS.length],
      versions,
      parent_y: parentRow ? parentRow.y : null,
    })
  }

  // 根分支优先
  branches.value.filter(b => !b.parent_branch_id).forEach(r => place(r))
  // 子分支递归
  let changed = true
  while (changed) {
    changed = false
    branches.value.forEach(b => {
      if (!placed.has(b.id) && b.parent_branch_id && placed.has(b.parent_branch_id)) {
        place(b); changed = true
      }
    })
  }
  branches.value.forEach(b => { if (!placed.has(b.id)) place(b) })

  return rows
})

// ── 工具函数 ──
function xPos(ratio) {
  return leftMargin + ratio * chartWidth.value
}

function statusColor(s) {
  return { released: '#67c23a', draft: '#e6a23c', deprecated: '#909399', revoked: '#f56c6c' }[s] || '#909399'
}

// ── 悬浮提示 ──
let hideTimer = null
function showTooltip(v, b) {
  if (hideTimer) clearTimeout(hideTimer)
  hoverVersion.value = v
  hoverX.value = xPos(v.xRatio)
  hoverY.value = b.y
}
function hideTooltip() {
  hideTimer = setTimeout(() => { hoverVersion.value = null }, 120)
}

// ── 缩放 ──
function onWheel(e) {
  const area = svgArea.value
  if (!area) return

  const rect = area.getBoundingClientRect()
  const mouseX = e.clientX - rect.left + area.scrollLeft
  const oldWidth = chartWidth.value

  if (e.deltaY < 0) {
    zoom.value = Math.min(ZOOM_MAX, zoom.value * ZOOM_STEP)
  } else {
    zoom.value = Math.max(ZOOM_MIN, zoom.value / ZOOM_STEP)
  }

  nextTick(() => {
    const ratio = oldWidth > 0 ? mouseX / oldWidth : 0
    area.scrollLeft = Math.max(0, ratio * chartWidth.value - (e.clientX - rect.left))
  })
}

// ── 拖拽平移 ──
let isDragging = false
let dragStartX = 0
let dragStartScroll = 0

function onPointerDown(e) {
  if (e.button !== 0) return
  isDragging = true
  dragStartX = e.clientX
  dragStartScroll = svgArea.value.scrollLeft
  svgArea.value.style.cursor = 'grabbing'
  svgArea.value.style.userSelect = 'none'
}

function onPointerMove(e) {
  if (!isDragging) return
  svgArea.value.scrollLeft = dragStartScroll + dragStartX - e.clientX
}

function onPointerUp() {
  if (!isDragging) return
  isDragging = false
  if (svgArea.value) {
    svgArea.value.style.cursor = ''
    svgArea.value.style.userSelect = ''
  }
}

// ── 重置视图 ──
function resetView() {
  zoom.value = 1
  nextTick(() => scrollToDefaultView())
}

function scrollToDefaultView() {
  const area = svgArea.value
  if (!area) return
  const range = timeRange.value
  const maxRatio = (range.defaultViewMax - range.fullMin) / range.fullSpan
  const rightEdgeX = xPos(maxRatio)
  const viewportW = area.clientWidth
  area.scrollLeft = Math.max(0, rightEdgeX - viewportW + 40)
}

// ── 生命周期 ──
let resizeObserver = null

onMounted(async () => {
  try {
    const [bRes, vRes] = await Promise.all([
      getAllBranches(),
      getVersions({ limit: 9999 }),
    ])
    branches.value = bRes.data || []
    allVersions.value = vRes.data || []
    totalVersions.value = allVersions.value.length

    branches.value.forEach(b => {
      b._versions = allVersions.value.filter(v => v.branch_id === b.id)
    })
  } catch (e) {
    console.error(e)
  }

  await nextTick()

  // 初始容器尺寸
  if (mainArea.value) {
    containerWidth.value = mainArea.value.clientWidth
    containerHeight.value = mainArea.value.clientHeight
  }
  await nextTick()

  // 默认显示最近一年
  scrollToDefaultView()

  // 监听容器尺寸变化
  if (mainArea.value) {
    resizeObserver = new ResizeObserver(entries => {
      for (const entry of entries) {
        containerWidth.value = entry.contentRect.width
        containerHeight.value = entry.contentRect.height
      }
    })
    resizeObserver.observe(mainArea.value)
  }
})

onBeforeUnmount(() => {
  if (resizeObserver) resizeObserver.disconnect()
})
</script>
