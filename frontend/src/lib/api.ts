import axios from "axios"
import { toast } from "vue-sonner"

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8000/api/v1",
  timeout: 10000,
})

let isRefreshing = false
let refreshQueue: Array<(token: string | null) => void> = []

const getAccessToken = () => localStorage.getItem("access_token")
const getRefreshToken = () => localStorage.getItem("refresh_token")
const setAccessToken = (token: string) => localStorage.setItem("access_token", token)
const clearTokens = () => {
  localStorage.removeItem("access_token")
  localStorage.removeItem("refresh_token")
}

api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (response) => {
    const payload = response.data
    if (payload?.code !== undefined) {
      if (payload.code !== 200) {
        return Promise.reject(new Error(payload.msg || "操作失败"))
      }
      return payload.data
    }
    return payload
  },
  async (error) => {
    const { response, config } = error
    const originalRequest = config

    if (response?.status === 401) {
      const refreshToken = getRefreshToken()
      if (!refreshToken) {
        clearTokens()
        window.location.href = "/login"
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          refreshQueue.push((token) => {
            if (!token) {
              reject(error)
              return
            }
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(api(originalRequest))
          })
        })
      }

      isRefreshing = true
      try {
        const res = await axios.post(
          `${api.defaults.baseURL}/refresh`,
          { refresh_token: refreshToken },
          { timeout: 8000 }
        )
        const newToken = res.data?.access_token
        if (!newToken) {
          throw new Error("刷新失败")
        }
        setAccessToken(newToken)
        refreshQueue.forEach((cb) => cb(newToken))
        refreshQueue = []
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return api(originalRequest)
      } catch (refreshErr) {
        clearTokens()
        refreshQueue.forEach((cb) => cb(null))
        refreshQueue = []
        window.location.href = "/login"
        return Promise.reject(refreshErr)
      } finally {
        isRefreshing = false
      }
    }

    const msg = response?.data?.error || response?.data?.msg || error.message || "系统繁忙"
    toast.error(msg)
    return Promise.reject(error)
  }
)

export default api

export async function uploadImage(file: File) {
  const formData = new FormData()
  formData.append("file", file)
  const res = await api.post<{ url: string }, { url: string }>("/upload", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  })
  return resolveAssetUrl(res.url)
}

const assetOrigin = (() => {
  const base = api.defaults.baseURL || window.location.origin
  try {
    return new URL(base).origin
  } catch {
    return window.location.origin
  }
})()

export function resolveAssetUrl(src: string | undefined | null) {
  if (!src) return ""
  if (/^https?:\/\//i.test(src) || src.startsWith("data:")) return src
  if (src.startsWith("/")) return `${assetOrigin}${src}`
  return `${assetOrigin}/${src}`
}
