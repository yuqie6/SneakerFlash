import { defineStore } from "pinia"
import { toast } from "vue-sonner"
import api from "@/lib/api"
import type { User } from "@/types/user"

type LoginPayload = {
  user_name: string
  user_password: string
}

export const useUserStore = defineStore("user", {
  state: () => ({
    accessToken: localStorage.getItem("access_token") || localStorage.getItem("jwt_token") || "",
    refreshToken: localStorage.getItem("refresh_token") || "",
    profile: null as User | null,
    loading: false,
  }),
  getters: {
    permissions: (state) => state.profile?.permissions || [],
    isAdmin: (state) => (state.profile?.permissions?.length || 0) > 0,
    hasPermission: (state) => (permission: string) => (state.profile?.permissions || []).includes(permission),
  },
  actions: {
    setTokens(access: string, refresh?: string) {
      this.accessToken = access
      if (refresh !== undefined) {
        this.refreshToken = refresh
      } else if (!access) {
        this.refreshToken = ""
      }
      if (access) {
        localStorage.setItem("access_token", access)
      } else {
        localStorage.removeItem("access_token")
      }
      if (refresh) {
        localStorage.setItem("refresh_token", refresh)
      } else if (!access) {
        localStorage.removeItem("refresh_token")
      }
    },
    async login(payload: LoginPayload) {
      this.loading = true
      try {
        type LoginResp = { access_token: string; refresh_token: string; expires_in: number }
        const res = (await api.post<LoginResp>("/login", payload)) as unknown as LoginResp

        if (res.access_token) {
          this.setTokens(res.access_token, res.refresh_token)
          const profile = await this.fetchProfile()
          if (!profile) {
            const err = new Error("登录状态校验失败，请重试")
            toast.error(err.message)
            throw err
          }
          toast.success("登录成功")
        }
        return res
      } finally {
        this.loading = false
      }
    },
    async register(payload: LoginPayload) {
      this.loading = true
      try {
        await api.post("/register", payload)
        toast.success("注册成功")
      } finally {
        this.loading = false
      }
    },
    async fetchProfile() {
      if (!this.accessToken) return null
      try {
        const res = await api.get<User, User>("/profile")
        this.profile = res
        return res
      } catch {
        this.setTokens("")
        this.profile = null
        return null
      }
    },
    async refreshTokenIfNeeded() {
      if (!this.refreshToken) return
      type RefreshResp = { access_token: string; expires_in: number }
      const res = (await api.post<RefreshResp>("/refresh", { refresh_token: this.refreshToken })) as unknown as RefreshResp
      if (res.access_token) this.setTokens(res.access_token)
    },
    async updateProfile(payload: { user_name?: string; avatar?: string }) {
      this.loading = true
      try {
        const res = await api.put<User, User>("/profile", payload)
        this.profile = res
        toast.success("资料已更新")
      } finally {
        this.loading = false
      }
    },
    logout() {
      this.setTokens("")
      this.profile = null
    },
  },
})
