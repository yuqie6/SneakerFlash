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
    token: localStorage.getItem("jwt_token") || "",
    profile: null as User | null,
    loading: false,
  }),
  actions: {
    setToken(token: string) {
      this.token = token
      if (token) {
        localStorage.setItem("jwt_token", token)
      } else {
        localStorage.removeItem("jwt_token")
      }
    },
    async login(payload: LoginPayload) {
      this.loading = true
      try {
        const res = await api.post<{ msg: string; token: string }, { msg: string; token: string }>(
          "/login",
          payload
        )
        if (res.token) {
          this.setToken(res.token)
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
      if (!this.token) return
      try {
        const res = await api.get<{ data: User }, { data: User }>("/profile")
        this.profile = res.data
      } catch {
        this.setToken("")
        this.profile = null
      }
    },
    logout() {
      this.setToken("")
      this.profile = null
    },
  },
})
