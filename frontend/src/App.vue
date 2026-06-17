<template>
  <!-- 时间轴全屏模式 -->
  <div v-if="isTimeline" style="width:100vw;height:100vh;overflow:hidden">
    <router-view />
  </div>

  <!-- 管理端布局 -->
  <el-container v-else style="min-height: 100vh">
    <el-aside width="220px" style="background: #1d1e2c; color: #fff">
      <div style="padding: 20px 16px; font-size: 20px; font-weight: 700; letter-spacing: 1px; display: flex; align-items: center; gap: 8px">
        <el-icon :size="24"><Box /></el-icon>
        VerMan
      </div>
      <el-menu
        :default-active="currentRoute"
        :router="true"
        background-color="#1d1e2c"
        text-color="#a6a7b3"
        active-text-color="#409eff"
        style="border-right: none"
      >
        <el-menu-item index="/">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/branches">
          <el-icon><Share /></el-icon>
          <span>分支管理</span>
        </el-menu-item>
        <el-menu-item index="/versions">
          <el-icon><Box /></el-icon>
          <span>版本管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-main style="background: #f0f2f5; padding: 24px">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const currentRoute = computed(() => route.path)
const isTimeline = computed(() => route.path === '/timeline')
</script>
