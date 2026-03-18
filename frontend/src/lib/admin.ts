import axios from "axios"

export function getAdminErrorMessage(error: unknown, fallback = "管理数据加载失败") {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status
    if (status === 401) return "登录状态已失效，请重新登录"
    if (status === 403) return "需要管理员权限"
    if (status === 404) return "管理接口暂不可用"

    const message = error.response?.data?.msg
    if (typeof message === "string" && message.trim()) return message
  }

  if (error instanceof Error && error.message.trim()) {
    return error.message
  }

  return fallback
}
