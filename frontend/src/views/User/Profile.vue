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
import { formatPrice } from "@/lib/utils"
import { toast } from "vue-sonner"
import api, { resolveAssetUrl, uploadImage } from "@/lib/api"
import type { Order } from "@/types/order"

const userStore = useUserStore()
const router = useRouter()
const form = reactive({ user_name: "", avatar: "" })
const orderStats = reactive({ total: 0, pending: 0, paid: 0, loading: true })
const avatarUploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const isBusy = computed(() => userStore.loading)
const avatarPreview = computed(() => {
  const username = userStore.profile?.username || "guest"
  const raw = form.avatar || userStore.profile?.avatar
  return resolveAssetUrl(raw) || `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(username)}`
})
const userInitials = computed(() => (userStore.profile?.username || "U").slice(0, 2).toUpperCase())

const fetchOrderStats = async () => {
  orderStats.loading = true
  try {
    const res = await api.get<{ list: Order[]; total: number }, { list: Order[]; total: number }>("/orders", { params: { page: 1, page_size: 100 } })
    orderStats.total = res.total; orderStats.pending = res.list.filter((o) => o.status === 0).length; orderStats.paid = res.list.filter((o) => o.status === 1).length
  } catch { /* ignore */ } finally { orderStats.loading = false }
}
onMounted(() => { userStore.fetchProfile(); fetchOrderStats() })
watch(() => userStore.profile, (profile) => { if (!profile) return; form.user_name = profile.username; form.avatar = profile.avatar || "" }, { immediate: true })

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
  const target = event.target as HTMLInputElement | null; const file = target?.files?.[0]; if (!file) return
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
      <!-- 骨架屏 -->
      <div v-if="!userStore.profile" class="flex flex-col gap-6">
        <div class="h-10 w-1/3 animate-pulse bg-[#1C1C1C]/5"></div>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="i in 4" :key="i" class="h-24 animate-pulse border border-[#1C1C1C]/10"></div>
        </div>
      </div>

      <!-- 主内容 -->
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

        <!-- 四宫格 -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Balance</p>
            <div class="mt-2 text-2xl">{{ formatPrice(userStore.profile.balance) }}</div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Orders</p>
            <div class="mt-2 text-2xl">{{ orderStats.loading ? "--" : orderStats.total }} <span class="text-xs text-[#1C1C1C]/40">笔</span></div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Pending</p>
            <div class="mt-2 text-2xl">{{ orderStats.loading ? "--" : orderStats.pending }} <span class="text-xs text-[#1C1C1C]/40">笔</span></div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">VIP</p>
            <div class="mt-2"><Badge variant="outline">普通会员</Badge></div>
          </div>
        </div>

        <!-- 用户卡片 + 编辑 -->
        <div class="grid gap-8 lg:grid-cols-[1fr_1.2fr]">
          <!-- 用户卡片 -->
          <div class="border border-[#1C1C1C]/10 bg-white p-6">
            <div class="flex flex-col items-center gap-4 text-center">
              <Avatar class="h-28 w-28 border border-[#1C1C1C]/10">
                <AvatarImage :src="avatarPreview" :alt="userStore.profile.username" />
                <AvatarFallback class="bg-[#1C1C1C]/5 text-3xl">{{ userInitials }}</AvatarFallback>
              </Avatar>
              <div class="space-y-2">
                <h2 class="font-serif text-2xl tracking-tight">{{ userStore.profile.username }}</h2>
                <Badge variant="outline">活跃用户</Badge>
              </div>
              <p class="text-xs tracking-[0.12em] text-[#1C1C1C]/30">UID: {{ userStore.profile.id }}</p>
              <div class="my-2 h-px w-full bg-[#1C1C1C]/10"></div>
              <div class="grid w-full grid-cols-3 gap-4">
                <div class="text-center">
                  <div class="text-xl font-medium">{{ orderStats.total }}</div>
                  <div class="text-xs text-[#1C1C1C]/40">订单</div>
                </div>
                <div class="text-center">
                  <div class="text-xl font-medium">{{ orderStats.paid }}</div>
                  <div class="text-xs text-[#1C1C1C]/40">已完成</div>
                </div>
                <div class="text-center">
                  <div class="text-xl font-medium">{{ orderStats.pending }}</div>
                  <div class="text-xs text-[#1C1C1C]/40">待支付</div>
                </div>
              </div>
            </div>
          </div>

          <!-- 编辑 + 快捷入口 -->
          <div class="flex flex-col gap-6">
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

            <Card>
              <CardHeader class="pb-3">
                <CardTitle class="font-serif text-lg tracking-tight">快捷入口</CardTitle>
              </CardHeader>
              <CardContent>
                <div class="grid grid-cols-3 gap-3">
                  <RouterLink to="/orders" class="flex flex-col items-center gap-2 border border-[#1C1C1C]/10 p-4 transition-colors hover:border-[#1C1C1C]/30">
                    <ShoppingBag class="h-6 w-6 text-[#1C1C1C]/40" />
                    <span class="text-xs text-[#1C1C1C]/60">我的订单</span>
                  </RouterLink>
                  <RouterLink to="/vip" class="flex flex-col items-center gap-2 border border-[#1C1C1C]/10 p-4 transition-colors hover:border-[#1C1C1C]/30">
                    <Crown class="h-6 w-6 text-[#1C1C1C]/40" />
                    <span class="text-xs text-[#1C1C1C]/60">VIP 中心</span>
                  </RouterLink>
                  <RouterLink to="/products/publish" class="flex flex-col items-center gap-2 border border-[#1C1C1C]/10 p-4 transition-colors hover:border-[#1C1C1C]/30">
                    <Package class="h-6 w-6 text-[#1C1C1C]/40" />
                    <span class="text-xs text-[#1C1C1C]/60">发布商品</span>
                  </RouterLink>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        <div class="flex items-start gap-3 border border-[#1C1C1C]/10 p-4 text-sm">
          <ShieldCheck class="mt-0.5 h-4 w-4 shrink-0 text-[#1C1C1C]/40" />
          <p class="text-[#1C1C1C]/40">如遇登录异常请重新登录。修改用户名后，历史订单中的昵称不会自动更新。</p>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
