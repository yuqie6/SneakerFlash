<script setup lang="ts">
import { reactive, computed, ref, onMounted } from "vue"
import { useRouter } from "vue-router"
import {
  Package,
  Upload,
  Eye,
  Clock,
  DollarSign,
  Boxes,
  ImageIcon,
  Tag,
  Sparkles,
  Edit3,
  Trash2,
  RotateCcw,
  Zap,
  CheckCircle,
  AlertCircle,
} from "lucide-vue-next"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
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
  end_time: "",
  image: "",
})
const editingId = ref<number | null>(null)

const loading = reactive({ submitting: false })
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const placeholderImg = "https://dummyimage.com/900x600/0f0f14/ffffff&text=SneakerFlash"
const preview = computed(() => ({
  name: form.name || "未命名球鞋",
  price: form.price ? Number(form.price) : 0,
  stock: form.stock ? Number(form.stock) : 0,
  start: form.start_time ? new Date(form.start_time).toLocaleString() : "未设置",
  image: resolveAssetUrl(form.image) || placeholderImg,
}))

const stats = computed(() => {
  const items = productStore.myItems
  return {
    total: items.length,
    totalStock: items.reduce((sum, p) => sum + (p.stock || 0), 0),
    active: items.filter((p) => new Date(p.start_time) <= new Date()).length,
    upcoming: items.filter((p) => new Date(p.start_time) > new Date()).length,
  }
})

const formProgress = computed(() => {
  let filled = 0
  if (form.name) filled++
  if (form.price) filled++
  if (form.stock) filled++
  if (form.start_time) filled++
  return (filled / 4) * 100
})

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
    const payload: Record<string, unknown> = {
      name: form.name,
      price: Number(form.price),
      stock: Number(form.stock),
      start_time: parsedTime.toISOString(),
      image: form.image,
    }
    // 结束时间（可选）
    if (form.end_time) {
      const endParsed = new Date(form.end_time)
      if (Number.isNaN(endParsed.getTime())) {
        toast.error("结束时间不合法")
        loading.submitting = false
        return
      }
      if (endParsed <= parsedTime) {
        toast.error("结束时间必须晚于开抢时间")
        loading.submitting = false
        return
      }
      payload.end_time = endParsed.toISOString()
    }
    if (editingId.value) {
      await api.put(`/products/${editingId.value}`, payload)
      toast.success("商品更新成功")
    } else {
      await api.post("/products", payload)
      toast.success("商品发布成功，库存已预热")
    }
    productStore.fetchProducts(1, 12)
    productStore.fetchMyProducts()
    if (!editingId.value) router.push("/")
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

const startEdit = (p: any) => {
  editingId.value = p.id
  form.name = p.name
  form.price = String(p.price)
  form.stock = String(p.stock)
  form.start_time = p.start_time?.slice(0, 16) || ""
  form.end_time = p.end_time?.slice(0, 16) || ""
  form.image = p.image || ""
}

const resetForm = () => {
  editingId.value = null
  form.name = ""
  form.price = ""
  form.stock = ""
  form.start_time = ""
  form.end_time = ""
  form.image = ""
}

onMounted(() => {
  productStore.fetchMyProducts()
})

