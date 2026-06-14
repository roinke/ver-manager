<template>
  <div style="display:flex;flex-direction:column;height:100vh;background:#1a1b2e;color:#e0e0e0">
    <!-- 顶部标题栏 -->
    <div style="padding:12px 24px;background:#22243a;display:flex;justify-content:space-between;align-items:center;flex-shrink:0">
      <div style="display:flex;align-items:center;gap:12px">
        <h3 style="margin:0;color:#fff;font-size:18px">版本发布时间轴</h3>
        <span style="font-size:13px;color:#909399">{{ branches.length }} 个分支 · {{ totalVersions }} 个版本</span>
      </div>
      <span style="font-size:12px;color:#606080">VerMan v2.0</span>
    </div>

    <!-- SVG 图表区域（可滚动，撑满剩余空间） -->
    <div ref="scrollContainer" style="flex:1;overflow:auto;background:#1a1b2e" @scroll="onScroll">
      <svg
        :width="svgWidth" :height="svgHeight"
        style="display:block;min-width:800px"
      >
        <!-- 背景网格线 -->
        <line
          v-for="(tick, i) in timeTicks"
          :key="'g'+i"
          :x1="leftMargin + (tick.ratio * chartWidth)"
          :y1="topMargin"
          :x2="leftMargin + (tick.ratio * chartWidth)"
          :y2="svgHeight - 10"
          stroke="#2a2d45" stroke-width="1"
        />

        <!-- 时间轴 -->
        <line :x1="leftMargin" :y1="topMargin" :x2="leftMargin+chartWidth" :y2="topMargin" stroke="#4a4d65" stroke-width="1" />
        <text
          v-for="(tick, i) in timeTicks"
          :key="'t'+i"
          :x="leftMargin + (tick.ratio * chartWidth)"
          :y="topMargin - 8"
          text-anchor="middle" fill="#606380" font-size="12"
        >{{ tick.label }}</text>

        <!-- 分支线段 + 版本点 -->
        <template v-for="(b, bi) in layoutBranches" :key="b.id">
          <!-- 分支水平线 -->
          <line
            :x1="leftMargin + b.xStart * chartWidth"
            :y1="b.y"
            :x2="leftMargin + chartWidth + 20"
            :y2="b.y"
            :stroke="b.color" :stroke-width="2.5"
          />

          <!-- 分支拉取点（小圆） -->
          <circle
            v-if="b.xStart > 0"
            :cx="leftMargin + b.xStart * chartWidth"
            :cy="b.y" r="4"
            :fill="b.color" stroke="#fff" stroke-width="1.5"
          />

          <!-- 派生连线（父→子） -->
          <line
            v-if="b.parent_y"
            :x1="leftMargin + b.xStart * chartWidth"
            :y1="b.parent_y"
            :x2="leftMargin + b.xStart * chartWidth"
            :y2="b.y"
            :stroke="b.color" stroke-width="1.5" stroke-dasharray="4,3"
          />

          <!-- 版本点 -->
          <g v-for="v in b.versions" :key="v.id">
            <circle
              :cx="leftMargin + v.xRatio * chartWidth"
              :cy="b.y" r="6"
              :fill="statusColor(v.status)" stroke="#fff" stroke-width="2"
              style="cursor:pointer"
              @mouseenter="hoverVersion = v; hoverX = $event.offsetX; hoverY = b.y"
              @mouseleave="hoverVersion = null"
            />
            <text
              :x="leftMargin + v.xRatio * chartWidth"
              :y="b.y - 12"
              text-anchor="middle" fill="#a0a4c0" font-size="11"
            >{{ v.version_number }}</text>
          </g>

          <!-- 分支名标签 -->
          <text :x="10" :y="b.y + 4" fill="#c0c4e0" font-size="13" font-weight="600">{{ b.name }}</text>
        </template>

        <!-- 悬浮提示框 -->
        <g v-if="hoverVersion">
          <rect
            :x="hoverX + 12" :y="hoverY - 36"
            :width="hoverLabelW" height="44" rx="4"
            fill="#303133" opacity="0.9"
          />
          <text :x="hoverX + 20" :y="hoverY - 18" fill="#fff" font-size="12">
            <tspan x="0" dy="0">{{ hoverVersion.product_name }} {{ hoverVersion.version_number }}</tspan>
            <tspan :x="hoverX + 20" dy="16">{{ hoverVersion.build_time }}</tspan>
          </text>
        </g>
      </svg>
    </div>

    <!-- 图例 -->
    <div style="padding:10px 24px;border-top:1px solid #2a2d45;display:flex;gap:20px;font-size:12px;color:#606380;flex-shrink:0">
      <span>● 版本点</span>
      <span style="color:#67c23a">● released</span>
      <span style="color:#e6a23c">● draft</span>
      <span style="color:#909399">● deprecated</span>
      <span style="color:#f56c6c">● revoked</span>
      <span>┆ 分支拉取点</span>
      <span>- - 派生关系</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getAllBranches } from '../api/branch'
