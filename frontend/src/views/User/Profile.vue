<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue"
import { RouterLink, useRouter } from "vue-router"
import {
  LogOut, Upload, RotateCcw, Save,
  ShieldCheck, ShoppingBag, Crown, Package,
} from "lucide-vue-next"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import { useUserStore } from "@/stores/userStore"
import { useProductStore } from "@/stores/productStore"
import { formatPrice } from "@/lib/utils"
import { toast } from "vue-sonner"
import api, { resolveAssetUrl, uploadImage } from "@/lib/api"
import type { VIPProfile } from "@/types/vip"

const userStore = useUserStore()
const productStore = useProductStore()
const router = useRouter()

const form = reactive({ user_name: "", avatar: "" })
const avatarUploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const isBusy = computed(() => userStore.loading)
const avatarPreview = computed(() => {
  const username = userStore.profile?.username || "guest"
  const raw = form.avatar || userStore.profile?.avatar
  return resolveAssetUrl(raw) || `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(username)}`
})
const userInitials = computed(() => (userStore.profile?.username || "U").slice(0, 2).toUpperCase())

const orderStats = reactive({ total: 0, pending: 0, loading: true })
const fetchOrderStats = async () => {
  orderStats.loading = true
  try {
    const res = await api.get<{ list: { status: number }[]; total: number }, { list: { status: number }[]; total: number }>("/orders", { params: { page: 1, page_size: 100 } })
    orderStats.total = res.total
    orderStats.pending = res.list.filter((o) => o.status === 0).length
  } catch { /* ignore */ } finally { orderStats.loading = false }
}

const vipProfile = ref<VIPProfile | null>(null)
const effectiveLevel = computed(() => vipProfile.value?.effective_level || 1)
const growthLevel = computed(() => vipProfile.value?.growth_level || 1)
const fetchVipProfile = async () => {
  try { vipProfile.value = await api.get<VIPProfile, VIPProfile>("/vip/profile") } catch { /* ignore */ }
}

const availableCouponCount = ref(0)
const fetchCouponCount = async () => {
  try {
    const res = await api.get<{ total: number }, { total: number }>("/coupons/mine", { params: { status: "available", page: 1, page_size: 1 } })
    availableCouponCount.value = res.total
  } catch { /* ignore */ }
}

const memberSince = computed(() => {
  const raw = userStore.profile?.created_at
  if (!raw) return ""
  const d = new Date(raw)
  return Number.isNaN(d.getTime()) ? "" : d.toLocaleDateString("zh-CN", { year: "numeric", month: "long", day: "numeric" })
})

onMounted(() => {
  userStore.fetchProfile()
  fetchOrderStats()
  fetchVipProfile()
  fetchCouponCount()
  productStore.fetchMyProducts(1, 1)
})
watch(() => userStore.profile, (p) => { if (!p) return; form.user_name = p.username; form.avatar = p.avatar || "" }, { immediate: true })

const submitProfile = async () => {
  if (!userStore.profile) return
  const payload: { user_name?: string; avatar?: string } = {}
  const name = form.user_name.trim()
  if (name && name !== userStore.profile.username) payload.user_name = name
  if (form.avatar !== userStore.profile.avatar) payload.avatar = form.avatar
  if (!payload.user_name && payload.avatar === undefined) { toast.error("请修改后再提交"); return }
  try { await userStore.updateProfile(payload) } catch (err: any) { toast.error(err?.message || "更新失败") }
}
const resetForm = () => { if (!userStore.profile) return; form.user_name = userStore.profile.username; form.avatar = userStore.profile.avatar || "" }
const onAvatarFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement | null
  const file = target?.files?.[0]
  if (!file) return
  avatarUploading.value = true
  try { const url = await uploadImage(file); form.avatar = url; toast.success("头像上传成功") }
  catch (err: any) { toast.error(err?.message || "上传失败") }
  finally { avatarUploading.value = false; if (target) target.value = "" }
}
const logout = () => { userStore.logout(); router.push("/login") }
</script>

