<script setup lang="ts">
import { reactive, computed, ref, onMounted } from "vue"
import { useRouter } from "vue-router"
import {
  Upload, Eye, Edit3, Trash2, RotateCcw, CheckCircle, AlertCircle,
} from "lucide-vue-next"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import api, { resolveAssetUrl, uploadImage } from "@/lib/api"
import { toast } from "vue-sonner"
import { formatPrice } from "@/lib/utils"
import { useProductStore } from "@/stores/productStore"

const router = useRouter()
const productStore = useProductStore()

const form = reactive({ name: "", price: "", stock: "", start_time: "", end_time: "", image: "" })
const editingId = ref<number | null>(null)
const loading = reactive({ submitting: false })
const uploading = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const placeholderImg = "https://dummyimage.com/900x600/F9F8F6/1C1C1C&text=SneakerFlash"
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
  if (!form.name || !form.price || !form.stock || !form.start_time) { toast.error("请填写完整信息"); return }
  const parsedTime = new Date(form.start_time)
  if (Number.isNaN(parsedTime.getTime())) { toast.error("开抢时间不合法"); return }
  loading.submitting = true
  try {
    const payload: Record<string, unknown> = {
      name: form.name, price: Number(form.price), stock: Number(form.stock),
      start_time: parsedTime.toISOString(), image: form.image,
    }
    if (form.end_time) {
      const endParsed = new Date(form.end_time)
      if (Number.isNaN(endParsed.getTime())) { toast.error("结束时间不合法"); loading.submitting = false; return }
      if (endParsed <= parsedTime) { toast.error("结束时间必须晚于开抢时间"); loading.submitting = false; return }
      payload.end_time = endParsed.toISOString()
    }
    if (editingId.value) { await api.put(`/products/${editingId.value}`, payload); toast.success("商品更新成功") }
    else { await api.post("/products", payload); toast.success("商品发布成功，库存已预热") }
    productStore.fetchProducts(1, 12); productStore.fetchMyProducts()
    if (!editingId.value) router.push("/")
  } catch (err: any) { toast.error(err?.message || "发布失败") }
  finally { loading.submitting = false }
}

const onImageSelected = async (event: Event) => {
  const target = event.target as HTMLInputElement | null; const file = target?.files?.[0]; if (!file) return
  uploading.value = true
  try { const url = await uploadImage(file); form.image = url; toast.success("图片上传成功") }
  catch (err: any) { toast.error(err?.message || "上传失败") }
  finally { uploading.value = false; if (target) target.value = "" }
}

const startEdit = (p: any) => {
  editingId.value = p.id; form.name = p.name; form.price = String(p.price); form.stock = String(p.stock)
  form.start_time = p.start_time?.slice(0, 16) || ""; form.end_time = p.end_time?.slice(0, 16) || ""; form.image = p.image || ""
}
const resetForm = () => { editingId.value = null; form.name = ""; form.price = ""; form.stock = ""; form.start_time = ""; form.end_time = ""; form.image = "" }
onMounted(() => { productStore.fetchMyProducts() })
const deleteProduct = async (id: number) => {
  try { await productStore.deleteProduct(id); toast.success("已删除"); if (editingId.value === id) resetForm() }
  catch (err: any) { toast.error(err?.message || "删除失败") }
}
</script>

