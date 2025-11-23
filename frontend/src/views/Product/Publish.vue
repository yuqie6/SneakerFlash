<script setup lang="ts">
import { reactive, computed, ref } from "vue"
import { useRouter } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import api, { resolveAssetUrl, uploadImage } from "@/lib/api"
import { toast } from "vue-sonner"
import { formatPrice } from "@/lib/utils"
import { useProductStore } from "@/stores/productStore"

const router = useRouter()
const productStore = useProductStore()

const form = reactive({
  name: "",
  price: "",
  stock: "",
  start_time: "",
  image: "",
})

const loading = reactive({ submitting: false })
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const preview = computed(() => ({
  name: form.name || "未命名球鞋",
  price: form.price ? Number(form.price) : 0,
  stock: form.stock ? Number(form.stock) : 0,
  start: form.start_time ? new Date(form.start_time).toLocaleString() : "未设置",
  image: resolveAssetUrl(form.image) || "/placeholder.svg",
}))

const submit = async () => {
  if (!form.name || !form.price || !form.stock || !form.start_time) {
    toast.error("请填写完整信息")
    return
  }
  const parsedTime = new Date(form.start_time)
  if (Number.isNaN(parsedTime.getTime())) {
    toast.error("开抢时间不合法")
    return
  }

  loading.submitting = true
  try {
    await api.post("/products", {
      name: form.name,
      price: Number(form.price),
      stock: Number(form.stock),
      start_time: parsedTime.toISOString(),
      image: form.image,
    })
    toast.success("商品发布成功，库存已预热")
    productStore.fetchProducts(1, 12)
    router.push("/")
  } catch (err: any) {
    toast.error(err?.message || "发布失败")
  } finally {
    loading.submitting = false
  }
}

const onImageSelected = async (event: Event) => {
  const target = event.target as HTMLInputElement | null
  const file = target?.files?.[0]
  if (!file) return

  uploading.value = true
  try {
    const url = await uploadImage(file)
    form.image = url
    toast.success("图片上传成功")
  } catch (err: any) {
    toast.error(err?.message || "上传失败")
  } finally {
    uploading.value = false
    if (target) target.value = ""
  }
}
</script>

<template>
  <MainLayout>
    <section class="relative mx-auto max-w-6xl px-6 py-12">
      <div class="pointer-events-none absolute inset-0 opacity-60 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-10 right-0 h-80 w-80 rounded-full bg-[#ea580c55] blur-3xl"></div>
      </div>
      <div class="relative grid gap-8 lg:grid-cols-[1.1fr_0.9fr]">
        <Card class="overflow-hidden border-obsidian-border/70 bg-obsidian-card/80">
          <CardHeader class="space-y-3">
            <p class="text-sm uppercase tracking-[0.3em] text-magma">Publish</p>
            <CardTitle class="text-3xl">发布商品</CardTitle>
            <CardDescription>写入后端并同步 Redis 库存，确保秒杀预热。</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <label class="text-sm text-white/70">商品名称</label>
              <Input v-model="form.name" placeholder="如：Air Zoom 1" class="bg-obsidian-card" />
            </div>
            <div class="grid gap-4 md:grid-cols-3">
              <div class="space-y-2">
                <label class="text-sm text-white/70">价格（元）</label>
                <Input v-model="form.price" type="number" min="0" step="0.01" placeholder="999.00" class="bg-obsidian-card" />
              </div>
              <div class="space-y-2">
                <label class="text-sm text-white/70">库存</label>
                <Input v-model="form.stock" type="number" min="0" step="1" placeholder="100" class="bg-obsidian-card" />
              </div>
              <div class="space-y-2">
                <label class="text-sm text-white/70">开抢时间</label>
                <Input v-model="form.start_time" type="datetime-local" class="bg-obsidian-card" />
              </div>
            </div>
            <div class="space-y-2">
              <label class="text-sm text-white/70">封面图（可选）</label>
              <Input v-model="form.image" type="url" placeholder="https://..." class="bg-obsidian-card" />
              <div class="flex flex-wrap items-center gap-3 text-xs text-white/60">
                <MagmaButton :disabled="uploading" class="px-3 py-1.5" @click="fileInput?.click()">
                  {{ uploading ? "上传中..." : "上传图片" }}
                </MagmaButton>
                <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onImageSelected" />
                <span>可填写外链，或上传图片自动生成链接</span>
              </div>
            </div>
            <MagmaButton class="w-full justify-center" :loading="loading.submitting" @click="submit">
              发布商品并预热
            </MagmaButton>
            <p class="text-xs text-white/60">发布成功将立即写库并调用库存预热逻辑（Redis 缓存）。</p>
          </CardContent>
        </Card>

        <Card class="overflow-hidden border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card">
          <CardHeader>
            <CardTitle class="text-xl">实时预览</CardTitle>
            <CardDescription>核对展示信息，避免秒杀页出现空图/错误数据。</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="relative overflow-hidden rounded-2xl border border-obsidian-border/70">
              <img :src="preview.image" alt="" class="h-60 w-full object-cover" />
              <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent"></div>
              <div class="absolute left-4 top-4 rounded-full bg-white/10 px-3 py-1 text-xs text-white">
                {{ preview.start }}
              </div>
              <div class="absolute bottom-0 left-0 right-0 p-5">
                <p class="text-sm text-white/70">预览 · 秒杀页</p>
                <h3 class="text-2xl font-semibold">{{ preview.name }}</h3>
                <p class="mt-2 text-magma text-lg">{{ formatPrice(preview.price) }}</p>
              </div>
            </div>
            <div class="rounded-2xl border border-obsidian-border/70 bg-black/30 p-4">
              <div class="flex items-center justify-between text-sm text-white/70">
                <span>库存</span>
                <span>{{ preview.stock }}</span>
              </div>
              <div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-obsidian-border">
                <div class="h-full bg-magma" :style="{ width: `${preview.stock > 0 ? 100 : 0}%` }"></div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </section>
  </MainLayout>
</template>
