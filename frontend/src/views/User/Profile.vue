<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue"
import { RouterLink, useRouter } from "vue-router"
import {
  Wallet,
  LogOut,
  Upload,
  RotateCcw,
  Save,
  Camera,
  ShieldCheck,
  ShoppingBag,
  Crown,
  Sparkles,
  Package,
  CreditCard,
  User,
} from "lucide-vue-next"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import { useUserStore } from "@/stores/userStore"
import { formatPrice } from "@/lib/utils"
import { toast } from "vue-sonner"
import api, { resolveAssetUrl, uploadImage } from "@/lib/api"
import type { Order } from "@/types/order"

const userStore = useUserStore()
const router = useRouter()

const form = reactive({
  user_name: "",
  avatar: "",
})

const orderStats = reactive({
  total: 0,
  pending: 0,
  paid: 0,
  loading: true,
})

const avatarUploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const isBusy = computed(() => userStore.loading)
const avatarPreview = computed(() => {
  const username = userStore.profile?.username || "guest"
  const raw = form.avatar || userStore.profile?.avatar
  const fallback = `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(username)}`
  return resolveAssetUrl(raw) || fallback
})

const userInitials = computed(() => {
  const name = userStore.profile?.username || "U"
  return name.slice(0, 2).toUpperCase()
})

const fetchOrderStats = async () => {
  orderStats.loading = true
  try {
    const res = await api.get<
      { list: Order[]; total: number },
      { list: Order[]; total: number }
    >("/orders", { params: { page: 1, page_size: 100 } })
    orderStats.total = res.total
    orderStats.pending = res.list.filter((o) => o.status === 0).length
    orderStats.paid = res.list.filter((o) => o.status === 1).length
  } catch {
    /* ignore */
  } finally {
    orderStats.loading = false
  }
}

onMounted(() => {
  userStore.fetchProfile()
  fetchOrderStats()
})

watch(
  () => userStore.profile,
  (profile) => {
    if (!profile) return
    form.user_name = profile.username
    form.avatar = profile.avatar || ""
  },
  { immediate: true }
)

const submitProfile = async () => {
  if (!userStore.profile) return

  const payload: { user_name?: string; avatar?: string } = {}
  const name = form.user_name.trim()
  if (name && name !== userStore.profile.username) {
    payload.user_name = name
  }
  if (form.avatar !== userStore.profile.avatar) {
    payload.avatar = form.avatar
  }

  if (!payload.user_name && payload.avatar === undefined) {
    toast.error("请修改后再提交")
    return
  }

  try {
    await userStore.updateProfile(payload)
  } catch (err: any) {
    toast.error(err?.message || "更新失败")
  }
}

const resetForm = () => {
  if (!userStore.profile) return
  form.user_name = userStore.profile.username
  form.avatar = userStore.profile.avatar || ""
}

const onAvatarFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement | null
  const file = target?.files?.[0]
  if (!file) return

  avatarUploading.value = true
  try {
    const url = await uploadImage(file)
    form.avatar = url
    toast.success("头像上传成功")
  } catch (err: any) {
    toast.error(err?.message || "上传失败")
  } finally {
    avatarUploading.value = false
    if (target) target.value = ""
  }
}

const logout = () => {
  userStore.logout()
  router.push("/login")
}
</script>

