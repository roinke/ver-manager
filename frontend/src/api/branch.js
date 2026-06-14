import api from './index'

// 获取分支列表（分页）
export function getBranches(page = 1, pageSize = 20) {
  return api.get('/branches', { params: { page, page_size: pageSize } })
}

// 获取全部分支（不分页，供下拉框等使用）
export function getAllBranches() {
  return api.get('/branches', { params: { page_size: 9999 } })
}

// 获取单个分支
export function getBranch(id) {
  return api.get(`/branches/${id}`)
}

// 创建分支
export function createBranch(data) {
  return api.post('/branches', data)
}

// 更新分支
export function updateBranch(id, data) {
  return api.put(`/branches/${id}`, data)
}

// 停用分支
export function deleteBranch(id) {
  return api.delete(`/branches/${id}`)
}
