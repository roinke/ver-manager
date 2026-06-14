import api from './index'

// 获取版本列表
export function getVersions(params) {
  return api.get('/versions', { params })
}

// 获取最新版本
export function getLatestVersions(product) {
  return api.get('/versions/latest', { params: { product } })
}

// 获取单个版本
export function getVersion(id) {
  return api.get(`/versions/${id}`)
}

// 创建版本
export function createVersion(data) {
  return api.post('/versions', data)
}

// 更新版本
export function updateVersion(id, data) {
  return api.put(`/versions/${id}`, data)
}

// 获取仪表盘数据
export function getDashboard() {
  return api.get('/dashboard')
}
