import axios from "axios"
import { toast } from "vue-sonner"

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8000/api/v1",
  timeout: 10000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("jwt_token")
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (response) => {
    if (response.data?.code && response.data.code !== 200) {
      return Promise.reject(new Error(response.data.msg || "操作失败"))
    }
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("jwt_token")
      window.location.href = "/login"
    }
    const msg = error.response?.data?.error || error.message || "系统繁忙"
    toast.error(msg)
    return Promise.reject(error)
  }
)

export default api