<template>
  <MainLayout>
    <section class="mx-auto max-w-6xl px-6 py-16 md:py-24">
      <div class="flex flex-col gap-8">
        <!-- 标题 -->
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Publish</p>
            <h1 class="font-serif text-3xl tracking-tight">{{ editingId ? "编辑商品" : "发布商品" }}</h1>
          </div>
          <Badge v-if="editingId" variant="outline">编辑模式</Badge>
          <Badge v-else variant="outline">新品发布</Badge>
        </div>

        <!-- 四宫格 -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Published</p>
            <div class="mt-2 text-2xl">{{ stats.total }} <span class="text-xs text-[#1C1C1C]/40">件</span></div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Total Stock</p>
            <div class="mt-2 text-2xl">{{ stats.totalStock }} <span class="text-xs text-[#1C1C1C]/40">件</span></div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Active</p>
            <div class="mt-2 text-2xl">{{ stats.active }}</div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Upcoming</p>
            <div class="mt-2 text-2xl">{{ stats.upcoming }}</div>
          </div>
        </div>

        <!-- 表单 + 预览 -->
        <div class="grid gap-8 lg:grid-cols-[1.1fr_0.9fr]">
          <!-- 表单 -->
          <Card>
            <CardHeader class="space-y-3 pb-4">
              <div class="flex items-center justify-between">
                <CardTitle class="font-serif text-xl tracking-tight">商品信息</CardTitle>
                <div class="flex items-center gap-2 text-xs text-[#1C1C1C]/40">
                  <span>完成度</span>
                  <div class="h-1 w-16 overflow-hidden bg-[#1C1C1C]/10">
                    <div class="h-full bg-[#1C1C1C] transition-all duration-300" :style="{ width: `${formProgress}%` }"></div>
                  </div>
                  <span>{{ formProgress }}%</span>
                </div>
              </div>
              <CardDescription class="text-[#1C1C1C]/40">填写商品信息，发布后即可参与秒杀活动</CardDescription>
            </CardHeader>
            <CardContent class="space-y-5">
              <div class="space-y-2">
                <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">商品名称 <span class="text-[#1C1C1C]/60">*</span></label>
                <Input v-model="form.name" placeholder="如：Air Zoom 1" />
              </div>
              <div class="grid gap-4 md:grid-cols-3">
                <div class="space-y-2">
                  <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">价格（元） <span class="text-[#1C1C1C]/60">*</span></label>
                  <Input v-model="form.price" type="number" min="0" step="0.01" placeholder="999.00" />
                </div>
                <div class="space-y-2">
                  <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">库存 <span class="text-[#1C1C1C]/60">*</span></label>
                  <Input v-model="form.stock" type="number" min="0" step="1" placeholder="100" />
                </div>
                <div class="space-y-2">
                  <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">开抢时间 <span class="text-[#1C1C1C]/60">*</span></label>
                  <Input v-model="form.start_time" type="datetime-local" />
                </div>
              </div>
              <div class="space-y-2">
                <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">结束时间（可选）</label>
                <Input v-model="form.end_time" type="datetime-local" />
                <p class="text-xs text-[#1C1C1C]/30">不设置则表示活动永不过期</p>
              </div>
              <div class="space-y-3">
                <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">封面图（可选）</label>
                <Input v-model="form.image" type="url" placeholder="https://..." />
                <div class="flex flex-wrap items-center gap-3">
                  <Button variant="outline" size="sm" :disabled="uploading" @click="fileInput?.click()">
                    <Upload class="h-3.5 w-3.5" />
                    {{ uploading ? "上传中..." : "上传图片" }}
                  </Button>
                  <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onImageSelected" />
                  <span class="text-xs text-[#1C1C1C]/30">可填写外链，或上传图片自动生成链接</span>
                </div>
              </div>
              <div class="flex flex-wrap gap-3 border-t border-[#1C1C1C]/10 pt-5">
                <MagmaButton class="flex-1 justify-center gap-2" :loading="loading.submitting" @click="submit">
                  <CheckCircle v-if="!loading.submitting" class="h-4 w-4" />
                  {{ editingId ? "保存修改" : "发布商品并预热" }}
                </MagmaButton>
                <Button v-if="editingId" variant="outline" class="flex-1 gap-2" @click="resetForm">
                  <RotateCcw class="h-4 w-4" />
                  取消编辑
                </Button>
              </div>
              <div class="flex items-start gap-2 border border-[#1C1C1C]/10 p-3 text-xs text-[#1C1C1C]/40">
                <AlertCircle class="mt-0.5 h-3.5 w-3.5 shrink-0" />
                <span>发布成功后商品将立即上架，用户可在开抢时间参与秒杀。</span>
              </div>
            </CardContent>
          </Card>

          <!-- 预览 + 须知 -->
          <div class="flex flex-col gap-6">
            <div class="border border-[#1C1C1C]/10 bg-white">
              <div class="border-b border-[#1C1C1C]/10 p-4">
                <div class="flex items-center gap-2 font-serif text-lg tracking-tight">
                  <Eye class="h-5 w-5 text-[#1C1C1C]/40" />
                  实时预览
                </div>
                <p class="mt-1 text-xs text-[#1C1C1C]/30">核对展示信息，避免秒杀页出现空图/错误数据</p>
              </div>
              <div class="p-4">
                <div class="border border-[#1C1C1C]/10">
                  <img :src="preview.image" alt="" class="h-52 w-full object-cover" />
                  <div class="border-t border-[#1C1C1C]/10 p-4">
                    <p class="text-xs text-[#1C1C1C]/30">{{ preview.start }}</p>
                    <h3 class="mt-1 font-serif text-xl tracking-tight">{{ preview.name }}</h3>
                    <p class="mt-1.5 text-lg">{{ formatPrice(preview.price) }}</p>
                  </div>
                </div>
              </div>
              <div class="border-t border-[#1C1C1C]/10 p-4">
                <div class="flex items-center justify-between text-sm">
                  <span class="text-[#1C1C1C]/60">库存进度</span>
                  <span class="text-xs text-[#1C1C1C]/40">剩余 {{ preview.stock }} 件</span>
                </div>
                <div class="mt-3 h-1 w-full bg-[#1C1C1C]/10">
                  <div class="h-full bg-[#1C1C1C] transition-all duration-300" :style="{ width: `${preview.stock > 0 ? Math.min(100, preview.stock) : 0}%` }"></div>
                </div>
              </div>
            </div>

            <Card>
              <CardHeader class="pb-3">
                <CardTitle class="font-serif text-base tracking-tight">发布须知</CardTitle>
              </CardHeader>
              <CardContent class="space-y-3 text-sm">
                <div class="flex items-start gap-3">
                  <span class="mt-1.5 h-1 w-1 bg-[#1C1C1C]/70"></span>
                  <div>
                    <p class="font-medium">即时上架</p>
                    <p class="text-xs text-[#1C1C1C]/40">发布后商品立即展示在抢购大厅</p>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <span class="mt-1.5 h-1 w-1 bg-[#1C1C1C]"></span>
                  <div>
                    <p class="font-medium">定时开抢</p>
                    <p class="text-xs text-[#1C1C1C]/40">到达开抢时间后用户可参与秒杀</p>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <span class="mt-1.5 h-1 w-1 bg-[#1C1C1C]/40"></span>
                  <div>
                    <p class="font-medium">随时编辑</p>
                    <p class="text-xs text-[#1C1C1C]/40">可随时修改商品信息和库存数量</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        <!-- 我发布的商品 -->
        <div class="mt-4">
          <div class="mb-4 flex items-center justify-between border-b border-[#1C1C1C]/10 pb-4">
            <div>
              <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">My Products</p>
              <h2 class="font-serif text-xl tracking-tight">我发布的商品</h2>
            </div>
            <Button variant="outline" size="sm" class="gap-2" @click="productStore.fetchMyProducts()">
              <RotateCcw class="h-3.5 w-3.5" />
              刷新
            </Button>
          </div>

          <div v-if="productStore.loading" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <div v-for="i in 3" :key="i" class="h-36 animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>
          </div>

          <div v-else-if="productStore.myItems.length === 0" class="border border-dashed border-[#1C1C1C]/10 p-8 text-center text-[#1C1C1C]/40">
            暂无商品，先发布一条吧
          </div>

          <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card
              v-for="p in productStore.myItems"
              :key="p.id"
              class="group overflow-hidden transition hover:border-[#1C1C1C]/30"
              :class="{ 'border-[#1C1C1C]': editingId === p.id }"
            >
              <div class="flex gap-4 p-4">
                <div class="shrink-0 overflow-hidden">
                  <img :src="resolveAssetUrl(p.image) || placeholderImg" alt="" class="h-20 w-20 object-cover transition group-hover:scale-105" />
                </div>
                <div class="flex flex-1 flex-col justify-between">
                  <div>
                    <div class="flex items-start justify-between gap-2">
                      <h3 class="line-clamp-1 font-medium">{{ p.name }}</h3>
                      <Badge v-if="new Date(p.start_time) <= new Date()" variant="outline" class="shrink-0 text-xs">抢购中</Badge>
                      <Badge v-else variant="outline" class="shrink-0 text-xs">即将开始</Badge>
                    </div>
                    <div class="mt-1.5 flex items-center gap-3 text-sm text-[#1C1C1C]/40">
                      <span>{{ formatPrice(p.price) }}</span>
                      <span>·</span>
                      <span>库存 {{ p.stock }}</span>
                    </div>
                  </div>
                  <div class="mt-2 flex items-center gap-2">
                    <Button variant="outline" size="sm" class="h-7 flex-1 gap-1.5 text-xs" @click="startEdit(p)">
                      <Edit3 class="h-3 w-3" />
                      编辑
                    </Button>
                    <Button variant="outline" size="sm" class="h-7 flex-1 gap-1.5 text-xs hover:border-[#1C1C1C] hover:text-[#1C1C1C]" @click="deleteProduct(p.id)">
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
