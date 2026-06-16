<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:20px">
      <h2 style="margin:0">🌿 分支管理</h2>
      <el-button type="primary" @click="openCreate"><el-icon><Plus /></el-icon>新建分支</el-button>
    </div>

    <!-- 分支列表 -->
    <el-card shadow="never">
      <el-table :data="branches" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="分支名称" width="160">
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="父分支" width="120">
          <template #default="{ row }">{{ getParentName(row.parent_branch_id) }}</template>
        </el-table-column>
        <el-table-column prop="branch_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTag(row.branch_type)">{{ row.branch_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="拉取时间" width="160">
          <template #default="{ row }">{{ row.pulled_at || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.is_active ? 'success' : 'info'">
              {{ row.is_active ? '活跃' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="openDetail(row)">详情</el-button>
            <el-button size="small" text type="primary" @click="openEdit(row)">编辑</el-button>
            <el-popconfirm
              :title="row.is_active ? '确定停用该分支？' : '确定重新启用？'"
              @confirm="toggleBranch(row)"
            >
              <template #reference>
                <el-button size="small" text :type="row.is_active ? 'warning' : 'success'">
                  {{ row.is_active ? '停用' : '启用' }}
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div style="display:flex;justify-content:flex-end;margin-top:16px">
        <el-pagination
          v-model:current-page="page" v-model:page-size="pageSize"
          :total="total" :page-sizes="[10,20,50]" layout="total,sizes,prev,pager,next"
          @size-change="load" @current-change="load"
        />
      </div>
    </el-card>

    <!-- ===== 新建 / 编辑 弹窗 ===== -->
    <el-dialog
      v-model="formVisible" :title="formTitle" width="560px"
      :close-on-click-modal="false" @closed="resetForm"
    >
      <el-form :model="form" label-width="90px">
        <el-form-item label="分支名称" required>
          <el-input v-model="form.name" placeholder="如 master / feature-login" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="父分支">
              <el-select v-model="form.parent_branch_id" clearable placeholder="无 (根分支)" style="width:100%">
                <el-option v-for="b in parentOptions" :key="b.id" :label="b.name" :value="b.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="类型">
              <el-select v-model="form.branch_type" style="width:100%">
                <el-option label="main" value="main" />
                <el-option label="release" value="release" />
                <el-option label="feature" value="feature" />
                <el-option label="hotfix" value="hotfix" />
                <el-option label="custom" value="custom" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="基于版本">
          <el-select v-model="selectedVersionId" clearable
            placeholder="不指定（手动填写拉取时间）" style="width:100%"
            :disabled="!form.parent_branch_id" :loading="versionsLoading"
            @change="onVersionSelect"
          >
            <el-option v-for="v in parentVersions" :key="v.id"
              :label="`${v.version_number} — ${v.product_name} (${v.build_time})`"
              :value="v.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="拉取时间">
          <el-date-picker v-model="form.pulled_at" type="datetime" placeholder="选择实际拉分支时间"
            format="YYYY-MM-DD HH:mm:ss" value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" placeholder="分支用途说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">
          {{ editingId ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- ===== 详情弹窗 ===== -->
    <el-dialog v-model="detailVisible" title="分支详情" width="680px">
      <template v-if="detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="名称"><strong>{{ detail.name }}</strong></el-descriptions-item>
          <el-descriptions-item label="父分支">{{ getParentName(detail.parent_branch_id) }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag size="small" :type="typeTag(detail.branch_type)">{{ detail.branch_type }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag size="small" :type="detail.is_active ? 'success' : 'info'">
              {{ detail.is_active ? '活跃' : '已停用' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="拉取时间">{{ detail.pulled_at || '(未填写)' }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ detail.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detail.created_at }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ detail.updated_at }}</el-descriptions-item>
        </el-descriptions>

        <!-- 关联版本 -->
        <el-divider />
        <h4 style="margin-bottom:12px">📦 该分支的版本 ({{ detailTotal }})</h4>
        <template v-if="detailTotal > 0">
          <el-table :data="detailVersions" size="small" stripe v-loading="detailLoading">
            <el-table-column prop="id" label="ID" width="50" />
            <el-table-column prop="product_name" label="产品" width="120" />
            <el-table-column prop="version_number" label="版本号" min-width="120" />
            <el-table-column prop="build_time" label="构建时间" width="160" />
            <el-table-column prop="status" label="状态" width="90">
              <template #default="{ row }">
                <el-tag size="small" :type="statusTag(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div style="display:flex;justify-content:flex-end;margin-top:12px">
            <el-pagination
              v-model:current-page="detailPage" v-model:page-size="detailPageSize"
              :total="detailTotal" :page-sizes="[5,10,20]"
              layout="total,sizes,prev,pager,next" small
              @size-change="loadDetailVersions" @current-change="loadDetailVersions"
            />
          </div>
        </template>
        <el-empty v-else description="暂无版本" :image-size="60" />
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getBranches, createBranch, updateBranch } from '../api/branch'
import { getVersions } from '../api/version'

const branches = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// ---- 新建 / 编辑 ----
const formVisible = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const form = reactive({
  name: '', parent_branch_id: null, branch_type: 'custom',
  description: '', pulled_at: '',
})

// 父分支版本列表
const parentVersions = ref([])
const versionsLoading = ref(false)
const selectedVersionId = ref(null)

const formTitle = computed(() => editingId.value ? '编辑分支' : '新建分支')
const isEdit = computed(() => !!editingId.value)

// 排除自身的可选父分支
const parentOptions = computed(() => {
  if (!editingId.value) return branches.value
  return branches.value.filter(b => b.id !== editingId.value)
})

// 加载某个父分支下的全部版本
async function loadParentVersions(branchId) {
  versionsLoading.value = true
  try {
    const res = await getVersions({ branch_id: branchId, page_size: 9999 })
    parentVersions.value = res.data || []
  } catch { parentVersions.value = [] }
  finally { versionsLoading.value = false }
}

// 选中版本时自动填充拉取时间
function onVersionSelect(versionId) {
  if (!versionId) return
  const v = parentVersions.value.find(item => item.id === versionId)
  if (v) {
    form.pulled_at = v.build_time
  }
}

// 监听父分支变化 → 加载版本列表
watch(() => form.parent_branch_id, (newVal) => {
  selectedVersionId.value = null
  if (newVal) {
    loadParentVersions(newVal)
  } else {
    parentVersions.value = []
  }
})

function openCreate() {
  editingId.value = null
  formVisible.value = true
}

function openEdit(row) {
  editingId.value = row.id
  form.name = row.name
  form.parent_branch_id = row.parent_branch_id
  form.branch_type = row.branch_type
  form.description = row.description
  form.pulled_at = row.pulled_at || ''
  selectedVersionId.value = null
  formVisible.value = true
  if (row.parent_branch_id) {
    loadParentVersions(row.parent_branch_id)
  } else {
    parentVersions.value = []
  }
}

function resetForm() {
  editingId.value = null
  parentVersions.value = []
  selectedVersionId.value = null
  Object.assign(form, {
    name: '', parent_branch_id: null, branch_type: 'custom',
    description: '', pulled_at: '',
  })
}

async function submitForm() {
  if (!form.name.trim()) { ElMessage.warning('请输入分支名称'); return }
  submitting.value = true
  try {
    if (editingId.value) {
      await updateBranch(editingId.value, { ...form })
      ElMessage.success('保存成功')
    } else {
      await createBranch({ ...form })
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    await load()
  } catch (e) { ElMessage.error(e.message) }
  finally { submitting.value = false }
}

async function toggleBranch(row) {
  try {
    await updateBranch(row.id, { is_active: !row.is_active })
    ElMessage.success(row.is_active ? '已停用' : '已启用')
    load()
  } catch (e) { ElMessage.error(e.message) }
}

// ---- 详情 ----
const detailVisible = ref(false)
const detail = ref(null)
const detailVersions = ref([])
const detailLoading = ref(false)
const detailPage = ref(1)
const detailPageSize = ref(5)
const detailTotal = ref(0)

async function openDetail(row) {
  detail.value = row
  detailVisible.value = true
  detailPage.value = 1
  loadDetailVersions()
}

async function loadDetailVersions() {
  if (!detail.value) return
  detailLoading.value = true
  try {
    const res = await getVersions({
      branch_id: detail.value.id,
      page: detailPage.value,
      page_size: detailPageSize.value,
    })
    detailVersions.value = res.data || []
    detailTotal.value = res.total || 0
  } catch { detailVersions.value = [] }
  finally { detailLoading.value = false }
}

// ---- 工具 ----
function getParentName(pid) {
  if (!pid) return '-'
  const p = branches.value.find(b => b.id === pid)
  return p ? p.name : '-'
}
function typeTag(t) {
  return { main: '', release: 'success', feature: 'primary', hotfix: 'danger' }[t] || 'info'
}
function statusTag(s) {
  return { released: 'success', draft: 'warning', deprecated: 'info', revoked: 'danger' }[s] || 'info'
}

async function load() {
  loading.value = true
  try {
    const res = await getBranches(page.value, pageSize.value)
    branches.value = res.data || []
    total.value = res.total || 0
  } finally { loading.value = false }
}

onMounted(load)
</script>
