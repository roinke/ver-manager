<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:20px">
      <h2 style="margin:0">📦 版本管理</h2>
      <el-button type="primary" @click="openCreate"><el-icon><Plus /></el-icon>新建版本</el-button>
    </div>

    <!-- 筛选 -->
    <el-card shadow="never" style="margin-bottom:16px">
      <el-form :inline="true" :model="filters">
        <el-form-item label="分支">
          <el-select v-model="filters.branch_id" clearable placeholder="全部" @change="reload">
            <el-option v-for="b in branches" :key="b.id" :label="b.name" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="产品">
          <el-input v-model="filters.product" clearable placeholder="产品名称" @change="reload" style="width:180px" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filters.status" clearable placeholder="全部" @change="reload" style="width:120px">
            <el-option label="released" value="released" />
            <el-option label="draft" value="draft" />
            <el-option label="deprecated" value="deprecated" />
            <el-option label="revoked" value="revoked" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button @click="filters = {}; reload()">清除筛选</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 版本列表 -->
    <el-card shadow="never">
      <el-table :data="versions" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="product_name" label="产品" width="140" />
        <el-table-column prop="version_number" label="版本号" width="180">
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.version_number }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="branch_name" label="分支" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="primary">{{ row.branch_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="构建时间" width="170">
          <template #default="{ row }">{{ row.build_time }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="statusTag(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text type="primary" @click="openDetail(row)">详情</el-button>
            <el-button size="small" text type="primary" @click="openEdit(row)">编辑</el-button>
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
      v-model="formVisible" :title="formTitle" width="620px"
      :close-on-click-modal="false" @closed="resetForm"
    >
      <el-form :model="form" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="分支" required>
              <el-select v-model="form.branch_id" placeholder="选择分支" style="width:100%">
                <el-option v-for="b in branches" :key="b.id" :label="b.name" :value="b.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width:100%">
                <el-option label="released" value="released" />
                <el-option label="draft" value="draft" />
                <el-option label="deprecated" value="deprecated" />
                <el-option label="revoked" value="revoked" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="产品名称" required>
              <el-input v-model="form.product_name" placeholder="如 智能网关A型" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="版本号" required>
              <el-input v-model="form.version_number" placeholder="任意格式，如 V2.0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="构建时间" required>
              <el-date-picker v-model="form.build_time" type="datetime"
                placeholder="选择构建时间" format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="提交哈希">
              <el-input v-model="form.commit_hash" placeholder="Git SHA" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="构建产物">
              <el-input v-model="form.artifact_url" placeholder="URL 或路径" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="描述">
          <el-input v-model="form.description" placeholder="版本简要描述" />
        </el-form-item>
        <el-form-item label="发布说明">
          <el-input v-model="form.release_notes" type="textarea" :rows="2" placeholder="详细 changelog" />
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
    <el-dialog v-model="detailVisible" title="版本详情" width="620px">
      <template v-if="detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="产品"><strong>{{ detail.product_name }}</strong></el-descriptions-item>
          <el-descriptions-item label="版本号">
            <el-tag>{{ detail.version_number }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="分支">
            <el-tag type="primary">{{ detail.branch_name }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTag(detail.status)">{{ detail.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="构建时间">{{ detail.build_time }}</el-descriptions-item>
          <el-descriptions-item label="Git 提交">
            <code v-if="detail.commit_hash">{{ detail.commit_hash }}</code>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="构建产物">
            <el-link v-if="detail.artifact_url" :href="detail.artifact_url" target="_blank" type="primary">
              {{ detail.artifact_url }}
            </el-link>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ detail.description || '-' }}</el-descriptions-item>
          <el-descriptions-item label="发布说明" :span="2">
            <pre v-if="detail.release_notes" style="margin:0;white-space:pre-wrap;font-size:13px">{{ detail.release_notes }}</pre>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detail.created_at }}</el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getVersions, createVersion, updateVersion } from '../api/version'
import { getAllBranches } from '../api/branch'

const branches = ref([])
const versions = ref([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filters = reactive({ branch_id: null, product: '', status: '' })

// ---- 新建 / 编辑 ----
const formVisible = ref(false)
const editingId = ref(null)
const submitting = ref(false)
const form = reactive({
  branch_id: null, product_name: '', version_number: '', status: 'released',
  build_time: '', description: '', release_notes: '', commit_hash: '', artifact_url: '',
})

const formTitle = computed(() => editingId.value ? '编辑版本' : '新建版本')

function openCreate() {
  editingId.value = null
  formVisible.value = true
}

function openEdit(row) {
  editingId.value = row.id
  form.branch_id = row.branch_id
  form.product_name = row.product_name
  form.version_number = row.version_number
  form.status = row.status
  form.build_time = row.build_time || ''
  form.description = row.description
  form.release_notes = row.release_notes
  form.commit_hash = row.commit_hash
  form.artifact_url = row.artifact_url
  formVisible.value = true
}

function resetForm() {
  editingId.value = null
  Object.assign(form, {
    branch_id: null, product_name: '', version_number: '', status: 'released',
    build_time: '', description: '', release_notes: '', commit_hash: '', artifact_url: '',
  })
}

async function submitForm() {
  if (!form.branch_id || !form.product_name.trim() || !form.version_number.trim() || !form.build_time) {
    ElMessage.warning('请填写分支、产品名称、版本号、构建时间'); return
  }
  submitting.value = true
  try {
    if (editingId.value) {
      await updateVersion(editingId.value, { ...form })
      ElMessage.success('保存成功')
    } else {
      await createVersion({ ...form })
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    await load()
  } catch (e) { ElMessage.error(e.message) }
  finally { submitting.value = false }
}

// ---- 详情 ----
const detailVisible = ref(false)
const detail = ref(null)

async function openDetail(row) {
  detail.value = row
  detailVisible.value = true
}

// ---- 工具 ----
function statusTag(s) {
  return { released: 'success', draft: 'warning', deprecated: 'info', revoked: 'danger' }[s] || 'info'
}

function reload() { page.value = 1; load() }
async function load() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (filters.branch_id) params.branch_id = filters.branch_id
    if (filters.product) params.product = filters.product
    if (filters.status) params.status = filters.status
    const [vRes, bRes] = await Promise.all([getVersions(params), getAllBranches()])
    versions.value = vRes.data || []
    total.value = vRes.total || 0
    branches.value = bRes.data || []
  } finally { loading.value = false }
}

onMounted(load)
</script>
