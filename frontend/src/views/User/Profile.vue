<script setup lang="ts">
import { onMounted } from "vue"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import { useUserStore } from "@/stores/userStore"
import { useRouter } from "vue-router"
import { formatPrice } from "@/lib/utils"

const userStore = useUserStore()
const router = useRouter()

onMounted(() => {
  userStore.fetchProfile()
})

const logout = () => {
  userStore.logout()
  router.push("/login")
}
</script>

<template>
  <MainLayout>
    <section class="relative mx-auto max-w-4xl px-6 py-12">
      <div class="pointer-events-none absolute inset-0 opacity-60 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-10 right-0 h-80 w-80 rounded-full bg-[#ea580c55] blur-3xl"></div>
      </div>
      <div class="relative grid gap-6 md:grid-cols-2">
        <Card class="border-obsidian-border/70 bg-obsidian-card/80">
          <CardHeader>
            <CardTitle class="text-2xl">账号信息</CardTitle>
            <CardDescription>登录凭据来自后端 JWT，401 将自动跳转登录。</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3 text-sm text-white/80">
            <div class="flex items-center justify-between">
              <span>用户名</span>
              <span class="font-semibold text-white">{{ userStore.profile?.username || "未加载" }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>余额</span>
              <span class="font-semibold text-magma">{{ formatPrice(userStore.profile?.balance || 0) }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>Token</span>
              <span class="truncate text-white/60" title="保存在 localStorage">jwt_token</span>
            </div>
            <MagmaButton class="w-full justify-center" @click="logout">退出登录</MagmaButton>
          </CardContent>
        </Card>

        <Card class="border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card">
          <CardHeader>
            <CardTitle class="text-xl">安全提示</CardTitle>
            <CardDescription>无 token 时访问受保护接口会被后端 401，前端自动清 token。</CardDescription>
          </CardHeader>
          <CardContent class="space-y-2 text-sm text-white/70">
            <p>· 重新登录会刷新本地缓存并显示最新余额/用户信息。</p>
            <p>· 访问发布/秒杀等接口前确保 token 存在，否则会跳转登录。</p>
            <p>· 如遇 401，可尝试重新登录或清除浏览器缓存。</p>
          </CardContent>
        </Card>
      </div>
    </section>
  </MainLayout>
</template>