import { getVersions } from '../api/version'

const branches = ref([])
const allVersions = ref([])
const totalVersions = ref(0)
const hoverVersion = ref(null)
const hoverX = ref(0)
const hoverY = ref(0)
const hoverLabelW = ref(220)
const containerWidth = ref(1200)

const COLORS = ['#409eff','#67c23a','#e6a23c','#f56c6c','#909399','#8e71c7','#3cb4c4','#e085b4']

const leftMargin = 140
const topMargin = 50
const chartWidth = computed(() => Math.max(containerWidth.value - leftMargin - 40, 600))
const svgWidth = computed(() => leftMargin + chartWidth.value + 40)
const svgHeight = computed(() => topMargin + branches.value.length * 64 + 20)

// 计算时间范围
function calcTimeRange() {
  let minT = Infinity, maxT = -Infinity
  const now = Date.now()
  branches.value.forEach(b => {
    const t = b.pulled_at ? new Date(b.pulled_at).getTime() : new Date(b.created_at).getTime()
    if (t < minT) minT = t
    if (t > maxT) maxT = t
  })
  // 给 10% padding
  const span = (maxT - minT) || 86400000 * 30 // 最少 30 天
  return { min: minT - span * 0.05, max: maxT + span * 0.15, span: span * 1.2 }
}

function xRatio(ts) {
  if (!ts) return 0
  const t = new Date(ts).getTime()
  const range = calcTimeRange()
  return Math.max(0, Math.min(1, (t - range.min) / range.span))
}

// 分支 Y 轴布局（确保父在子上方）
function layoutY() {
  const range = calcTimeRange()
  const placed = new Set()
  const rows = []

  function place(branch, depth) {
    if (placed.has(branch.id)) return
    placed.add(branch.id)
    const y = topMargin + 10 + rows.length * 64

    const start = branch.pulled_at || branch.created_at
    const xS = start ? Math.max(0, (new Date(start).getTime() - range.min) / range.span) : 0

    const versions = []
    if (branch._versions) {
      branch._versions.forEach(v => {
        const xR = v.build_time ? Math.max(0, (new Date(v.build_time).getTime() - range.min) / range.span) : xS
        if (xR >= 0 && xR <= 1.05) {
          versions.push({ ...v, xRatio: xR })
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

  // 先放根分支，再放子分支
  const roots = branches.value.filter(b => !b.parent_branch_id)
  roots.forEach(r => place(r, 0))

  // 递归放置子分支
  let changed = true
  while (changed) {
    changed = false
    branches.value.forEach(b => {
      if (!placed.has(b.id) && b.parent_branch_id && placed.has(b.parent_branch_id)) {
        place(b, 0)
        changed = true
      }
    })
  }
  // 剩余未放置的
  branches.value.forEach(b => { if (!placed.has(b.id)) place(b, 0) })

  return rows
}

const layoutBranches = computed(() => layoutY())

// 时间刻度
const timeTicks = computed(() => {
  const range = calcTimeRange()
  const span = range.span
  const count = Math.max(4, Math.floor(span / (86400000 * 45))) // ~每45天一个刻度
  const ticks = []
  for (let i = 0; i <= count; i++) {
    const t = range.min + (span * i / count)
    const d = new Date(t)
    const label = d.getFullYear() + '-' + String(d.getMonth()+1).padStart(2,'0')
    ticks.push({ ratio: i / count, label })
  }
  return ticks
})

function statusColor(s) {
  return { released:'#67c23a', draft:'#e6a23c', deprecated:'#909399', revoked:'#f56c6c' }[s] || '#909399'
}

function onScroll() {}

onMounted(async () => {
  try {
    const [bRes, vRes] = await Promise.all([
      getAllBranches(),
      getVersions({ limit: 9999 }),
    ])
    branches.value = bRes.data || []
    allVersions.value = vRes.data || []
    totalVersions.value = allVersions.value.length

    // 给每个分支附加其版本
    branches.value.forEach(b => {
      b._versions = allVersions.value.filter(v => v.branch_id === b.id)
    })
    containerWidth.value = window.innerWidth - 260
  } catch (e) { console.error(e) }
})
</script>
