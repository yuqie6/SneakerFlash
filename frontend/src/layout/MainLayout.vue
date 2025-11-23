<script setup lang="ts">
import { RouterLink, useRouter } from "vue-router"
import { computed, onMounted } from "vue"
import { useUserStore } from "@/stores/userStore"
import MagmaButton from "@/components/motion/MagmaButton.vue"

const userStore = useUserStore()
const router = useRouter()

const isLoggedIn = computed(() => !!userStore.accessToken)
const username = computed(() => userStore.profile?.username || "Guest")
const avatar = computed(() => {
  const name = userStore.profile?.username || "guest"
  return userStore.profile?.avatar || `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(name)}`
})

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
  <div class="relative flex min-h-screen flex-col bg-obsidian-bg text-white">
    <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(249,115,22,0.12),transparent_40%),radial-gradient(circle_at_80%_0%,rgba(234,88,12,0.14),transparent_35%)]"></div>
    <header
      class="sticky top-0 z-30 border-b border-obsidian-border/70 bg-obsidian-bg/70 backdrop-blur-xl supports-[backdrop-filter]:bg-obsidian-bg/60"
    >
      <div class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <RouterLink to="/" class="flex items-center gap-2 text-lg font-semibold tracking-tight">
          <span class="flex h-9 w-9 items-center justify-center rounded-2xl bg-magma-gradient text-sm font-bold shadow-lg shadow-magma-glow/30">
            SF
          </span>
          <div class="leading-tight">
            <div>SneakerFlash</div>
            <p class="text-[11px] uppercase tracking-[0.3em] text-white/60">Midnight Magma</p>
          </div>
        </RouterLink>
        <nav class="flex items-center gap-4 text-sm text-white/80">
          <RouterLink class="transition hover:text-white" to="/">抢购大厅</RouterLink>
          <RouterLink v-if="isLoggedIn" class="transition hover:text-white" to="/orders">订单中心</RouterLink>
          <RouterLink v-if="isLoggedIn" class="transition hover:text-white" to="/products/publish">发布商品</RouterLink>
          <RouterLink v-if="isLoggedIn" class="transition hover:text-white" to="/profile">个人中心</RouterLink>
          <RouterLink v-if="!isLoggedIn" class="transition hover:text-white" to="/login">登录</RouterLink>
          <RouterLink v-if="!isLoggedIn" class="transition hover:text-white" to="/register">注册</RouterLink>
        </nav>
        <div class="flex items-center gap-3">
          <div
            v-if="isLoggedIn"
            class="flex items-center gap-2 rounded-full border border-obsidian-border/80 bg-obsidian-card/80 px-3 py-1 text-xs text-white/70"
          >
            <img :src="avatar" alt="avatar" class="h-8 w-8 rounded-full border border-obsidian-border/70 object-cover" />
            <span class="font-medium text-white">{{ username }}</span>
          </div>
          <MagmaButton v-if="!isLoggedIn" class="px-4 py-2 text-sm" @click="router.push('/login')">立即登录</MagmaButton>
          <button
            v-else
            class="rounded-full border border-obsidian-border px-4 py-2 text-sm text-white transition hover:border-magma hover:text-magma"
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
  </div>
</template>
