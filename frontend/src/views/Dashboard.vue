<template>
  <div>
    <h2 style="margin-bottom: 24px">📊 仪表盘</h2>

    <!-- 统计卡片 -->
    <el-row :gutter="20" style="margin-bottom: 24px">
      <el-col :span="8">
        <el-card shadow="hover">
          <div style="text-align: center">
            <div style="font-size: 36px; font-weight: 700; color: #409eff">{{ stats.branch_count }}</div>
            <div style="color: #909399; margin-top: 4px">活跃分支</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div style="text-align: center">
            <div style="font-size: 36px; font-weight: 700; color: #67c23a">{{ stats.version_count }}</div>
            <div style="color: #909399; margin-top: 4px">版本总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div style="text-align: center">
            <div style="font-size: 36px; font-weight: 700; color: #e6a23c">{{ stats.total_branches }}</div>
            <div style="color: #909399; margin-top: 4px">分支总数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <!-- 分支树 -->
      <el-col :span="12">
        <el-card header="🌿 分支树" shadow="hover">
          <div v-if="stats.branches && stats.branches.length" style="font-family: monospace; line-height: 2; font-size: 14px">
            <BranchTreeNode
              v-for="(b, i) in rootBranches"
              :key="b.id"
              :branch="b"
              :all-branches="stats.branches"
              :is-last="i === rootBranches.length - 1"
              prefix=""
            />
          </div>
          <el-empty v-else description="暂无分支" />
        </el-card>
      </el-col>

      <!-- 最近版本 -->
      <el-col :span="12">
        <el-card header="🕐 最近版本" shadow="hover">
          <el-table v-if="stats.recent_versions && stats.recent_versions.length" :data="stats.recent_versions" size="small" stripe>
            <el-table-column prop="product_name" label="产品" width="120" />
            <el-table-column prop="version_number" label="版本号" width="130">
              <template #default="{ row }">
                <router-link :to="`/versions/${row.id}`" style="color: #409eff; text-decoration: none">
                  {{ row.version_number }}
                </router-link>
              </template>
            </el-table-column>
            <el-table-column prop="branch_name" label="分支" width="100">
              <template #default="{ row }">
                <el-tag size="small" type="primary" effect="plain">{{ row.branch_name }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="时间" width="120">
              <template #default="{ row }">
                {{ formatTime(row.build_time) }}
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="暂无版本记录" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getDashboard } from '../api/version'
import BranchTreeNode from '../components/BranchTreeNode.vue'

const stats = ref({})

const rootBranches = computed(() => {
  if (!stats.value.branches) return []
  return stats.value.branches.filter(b => !b.parent_branch_id)
})

function formatTime(t) {
  if (!t) return '-'
  return t.substring(0, 16).replace('T', ' ')
}

onMounted(async () => {
  try {
    const res = await getDashboard()
    stats.value = res
  } catch (e) {
    console.error(e)
  }
})
</script>