<template>
  <MainLayout>
    <section class="mx-auto max-w-6xl px-6 py-16 md:py-24">
      <div v-if="!userStore.profile" class="flex flex-col gap-6">
        <div class="h-10 w-1/3 animate-pulse bg-[#1C1C1C]/5"></div>
        <div class="h-32 animate-pulse border border-[#1C1C1C]/10"></div>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="i in 4" :key="i" class="h-24 animate-pulse border border-[#1C1C1C]/10"></div>
        </div>
      </div>

      <div v-else class="flex flex-col gap-8">
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Profile</p>
            <h1 class="font-serif text-3xl tracking-tight">个人中心</h1>
          </div>
          <Button variant="ghost" class="w-fit gap-2 text-[#1C1C1C]/40 hover:text-[#1C1C1C]" @click="logout">
            <LogOut class="h-4 w-4" />
            退出登录
          </Button>
        </div>

        <div class="flex items-center gap-6 border border-[#1C1C1C]/10 bg-white p-6">
          <Avatar class="h-24 w-24 shrink-0 border border-[#1C1C1C]/10">
            <AvatarImage :src="avatarPreview" :alt="userStore.profile.username" />
            <AvatarFallback class="bg-[#1C1C1C]/5 text-3xl">{{ userInitials }}</AvatarFallback>
          </Avatar>
          <div class="min-w-0 flex-1 space-y-2">
            <h2 class="font-serif text-2xl tracking-tight">{{ userStore.profile.username }}</h2>
            <div class="flex flex-wrap items-center gap-3 text-xs text-[#1C1C1C]/40">
              <span>UID: {{ userStore.profile.id }}</span>
              <template v-if="memberSince">
                <span class="text-[#1C1C1C]/20">·</span>
                <span>注册于 {{ memberSince }}</span>
              </template>
            </div>
            <div class="flex flex-wrap items-center gap-3">
              <Badge variant="outline">VIP L{{ effectiveLevel }}</Badge>
              <span class="text-sm text-[#1C1C1C]/60">余额 {{ formatPrice(userStore.profile.balance) }}</span>
            </div>
          </div>
        </div>

        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Balance</p>
            <div class="mt-2 text-2xl">{{ formatPrice(userStore.profile.balance) }}</div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">VIP Level</p>
            <div class="mt-2 text-2xl">VIP L{{ effectiveLevel }}</div>
            <p class="mt-1 text-xs text-[#1C1C1C]/30">成长 L{{ growthLevel }}</p>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Orders</p>
            <div class="mt-2 text-2xl">{{ orderStats.loading ? "--" : orderStats.total }} <span class="text-xs text-[#1C1C1C]/40">笔</span></div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Coupons</p>
            <div class="mt-2 text-2xl">{{ availableCouponCount }} <span class="text-xs text-[#1C1C1C]/40">张可用</span></div>
          </div>
        </div>

        <div class="grid gap-4 sm:grid-cols-3">
          <RouterLink to="/orders" class="flex items-center gap-4 border border-[#1C1C1C]/10 bg-white p-5 transition-colors hover:border-[#1C1C1C]/30">
            <ShoppingBag class="h-6 w-6 shrink-0 text-[#1C1C1C]/40" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium">我的订单</p>
              <p class="text-xs text-[#1C1C1C]/40">{{ orderStats.total }} 笔订单 · {{ orderStats.pending }} 笔待支付</p>
            </div>
          </RouterLink>
          <RouterLink to="/vip" class="flex items-center gap-4 border border-[#1C1C1C]/10 bg-white p-5 transition-colors hover:border-[#1C1C1C]/30">
            <Crown class="h-6 w-6 shrink-0 text-[#1C1C1C]/40" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium">权益中心</p>
              <p class="text-xs text-[#1C1C1C]/40">VIP L{{ effectiveLevel }} · {{ availableCouponCount }} 张可用券</p>
            </div>
          </RouterLink>
          <RouterLink to="/products/publish" class="flex items-center gap-4 border border-[#1C1C1C]/10 bg-white p-5 transition-colors hover:border-[#1C1C1C]/30">
            <Package class="h-6 w-6 shrink-0 text-[#1C1C1C]/40" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium">发布商品</p>
              <p class="text-xs text-[#1C1C1C]/40">{{ productStore.myTotal }} 件已发布</p>
            </div>
          </RouterLink>
        </div>

        <Card>
          <CardHeader class="pb-4">
            <CardTitle class="font-serif text-xl tracking-tight">编辑资料</CardTitle>
            <CardDescription class="text-[#1C1C1C]/40">修改头像与用户名</CardDescription>
          </CardHeader>
          <CardContent class="space-y-5">
            <div class="flex items-center gap-4">
              <div class="group relative shrink-0">
                <Avatar class="h-16 w-16 border border-[#1C1C1C]/10 transition">
                  <AvatarImage :src="avatarPreview" alt="预览" />
                  <AvatarFallback class="bg-[#1C1C1C]/5 text-lg">{{ userInitials }}</AvatarFallback>
                </Avatar>
                <button type="button" class="absolute inset-0 flex items-center justify-center bg-white/80 opacity-0 transition group-hover:opacity-100" :disabled="avatarUploading" @click="fileInput?.click()">
                  <Upload class="h-4 w-4" />
                </button>
              </div>
              <div class="flex-1 space-y-2">
                <Input v-model="form.avatar" type="url" placeholder="https://example.com/avatar.png" />
                <Button variant="outline" size="sm" class="gap-2 text-xs" :disabled="avatarUploading" @click="fileInput?.click()">
                  <Upload class="h-3 w-3" />
                  {{ avatarUploading ? "上传中..." : "上传图片" }}
                </Button>
              </div>
            </div>
            <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onAvatarFileChange" />
            <div class="space-y-2">
              <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">用户名</label>
              <Input v-model="form.user_name" type="text" maxlength="50" placeholder="输入新用户名" />
            </div>
            <div class="flex flex-wrap gap-3 border-t border-[#1C1C1C]/10 pt-4">
              <MagmaButton :disabled="isBusy" class="gap-2 px-6" @click="submitProfile">
                <Save class="h-4 w-4" />
                {{ isBusy ? "保存中..." : "保存资料" }}
              </MagmaButton>
              <Button variant="outline" class="gap-2" @click="resetForm">
                <RotateCcw class="h-4 w-4" />
                重置
              </Button>
            </div>
          </CardContent>
        </Card>

        <div class="flex items-start gap-3 border border-[#1C1C1C]/10 p-4 text-sm">
          <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-[#1C1C1C]/40" />
          <p class="text-[#1C1C1C]/40">如遇登录异常请重新登录。修改用户名后，历史订单中的昵称不会自动更新。</p>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
