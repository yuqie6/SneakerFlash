<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue"
import { useRouter } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import { useUserStore } from "@/stores/userStore"
import { formatPrice } from "@/lib/utils"
import { toast } from "vue-sonner"
import { resolveAssetUrl, uploadImage } from "@/lib/api"

const userStore = useUserStore()
const router = useRouter()

const form = reactive({
  user_name: "",
  avatar: "",
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

onMounted(() => {
  userStore.fetchProfile()
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
    <section class="relative mx-auto max-w-5xl px-6 py-12">
      <div class="pointer-events-none absolute inset-0 opacity-60 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-10 right-0 h-80 w-80 rounded-full bg-[#ea580c55] blur-3xl"></div>
      </div>
      <div class="relative grid gap-6 lg:grid-cols-3">
        <Card class="border-obsidian-border/70 bg-obsidian-card/80">
          <CardHeader>
            <CardTitle class="text-2xl">账号信息</CardTitle>
            <CardDescription>登录凭据来自后端 JWT，401 将自动跳转登录。</CardDescription>
          </CardHeader>
          <CardContent class="space-y-5 text-sm text-white/80">
            <div class="flex items-center gap-3 rounded-2xl border border-obsidian-border/60 bg-black/30 p-3">
              <img :src="avatarPreview" class="h-14 w-14 rounded-full border border-obsidian-border/70 object-cover" alt="avatar" />
              <div>
                <div class="text-lg font-semibold text-white">{{ userStore.profile?.username || "未加载" }}</div>
                <div class="text-xs text-white/60">ID: {{ userStore.profile?.id || "--" }}</div>
              </div>
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

        <Card class="border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card lg:col-span-2">
          <CardHeader class="flex flex-col gap-1">
            <CardTitle class="text-xl">编辑资料</CardTitle>
            <CardDescription>支持修改头像与用户名，提交后将同步最新资料。</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div>
              <label class="mb-2 block text-sm text-white/70">头像 URL</label>
              <input
                v-model="form.avatar"
                type="url"
                placeholder="https://example.com/avatar.png"
                class="w-full rounded-lg border border-obsidian-border/60 bg-black/40 px-3 py-2 text-sm text-white outline-none transition focus:border-magma"
              />
              <div class="mt-2 flex flex-wrap items-center gap-3">
                <MagmaButton :disabled="avatarUploading" class="px-4 py-2" @click="fileInput?.click()">
                  {{ avatarUploading ? "上传中..." : "上传图片" }}
                </MagmaButton>
                <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onAvatarFileChange" />
                <p class="text-xs text-white/50">可直接粘贴图片链接，或上传文件自动填充。</p>
              </div>
            </div>
            <div>
              <label class="mb-2 block text-sm text-white/70">用户名</label>
              <input
                v-model="form.user_name"
                type="text"
                maxlength="50"
                class="w-full rounded-lg border border-obsidian-border/60 bg-black/40 px-3 py-2 text-sm text-white outline-none transition focus:border-magma"
              />
            </div>
            <div class="flex flex-wrap gap-3">
              <MagmaButton :disabled="isBusy" class="px-5 py-2" @click="submitProfile">
                {{ isBusy ? "保存中..." : "保存资料" }}
              </MagmaButton>
              <button
                class="rounded-full border border-obsidian-border px-4 py-2 text-sm text-white/80 transition hover:border-magma hover:text-magma"
                type="button"
                @click="resetForm"
              >
                重置
              </button>
            </div>
          </CardContent>
        </Card>

        <Card class="border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card lg:col-span-3">
          <CardHeader>
            <CardTitle class="text-xl">安全提示</CardTitle>
            <CardDescription>无 token 时访问受保护接口会被后端 401，前端自动清 token。</CardDescription>
          </CardHeader>
          <CardContent class="space-y-2 text-sm text-white/70">
            <p>· 重新登录会刷新本地缓存并显示最新余额/用户信息。</p>
            <p>· 修改用户名后请重新确认订单、发布信息是否使用新昵称。</p>
            <p>· 如遇 401，可尝试重新登录或清除浏览器缓存。</p>
          </CardContent>
        </Card>
      </div>
    </section>
  </MainLayout>
</template>