const deleteProduct = async (id: number) => {
  try {
    await productStore.deleteProduct(id)
    toast.success("已删除")
    if (editingId.value === id) resetForm()
  } catch (err: any) {
    toast.error(err?.message || "删除失败")
  }
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

      <div class="relative flex flex-col gap-6">
        <!-- 页面标题区域 -->
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div class="space-y-2">
            <p class="flex items-center gap-2 text-sm uppercase tracking-[0.3em] text-magma">
              <Package class="h-4 w-4" />
              Product Publish
            </p>
            <h1 class="text-3xl font-semibold">{{ editingId ? "编辑商品" : "发布商品" }}</h1>
          </div>
          <div class="flex items-center gap-3">
            <Badge v-if="editingId" variant="outline" class="border-amber-500/40 bg-amber-500/10 text-amber-300">
              <Edit3 class="mr-1 h-3 w-3" />
              编辑模式
            </Badge>
            <Badge v-else variant="outline" class="border-emerald-500/40 bg-emerald-500/10 text-emerald-300">
              <Sparkles class="mr-1 h-3 w-3" />
              新品发布
            </Badge>
          </div>
        </div>

        <!-- 信息概览四宫格 -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="flex items-center gap-2 text-sm text-white/60">
              <Package class="h-4 w-4 text-magma" />
              已发布商品
            </p>
            <div class="mt-2 flex items-center gap-2">
              <span class="text-2xl font-semibold text-white">{{ stats.total }}</span>
              <span class="rounded-full border border-obsidian-border px-2 py-0.5 text-xs text-white/70">件</span>
            </div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="flex items-center gap-2 text-sm text-white/60">
              <Boxes class="h-4 w-4 text-emerald-400" />
              总库存量
            </p>
            <div class="mt-2 flex items-center gap-2">
              <span class="text-2xl font-semibold text-emerald-400">{{ stats.totalStock }}</span>
              <span class="rounded-full border border-emerald-500/40 px-2 py-0.5 text-xs text-emerald-400/70">件</span>
            </div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="flex items-center gap-2 text-sm text-white/60">
              <Zap class="h-4 w-4 text-amber-400" />
              抢购中
            </p>
            <div class="mt-2 flex items-center gap-2">
              <span class="text-2xl font-semibold text-amber-400">{{ stats.active }}</span>
              <span class="rounded-full border border-amber-500/40 px-2 py-0.5 text-xs text-amber-400/70">件</span>
            </div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="flex items-center gap-2 text-sm text-white/60">
              <Clock class="h-4 w-4 text-purple-400" />
              即将开始
            </p>
            <div class="mt-2 flex items-center gap-2">
              <span class="text-2xl font-semibold text-purple-400">{{ stats.upcoming }}</span>
              <span class="rounded-full border border-purple-500/40 px-2 py-0.5 text-xs text-purple-400/70">件</span>
            </div>
          </div>
        </div>

        <!-- 主要内容区：表单 + 预览 -->
        <div class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <!-- 左侧：发布表单 -->
          <Card class="overflow-hidden border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card">
            <CardHeader class="space-y-3 pb-4">
              <div class="flex items-center justify-between">
                <CardTitle class="flex items-center gap-2 text-xl">
                  <Tag class="h-5 w-5 text-magma" />
                  商品信息
                </CardTitle>
                <div class="flex items-center gap-2 text-xs text-white/50">
                  <span>完成度</span>
                  <div class="h-1.5 w-16 overflow-hidden rounded-full bg-obsidian-border">
                    <div
                      class="h-full rounded-full bg-magma transition-all duration-300"
                      :style="{ width: `${formProgress}%` }"
                    ></div>
                  </div>
                  <span class="text-magma">{{ formProgress }}%</span>
                </div>
              </div>
              <CardDescription>填写商品信息，发布后即可参与秒杀活动</CardDescription>
            </CardHeader>
            <CardContent class="space-y-5">
              <!-- 商品名称 -->
              <div class="space-y-2">
                <label class="flex items-center gap-2 text-sm font-medium text-white/70">
                  <Tag class="h-3.5 w-3.5 text-magma/70" />
                  商品名称
                  <span class="text-red-400">*</span>
                </label>
                <Input
                  v-model="form.name"
                  placeholder="如：Air Zoom 1"
                  class="border-obsidian-border/60 bg-black/40 text-white placeholder:text-white/40 focus-visible:ring-magma"
                />
              </div>

              <!-- 价格、库存、时间三列 -->
              <div class="grid gap-4 md:grid-cols-3">
                <div class="space-y-2">
                  <label class="flex items-center gap-2 text-sm font-medium text-white/70">
                    <DollarSign class="h-3.5 w-3.5 text-magma/70" />
                    价格（元）
                    <span class="text-red-400">*</span>
                  </label>
                  <Input
                    v-model="form.price"
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="999.00"
                    class="border-obsidian-border/60 bg-black/40 text-white placeholder:text-white/40 focus-visible:ring-magma"
                  />
                </div>
                <div class="space-y-2">
                  <label class="flex items-center gap-2 text-sm font-medium text-white/70">
                    <Boxes class="h-3.5 w-3.5 text-magma/70" />
                    库存
                    <span class="text-red-400">*</span>
                  </label>
                  <Input
                    v-model="form.stock"
                    type="number"
                    min="0"
                    step="1"
                    placeholder="100"
                    class="border-obsidian-border/60 bg-black/40 text-white placeholder:text-white/40 focus-visible:ring-magma"
                  />
                </div>
                <div class="space-y-2">
                  <label class="flex items-center gap-2 text-sm font-medium text-white/70">
                    <Clock class="h-3.5 w-3.5 text-magma/70" />
                    开抢时间
                    <span class="text-red-400">*</span>
                  </label>
                  <Input
                    v-model="form.start_time"
                    type="datetime-local"
                    class="border-obsidian-border/60 bg-black/40 text-white placeholder:text-white/40 focus-visible:ring-magma"
                  />
                </div>
              </div>

              <!-- 结束时间（可选） -->
              <div class="space-y-2">
                <label class="flex items-center gap-2 text-sm font-medium text-white/70">
                  <Clock class="h-3.5 w-3.5 text-purple-400/70" />
                  结束时间（可选）
                </label>
                <Input
                  v-model="form.end_time"
                  type="datetime-local"
                  class="border-obsidian-border/60 bg-black/40 text-white placeholder:text-white/40 focus-visible:ring-magma"
                />
                <p class="text-xs text-white/40">不设置则表示活动永不过期</p>
              </div>

              <!-- 封面图 -->
              <div class="space-y-3">
                <label class="flex items-center gap-2 text-sm font-medium text-white/70">
                  <ImageIcon class="h-3.5 w-3.5 text-magma/70" />
                  封面图（可选）
                </label>
                <Input
                  v-model="form.image"
                  type="url"
                  placeholder="https://..."
                  class="border-obsidian-border/60 bg-black/40 text-white placeholder:text-white/40 focus-visible:ring-magma"
                />
                <div class="flex flex-wrap items-center gap-3">
                  <Button
                    variant="outline"
                    size="sm"
                    class="gap-2 border-obsidian-border text-white/80 hover:border-magma hover:text-magma"
                    :disabled="uploading"
                    @click="fileInput?.click()"
                  >
                    <Upload class="h-3.5 w-3.5" />
                    {{ uploading ? "上传中..." : "上传图片" }}
                  </Button>
                  <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onImageSelected" />
                  <span class="text-xs text-white/50">可填写外链，或上传图片自动生成链接</span>
                </div>
              </div>

              <!-- 操作按钮 -->
              <div class="flex flex-wrap gap-3 border-t border-obsidian-border/40 pt-5">
                <MagmaButton class="flex-1 justify-center gap-2" :loading="loading.submitting" @click="submit">
                  <CheckCircle v-if="!loading.submitting" class="h-4 w-4" />
                  {{ editingId ? "保存修改" : "发布商品并预热" }}
                </MagmaButton>
                <Button
                  v-if="editingId"
                  variant="outline"
                  class="flex-1 gap-2 border-obsidian-border text-white/80 hover:border-magma hover:text-magma"
                  @click="resetForm"
                >
                  <RotateCcw class="h-4 w-4" />
                  取消编辑
                </Button>
              </div>

              <!-- 提示信息 -->
              <div class="flex items-start gap-2 rounded-lg border border-obsidian-border/50 bg-black/20 p-3 text-xs text-white/50">
                <AlertCircle class="mt-0.5 h-3.5 w-3.5 shrink-0 text-magma/70" />
                <span>发布成功后商品将立即上架，用户可在开抢时间参与秒杀。</span>
              </div>
            </CardContent>
          </Card>

          <!-- 右侧：实时预览 -->
          <div class="flex flex-col gap-4">
            <ParallaxCard class="glass">
              <div class="overflow-hidden rounded-2xl border border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card">
                <div class="border-b border-obsidian-border/50 p-4">
                  <div class="flex items-center gap-2 text-lg font-semibold">
                    <Eye class="h-5 w-5 text-magma" />
                    实时预览
                  </div>
                  <p class="mt-1 text-xs text-white/50">核对展示信息，避免秒杀页出现空图/错误数据</p>
                </div>
                <div class="p-4">
                  <div class="relative overflow-hidden rounded-xl border border-obsidian-border/70">
                    <img :src="preview.image" alt="" class="h-52 w-full object-cover" />
                    <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent"></div>
                    <div class="absolute left-3 top-3 flex items-center gap-2 rounded-full bg-white/10 px-3 py-1 text-xs text-white backdrop-blur-sm">
                      <Clock class="h-3 w-3" />
                      {{ preview.start }}
                    </div>
                    <div class="absolute bottom-0 left-0 right-0 p-4">
                      <p class="text-xs text-white/60">预览 · 秒杀页</p>
                      <h3 class="mt-1 text-xl font-semibold">{{ preview.name }}</h3>
                      <p class="mt-1.5 text-lg font-semibold text-magma">{{ formatPrice(preview.price) }}</p>
                    </div>
                  </div>
                </div>
                <div class="border-t border-obsidian-border/50 p-4">
                  <div class="rounded-xl border border-obsidian-border/70 bg-black/30 p-4">
                    <div class="flex items-center justify-between text-sm">
                      <span class="flex items-center gap-2 text-white/70">
                        <Boxes class="h-4 w-4 text-magma" />
                        库存进度
                      </span>
                      <span class="rounded-full border border-obsidian-border px-2.5 py-0.5 text-xs text-white/60">
                        剩余 {{ preview.stock }} 件
                      </span>
                    </div>
                    <div class="mt-3 h-2 w-full overflow-hidden rounded-full bg-obsidian-border">
                      <div
                        class="h-full rounded-full bg-gradient-to-r from-magma to-amber-300 transition-all duration-300"
                        :style="{ width: `${preview.stock > 0 ? Math.min(100, preview.stock) : 0}%` }"
                      ></div>
                    </div>
                  </div>
                </div>
              </div>
            </ParallaxCard>

            <!-- 发布须知 -->
            <Card class="border-obsidian-border/70 bg-obsidian-card/80">
              <CardHeader class="pb-3">
                <CardTitle class="flex items-center gap-2 text-base">
                  <Sparkles class="h-4 w-4 text-magma" />
                  发布须知
                </CardTitle>
              </CardHeader>
              <CardContent class="space-y-3 text-sm">
                <div class="flex items-start gap-3">
                  <span class="mt-1 h-2 w-2 shrink-0 rounded-full bg-magma"></span>
                  <div>
                    <p class="font-medium text-white">即时上架</p>
                    <p class="text-xs text-white/50">发布后商品立即展示在抢购大厅</p>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <span class="mt-1 h-2 w-2 shrink-0 rounded-full bg-emerald-400"></span>
                  <div>
                    <p class="font-medium text-white">定时开抢</p>
                    <p class="text-xs text-white/50">到达开抢时间后用户可参与秒杀</p>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <span class="mt-1 h-2 w-2 shrink-0 rounded-full bg-white/60"></span>
                  <div>
                    <p class="font-medium text-white">随时编辑</p>
                    <p class="text-xs text-white/50">可随时修改商品信息和库存数量</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        <!-- 我发布的商品 -->
        <div class="mt-4">
          <div class="mb-4 flex items-center justify-between">
            <div class="space-y-1">
              <p class="flex items-center gap-2 text-sm uppercase tracking-[0.3em] text-magma">
                <Package class="h-4 w-4" />
                My Products
              </p>
              <h2 class="text-xl font-semibold">我发布的商品</h2>
            </div>
            <Button
              variant="outline"
              size="sm"
              class="gap-2 border-obsidian-border text-white/70 hover:border-magma hover:text-magma"
              @click="productStore.fetchMyProducts()"
            >
              <RotateCcw class="h-3.5 w-3.5" />
              刷新
            </Button>
          </div>

          <!-- 骨架屏 -->
          <div v-if="productStore.loading" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <div v-for="i in 3" :key="i" class="h-36 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/70"></div>
          </div>

          <!-- 空状态 -->
          <div v-else-if="productStore.myItems.length === 0" class="rounded-2xl border border-dashed border-obsidian-border/70 bg-obsidian-card/50 p-8 text-center">
            <Package class="mx-auto h-12 w-12 text-white/20" />
            <p class="mt-3 text-white/50">暂无商品，先发布一条吧</p>
          </div>

          <!-- 商品列表 -->
          <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card
              v-for="p in productStore.myItems"
              :key="p.id"
              class="group overflow-hidden border-obsidian-border/70 bg-obsidian-card/80 transition hover:border-magma/30"
              :class="{ 'ring-2 ring-magma/50': editingId === p.id }"
            >
              <div class="flex gap-4 p-4">
                <div class="relative shrink-0 overflow-hidden rounded-xl">
                  <img :src="resolveAssetUrl(p.image) || placeholderImg" alt="" class="h-20 w-20 object-cover transition group-hover:scale-105" />
                  <div class="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent"></div>
                </div>
                <div class="flex flex-1 flex-col justify-between">
                  <div>
                    <div class="flex items-start justify-between gap-2">
                      <h3 class="line-clamp-1 font-semibold text-white">{{ p.name }}</h3>
                      <Badge
                        v-if="new Date(p.start_time) <= new Date()"
                        variant="outline"
                        class="shrink-0 border-emerald-500/40 bg-emerald-500/10 text-xs text-emerald-300"
                      >
                        抢购中
                      </Badge>
                      <Badge
                        v-else
                        variant="outline"
                        class="shrink-0 border-purple-500/40 bg-purple-500/10 text-xs text-purple-300"
                      >
                        即将开始
                      </Badge>
                    </div>
                    <div class="mt-1.5 flex items-center gap-3 text-sm text-white/60">
                      <span class="text-magma font-semibold">{{ formatPrice(p.price) }}</span>
                      <span class="text-white/40">·</span>
                      <span>库存 {{ p.stock }}</span>
                    </div>
                  </div>
                  <div class="mt-2 flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      class="h-7 flex-1 gap-1.5 border-obsidian-border text-xs text-white/70 hover:border-magma hover:text-magma"
                      @click="startEdit(p)"
                    >
                      <Edit3 class="h-3 w-3" />
                      编辑
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      class="h-7 flex-1 gap-1.5 border-obsidian-border text-xs text-white/70 hover:border-red-500/50 hover:text-red-400"
                      @click="deleteProduct(p.id)"
                    >
                      <Trash2 class="h-3 w-3" />
                      下架
                    </Button>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
