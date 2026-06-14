<template>
  <div>
    <div :style="{ marginLeft: prefix ? '0' : '0' }">
      {{ prefix }}{{ isLast ? '└── ' : '├── ' }}
      <router-link :to="`/branches/${branch.id}`" style="color: #409eff; text-decoration: none; font-weight: 500">
        {{ branch.name }}
      </router-link>
      <el-tag size="small" style="margin-left: 6px" :type="typeTag(branch.branch_type)" effect="plain">
        {{ branch.branch_type }}
      </el-tag>
      <span v-if="!branch.is_active" style="color: #909399; font-size: 12px">[停用]</span>
    </div>
    <BranchTreeNode
      v-for="(child, i) in children"
      :key="child.id"
      :branch="child"
      :all-branches="allBranches"
      :is-last="i === children.length - 1"
      :prefix="prefix + (isLast ? '    ' : '│   ')"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  branch: Object,
  allBranches: Array,
  isLast: Boolean,
  prefix: { type: String, default: '' },
})

const children = computed(() => {
  return props.allBranches.filter(b => b.parent_branch_id === props.branch.id)
})

function typeTag(type) {
  const map = { main: 'primary', release: 'success', feature: '', hotfix: 'danger' }
  return map[type] || 'info'
}
</script>
