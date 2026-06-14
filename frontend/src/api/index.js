import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

// 响应拦截：统一提取 data
api.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const msg = err.response?.data?.msg || err.message || '请求失败'
    return Promise.reject(new Error(msg))
  }
)

export default api
