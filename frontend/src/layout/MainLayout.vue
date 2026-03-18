<script setup lang="ts">
import { RouterLink, useRoute, useRouter } from "vue-router"
import { computed, onMounted } from "vue"
import { useUserStore } from "@/stores/userStore"
import { resolveAssetUrl } from "@/lib/api"
import MagmaButton from "@/components/motion/MagmaButton.vue"

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()

const isLoggedIn = computed(() => !!userStore.accessToken)
const username = computed(() => userStore.profile?.username || "Guest")
const avatar = computed(() => {
  const name = userStore.profile?.username || "guest"
  return resolveAssetUrl(userStore.profile?.avatar) || `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(name)}`
})

const isActive = (path: string) => {
  if (path === "/") return route.path === "/"
  return route.path === path || route.path.startsWith(`${path}/`)
}
const navClass = (path: string) => [
  "hover-underline pb-0.5 text-sm transition-colors",
  isActive(path) ? "text-[#1C1C1C] font-medium" : "text-[#1C1C1C]/60 hover:text-[#1C1C1C]",
]

const handleLogout = () => {
  userStore.logout()
  router.push("/login")
}

onMounted(() => {
  if (userStore.accessToken && !userStore.profile) {
    userStore.fetchProfile()
  }
})
</script>

<template>
  <div class="relative flex min-h-screen flex-col bg-[#F9F8F6] text-[#1C1C1C]">
    <header class="sticky top-0 z-30 border-b border-[#1C1C1C]/10 bg-[#F9F8F6]/90 backdrop-blur-sm">
      <div class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <RouterLink to="/" class="font-serif text-xl tracking-tight">
          SneakerFlash
        </RouterLink>
        <nav class="flex items-center gap-6">
          <RouterLink :class="navClass('/')" to="/">抢购大厅</RouterLink>
          <RouterLink v-if="isLoggedIn" :class="navClass('/orders')" to="/orders">订单</RouterLink>
          <RouterLink v-if="isLoggedIn" :class="navClass('/vip')" to="/vip">权益中心</RouterLink>
          <RouterLink v-if="isLoggedIn" :class="navClass('/profile')" to="/profile">个人中心</RouterLink>
          <RouterLink v-if="!isLoggedIn" :class="navClass('/login')" to="/login">登录</RouterLink>
          <RouterLink v-if="!isLoggedIn" :class="navClass('/register')" to="/register">注册</RouterLink>
        </nav>
        <div class="flex items-center gap-3">
          <div v-if="isLoggedIn" class="flex items-center gap-2 text-sm text-[#1C1C1C]/60">
            <img :src="avatar" alt="avatar" class="h-7 w-7 border border-[#1C1C1C]/10 object-cover" />
            <span>{{ username }}</span>
          </div>
          <MagmaButton v-if="!isLoggedIn" class="px-4 py-2 text-sm" @click="router.push('/login')">登录</MagmaButton>
          <button
            v-else
            class="border border-[#1C1C1C]/20 px-4 py-2 text-sm transition-colors hover:border-[#1C1C1C]"
            @click="handleLogout"
          >
            退出
          </button>
        </div>
      </div>
    </header>
    <main class="flex-1">
      <slot />
    </main>
    <footer class="border-t border-[#1C1C1C]/10">
      <div class="mx-auto flex max-w-6xl items-center justify-between px-6 py-8 text-xs text-[#1C1C1C]/40">
        <span class="font-serif">SneakerFlash</span>
        <span>限量球鞋 · 先到先得</span>
      </div>
    </footer>
  </div>
</template>