<template>
  <MainLayout>
    <section class="relative mx-auto max-w-6xl px-6 py-12">
      <!-- 背景光效 -->
      <div class="pointer-events-none absolute inset-0 opacity-70 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-10 right-0 h-80 w-80 rounded-full bg-[#ea580c55] blur-3xl"></div>
      </div>

      <!-- 加载状态骨架屏 -->
      <div v-if="!userStore.profile" class="relative flex flex-col gap-6">
        <div class="h-10 w-1/3 animate-pulse rounded-lg bg-white/10"></div>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="i in 4" :key="i" class="h-24 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
        </div>
        <div class="grid gap-6 lg:grid-cols-[1fr_1.2fr]">
          <div class="h-80 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
          <div class="space-y-4">
            <div class="h-48 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
            <div class="h-28 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
          </div>
        </div>
      </div>

      <!-- 主要内容 -->
      <div v-else class="relative flex flex-col gap-6">
        <!-- 页面标题区域 -->
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div class="space-y-2">
            <p class="flex items-center gap-2 text-sm uppercase tracking-[0.3em] text-magma">
              <User class="h-4 w-4" />
              Profile Center
            </p>
            <h1 class="text-3xl font-semibold">个人中心</h1>
          </div>
          <Button variant="ghost" class="w-fit gap-2 text-white/60 hover:text-red-400" @click="logout">
            <LogOut class="h-4 w-4" />
            退出登录
          </Button>
        </div>

        <!-- 信息概览四宫格 -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="flex items-center gap-2 text-sm text-white/60">
              <Wallet class="h-4 w-4 text-magma" />
              账户余额
            </p>
            <div class="mt-2 text-2xl font-semibold text-magma">{{ formatPrice(userStore.profile.balance) }}</div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="flex items-center gap-2 text-sm text-white/60">
              <Package class="h-4 w-4 text-emerald-400" />
              全部订单
            </p>
            <div class="mt-2 flex items-center gap-2">
              <span class="text-2xl font-semibold text-white">{{ orderStats.loading ? "--" : orderStats.total }}</span>
              <span class="rounded-full border border-obsidian-border px-2 py-0.5 text-xs text-white/70">笔</span>
            </div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="flex items-center gap-2 text-sm text-white/60">
              <CreditCard class="h-4 w-4 text-amber-400" />
              待支付
            </p>
            <div class="mt-2 flex items-center gap-2">
              <span class="text-2xl font-semibold text-amber-400">{{ orderStats.loading ? "--" : orderStats.pending }}</span>
              <span class="rounded-full border border-amber-500/40 px-2 py-0.5 text-xs text-amber-400/70">笔</span>
            </div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="flex items-center gap-2 text-sm text-white/60">
              <Crown class="h-4 w-4 text-purple-400" />
              会员等级
            </p>
            <div class="mt-2">
              <Badge variant="outline" class="border-purple-500/40 bg-purple-500/10 text-purple-300">普通会员</Badge>
            </div>
          </div>
        </div>

        <!-- 主要内容区：左侧用户卡片 + 右侧编辑和快捷入口 -->
        <div class="grid gap-6 lg:grid-cols-[1fr_1.2fr]">
          <!-- 左侧：用户卡片 -->
          <ParallaxCard class="glass h-fit">
            <div class="relative overflow-hidden rounded-2xl border border-obsidian-border/70 bg-gradient-to-br from-obsidian-card via-black to-obsidian-card p-6">
              <!-- 背景装饰 -->
              <div class="pointer-events-none absolute inset-0">
                <div class="absolute -right-10 -top-10 h-40 w-40 rounded-full bg-magma/20 blur-3xl"></div>
              </div>

              <div class="relative flex flex-col items-center gap-4 text-center">
                <!-- 头像 -->
                <div class="group relative">
                  <Avatar class="h-28 w-28 border-4 border-obsidian-border/70 shadow-lg shadow-magma/20">
                    <AvatarImage :src="avatarPreview" :alt="userStore.profile.username" />
                    <AvatarFallback class="bg-magma/20 text-3xl text-magma">{{ userInitials }}</AvatarFallback>
                  </Avatar>
                  <div class="absolute -bottom-1 -right-1 rounded-full border-2 border-obsidian-card bg-emerald-500 p-1.5">
                    <div class="h-2 w-2 rounded-full bg-white"></div>
                  </div>
                </div>

                <!-- 用户名 -->
                <div class="space-y-2">
                  <h2 class="text-2xl font-semibold">{{ userStore.profile.username }}</h2>
                  <div class="flex items-center justify-center gap-2">
                    <Badge variant="outline" class="border-magma/40 bg-magma/10 text-magma">
                      <Sparkles class="mr-1 h-3 w-3" />
                      活跃用户
                    </Badge>
                  </div>
                </div>

                <!-- 用户ID -->
                <p class="font-mono text-xs text-white/50">UID: {{ userStore.profile.id }}</p>

                <!-- 分隔线 -->
                <div class="my-2 h-px w-full bg-gradient-to-r from-transparent via-obsidian-border to-transparent"></div>

                <!-- 统计数据 -->
                <div class="grid w-full grid-cols-3 gap-4">
                  <div class="text-center">
                    <div class="text-xl font-semibold text-magma">{{ orderStats.total }}</div>
                    <div class="text-xs text-white/50">订单</div>
                  </div>
                  <div class="text-center">
                    <div class="text-xl font-semibold text-emerald-400">{{ orderStats.paid }}</div>
                    <div class="text-xs text-white/50">已完成</div>
                  </div>
                  <div class="text-center">
                    <div class="text-xl font-semibold text-amber-400">{{ orderStats.pending }}</div>
                    <div class="text-xs text-white/50">待支付</div>
                  </div>
                </div>
              </div>
            </div>
          </ParallaxCard>

          <!-- 右侧：编辑资料和快捷入口 -->
          <div class="flex flex-col gap-4">
            <!-- 编辑资料卡片 -->
            <Card class="border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card">
              <CardHeader class="pb-4">
                <CardTitle class="flex items-center gap-2 text-xl">
                  <Camera class="h-5 w-5 text-magma" />
                  编辑资料
                </CardTitle>
                <CardDescription>修改头像与用户名</CardDescription>
              </CardHeader>
              <CardContent class="space-y-5">
                <!-- 头像编辑区 -->
                <div class="flex items-center gap-4">
                  <div class="group relative shrink-0">
                    <Avatar class="h-16 w-16 border-2 border-obsidian-border/70 transition group-hover:border-magma/50">
                      <AvatarImage :src="avatarPreview" alt="预览" />
                      <AvatarFallback class="bg-magma/20 text-lg text-magma">{{ userInitials }}</AvatarFallback>
                    </Avatar>
                    <button
                      type="button"
                      class="absolute inset-0 flex items-center justify-center rounded-full bg-black/60 opacity-0 transition group-hover:opacity-100"
                      :disabled="avatarUploading"
                      @click="fileInput?.click()"
                    >
                      <Upload class="h-4 w-4 text-white" />
                    </button>
                  </div>
                  <div class="flex-1 space-y-2">
                    <Input
                      v-model="form.avatar"
                      type="url"
                      placeholder="https://example.com/avatar.png"
                      class="border-obsidian-border/60 bg-black/40 text-sm text-white placeholder:text-white/40 focus-visible:ring-magma"
                    />
                    <Button
                      variant="outline"
                      size="sm"
                      class="gap-2 border-obsidian-border text-xs text-white/80 hover:border-magma hover:text-magma"
                      :disabled="avatarUploading"
                      @click="fileInput?.click()"
                    >
                      <Upload class="h-3 w-3" />
                      {{ avatarUploading ? "上传中..." : "上传图片" }}
                    </Button>
                  </div>
                </div>
                <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onAvatarFileChange" />

                <!-- 用户名编辑 -->
                <div class="space-y-2">
                  <label class="block text-sm font-medium text-white/70">用户名</label>
                  <Input
                    v-model="form.user_name"
                    type="text"
                    maxlength="50"
                    placeholder="输入新用户名"
                    class="border-obsidian-border/60 bg-black/40 text-white placeholder:text-white/40 focus-visible:ring-magma"
                  />
                </div>

                <!-- 操作按钮 -->
                <div class="flex flex-wrap gap-3 border-t border-obsidian-border/40 pt-4">
                  <MagmaButton :disabled="isBusy" class="gap-2 px-6" @click="submitProfile">
                    <Save class="h-4 w-4" />
                    {{ isBusy ? "保存中..." : "保存资料" }}
                  </MagmaButton>
                  <Button
                    variant="outline"
                    class="gap-2 border-obsidian-border text-white/80 hover:border-magma hover:text-magma"
                    @click="resetForm"
                  >
                    <RotateCcw class="h-4 w-4" />
                    重置
                  </Button>
                </div>
              </CardContent>
            </Card>

            <!-- 快捷入口 -->
            <Card class="border-obsidian-border/70 bg-obsidian-card/80">
              <CardHeader class="pb-3">
                <CardTitle class="flex items-center gap-2 text-lg">
                  <Sparkles class="h-4 w-4 text-magma" />
                  快捷入口
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div class="grid grid-cols-3 gap-3">
                  <RouterLink
                    to="/orders"
                    class="flex flex-col items-center gap-2 rounded-xl border border-obsidian-border/70 bg-black/30 p-4 transition hover:border-magma/50 hover:bg-magma/5"
                  >
                    <ShoppingBag class="h-6 w-6 text-magma" />
                    <span class="text-xs text-white/70">我的订单</span>
                  </RouterLink>
                  <RouterLink
                    to="/vip"
                    class="flex flex-col items-center gap-2 rounded-xl border border-obsidian-border/70 bg-black/30 p-4 transition hover:border-purple-500/50 hover:bg-purple-500/5"
                  >
                    <Crown class="h-6 w-6 text-purple-400" />
                    <span class="text-xs text-white/70">VIP 中心</span>
                  </RouterLink>
                  <RouterLink
                    to="/products/publish"
                    class="flex flex-col items-center gap-2 rounded-xl border border-obsidian-border/70 bg-black/30 p-4 transition hover:border-emerald-500/50 hover:bg-emerald-500/5"
                  >
                    <Package class="h-6 w-6 text-emerald-400" />
                    <span class="text-xs text-white/70">发布商品</span>
                  </RouterLink>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        <!-- 安全提示 -->
        <div class="flex items-start gap-3 rounded-xl border border-obsidian-border/50 bg-obsidian-card/50 p-4 text-sm">
          <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-400" />
          <p class="text-white/60">
            如遇登录异常请重新登录。修改用户名后，历史订单中的昵称不会自动更新。
          </p>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
