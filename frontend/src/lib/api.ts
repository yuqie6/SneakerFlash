import axios, { AxiosHeaders, type InternalAxiosRequestConfig } from "axios"
import { toast } from "vue-sonner"

export const apiBaseURL = import.meta.env.VITE_API_BASE_URL || "/api/v1"

const api = axios.create({
  baseURL: apiBaseURL,
  timeout: 10000,
})

let isRefreshing = false
let refreshQueue: Array<(token: string | null) => void> = []

type RetryableRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean
}

const getAccessToken = () => localStorage.getItem("access_token")
const getRefreshToken = () => localStorage.getItem("refresh_token")
const setAccessToken = (token: string) => localStorage.setItem("access_token", token)
const clearTokens = () => {
  localStorage.removeItem("access_token")
  localStorage.removeItem("refresh_token")
}

const ensureHeaders = (config: InternalAxiosRequestConfig) => {
  if (!config.headers) {
    config.headers = new AxiosHeaders()
  }
  return config.headers
}

const setAuthorizationHeader = (config: InternalAxiosRequestConfig, token: string) => {
  ensureHeaders(config).set("Authorization", `Bearer ${token}`)
}

const unwrapPayload = <T>(payload: unknown): T => {
  if (payload && typeof payload === "object" && "code" in payload) {
    const apiPayload = payload as { code: number; msg?: string; data?: T }
    if (apiPayload.code !== 200) {
      throw new Error(apiPayload.msg || "操作失败")
    }
    return apiPayload.data as T
  }
  return payload as T
}

export const redirectToLogin = () => {
  if (import.meta.env.MODE === "test") return
  const redirect = `${window.location.pathname}${window.location.search}${window.location.hash}`
  window.location.href = `/login?redirect=${encodeURIComponent(redirect || "/")}`
}

api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) setAuthorizationHeader(config, token)
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
    const originalRequest = config as RetryableRequestConfig | undefined

    if (response?.status === 401) {
      const refreshToken = getRefreshToken()
      const isRefreshRequest = originalRequest?.url?.endsWith("/refresh")
      if (!refreshToken || !originalRequest || originalRequest._retry || isRefreshRequest) {
        clearTokens()
        redirectToLogin()
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          refreshQueue.push((token) => {
            if (!token) {
              reject(error)
              return
            }
            setAuthorizationHeader(originalRequest, token)
            resolve(api(originalRequest))
          })
        })
      }

      isRefreshing = true
      originalRequest._retry = true
      try {
        const res = await axios.post(
          `${api.defaults.baseURL}/refresh`,
          { refresh_token: refreshToken },
          { timeout: 8000 }
        )
        const refreshData = unwrapPayload<{ access_token?: string }>(res.data)
        const newToken = refreshData?.access_token
        if (!newToken) {
          throw new Error("刷新失败")
        }
        setAccessToken(newToken)
        refreshQueue.forEach((cb) => cb(newToken))
        refreshQueue = []
        setAuthorizationHeader(originalRequest, newToken)
        return api(originalRequest)
      } catch (refreshErr) {
        clearTokens()
        refreshQueue.forEach((cb) => cb(null))
        refreshQueue = []
        redirectToLogin()
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

export function resolveAssetUrl(src: string | undefined | null) {
  if (!src) return ""
  // 绝对URL或Data URI直接返回
  if (/^https?:\/\//i.test(src) || src.startsWith("data:")) return src
  // 相对路径直接返回（由Vite代理处理）
  if (src.startsWith("/")) return src
  return `/${src}`
}

export function buildStreamUrl(path: string, accessToken?: string) {
  const base = apiBaseURL.startsWith("http")
    ? apiBaseURL
    : `${window.location.origin}${apiBaseURL.startsWith("/") ? apiBaseURL : `/${apiBaseURL}`}`
  const url = new URL(`${base}${path.startsWith("/") ? path : `/${path}`}`)
  if (accessToken) url.searchParams.set("access_token", accessToken)
  return url.toString()
}
