<script setup lang="ts">
import { onMounted, computed } from "vue"
import { RouterLink } from "vue-router"
import { useProductStore } from "@/stores/productStore"
import { useUserStore } from "@/stores/userStore"
import MainLayout from "@/layout/MainLayout.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import CountdownPill from "@/components/CountdownPill.vue"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { formatPrice } from "@/lib/utils"
import { resolveAssetUrl } from "@/lib/api"
import type { Product } from "@/types/product"

const placeholderImg = "https://dummyimage.com/900x600/F9F8F6/1C1C1C&text=SneakerFlash"

const productStore = useProductStore()
const userStore = useUserStore()
const pageSize = 12
const currentPage = computed(() => Math.ceil(productStore.items.length / pageSize) || 1)
const isLoggedIn = computed(() => !!userStore.accessToken)
const displayName = computed(() => userStore.profile?.username || "尊贵用户")
const heroAvatar = computed(() => {
  const name = userStore.profile?.username || "guest"
  return resolveAssetUrl(userStore.profile?.avatar) || `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(name)}`
})

onMounted(() => {
  productStore.fetchProducts(1, pageSize)
})

const heroProduct = computed(() => {
  const first = Array.isArray(productStore.items) ? productStore.items[0] : undefined
  return first && first.id ? first : null
})
const skeletons = computed(() => Array.from({ length: 6 }))
const placeholders = computed(() => Array.from({ length: 3 }))

const stockPercent = (stock?: number) => {
  const n = Number(stock)
  if (!Number.isFinite(n) || n < 0) return 0
  return Math.min(100, n)
}

const safeProduct = (item: Partial<Product> | undefined): Product => {
  if (item && item.id) {
    return item as Product
  }
  return {
    id: 0,
    user_id: 0,
    name: "未命名",
    price: 0,
    stock: 0,
    start_time: new Date().toISOString(),
    image: "/placeholder.svg",
  }
}

const productCover = (src?: string) => resolveAssetUrl(src) || placeholderImg

const loadMore = () => {
  const nextPage = currentPage.value + 1
  if (productStore.items.length >= productStore.total) return
  productStore.fetchProducts(nextPage, pageSize, true)
}
</script>

<template>
  <MainLayout>
    <!-- Hero -->
    <section class="border-b border-[#1C1C1C]/10 bg-white">
      <div class="mx-auto flex max-w-6xl flex-col gap-10 px-6 py-16 md:py-24 md:flex-row md:items-center">
        <div class="flex-1 space-y-6">
          <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Flash Sale</p>
          <h1 class="font-serif text-4xl leading-tight tracking-tight md:text-6xl">
            以速度，<br /><em>抢下限量。</em>
          </h1>
          <p class="text-[#1C1C1C]/60">
            限量球鞋，先到先得。实时库存更新，抢到就是赚到。
          </p>
          <div class="flex flex-wrap gap-3">
            <template v-if="!isLoggedIn">
              <RouterLink v-slot="{ navigate }" to="/login" custom>
                <MagmaButton @click="navigate">立即登录</MagmaButton>
              </RouterLink>
              <RouterLink v-slot="{ navigate }" to="/register" custom>
                <button
                  class="border border-[#1C1C1C]/20 px-6 py-3 text-sm tracking-wide transition-colors hover:border-[#1C1C1C]"
                  @click="navigate"
                >
                  先去注册
                </button>
              </RouterLink>
            </template>
            <template v-else>
              <div class="flex items-center gap-3 text-sm text-[#1C1C1C]/60">
                <img :src="heroAvatar" :alt="`${displayName} avatar`" class="h-8 w-8 border border-[#1C1C1C]/10 object-cover" />
                <span>欢迎回来，{{ displayName }}</span>
              </div>
              <RouterLink v-slot="{ navigate }" to="/orders" custom>
                <MagmaButton @click="navigate">前往订单</MagmaButton>
              </RouterLink>
              <RouterLink v-slot="{ navigate }" to="/products/publish" custom>
                <button
                  class="border border-[#1C1C1C]/20 px-6 py-3 text-sm tracking-wide transition-colors hover:border-[#1C1C1C]"
                  @click="navigate"
                >
                  发布新品
                </button>
              </RouterLink>
            </template>
          </div>
          <div class="flex flex-wrap gap-4 text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">
            <span>正品保障</span>
            <span class="text-[#1C1C1C]/20">·</span>
            <span>极速发货</span>
          </div>
        </div>
        <div class="flex-1">
          <div v-if="heroProduct" class="border border-[#1C1C1C]/10 bg-white">
            <img :src="productCover(heroProduct?.image)" :alt="heroProduct?.name || 'Hero Product'" class="h-[360px] w-full object-cover" />
            <div class="border-t border-[#1C1C1C]/10 p-6">
              <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">当前抢购</p>
              <h3 class="mt-1 font-serif text-2xl tracking-tight">{{ heroProduct?.name }}</h3>
              <p class="mt-2 text-lg">{{ formatPrice(heroProduct?.price || 0) }}</p>
            </div>
          </div>
          <div v-else class="border border-[#1C1C1C]/10 p-6 text-[#1C1C1C]/40">
            暂无商品，去发布新品或刷新列表。
          </div>
        </div>
      </div>
    </section>

    <!-- Seckill Hall -->
    <section class="mx-auto max-w-6xl px-6 py-16 md:py-24">
      <div class="mb-8 flex items-end justify-between border-b border-[#1C1C1C]/10 pb-4">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Seckill Hall</p>
          <h2 class="font-serif text-2xl tracking-tight md:text-3xl">抢购大厅</h2>
        </div>
        <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">限量发售 · 先到先得</p>
      </div>

      <div v-if="productStore.loading" class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <div v-for="(_, idx) in skeletons" :key="idx" class="animate-pulse border border-[#1C1C1C]/10 p-4">
          <div class="h-44 w-full bg-[#1C1C1C]/5"></div>
          <div class="mt-4 h-4 w-3/4 bg-[#1C1C1C]/5"></div>
          <div class="mt-2 h-4 w-1/2 bg-[#1C1C1C]/5"></div>
        </div>
      </div>

      <div v-else-if="productStore.items.length === 0" class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <Card v-for="(_, idx) in placeholders" :key="idx" class="overflow-hidden">
          <div class="h-52 w-full bg-[#1C1C1C]/5"></div>
          <CardHeader class="space-y-2">
            <CardTitle class="flex items-center justify-between text-lg text-[#1C1C1C]/40">
              <span>即将上架</span>
              <span class="text-base">¥0.00</span>
            </CardTitle>
            <CardDescription class="flex items-center justify-between text-[#1C1C1C]/40">
              <span>库存 --</span>
              <span class="border border-[#1C1C1C]/10 px-3 py-1 text-xs">未开始</span>
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div class="h-1 w-full bg-[#1C1C1C]/5"></div>
          </CardContent>
        </Card>
      </div>

      <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <Card
          v-for="raw in productStore.items"
          :key="safeProduct(raw).id"
          class="group overflow-hidden hover:border-[#1C1C1C]/30"
        >
          <RouterLink :to="`/product/${safeProduct(raw).id}`">
            <div class="relative overflow-hidden">
              <img
                :src="productCover(safeProduct(raw).image)"
                :alt="safeProduct(raw).name"
                class="h-52 w-full object-cover"
              />
              <div class="absolute left-3 top-3 border border-[#1C1C1C]/20 bg-white/90 px-3 py-1 text-xs text-[#1C1C1C]/60">
                {{ new Date(safeProduct(raw).start_time).toLocaleString() }}
              </div>
            </div>
            <CardHeader class="space-y-2">
              <CardTitle class="flex items-center justify-between text-lg">
                <span class="font-serif tracking-tight">{{ safeProduct(raw).name }}</span>
                <span class="text-base">{{ formatPrice(safeProduct(raw).price) }}</span>
              </CardTitle>
              <CardDescription class="flex items-center justify-between text-[#1C1C1C]/60">
                <span>库存 {{ safeProduct(raw).stock }}</span>
                <span v-if="safeProduct(raw).end_time && new Date(safeProduct(raw).end_time!) < new Date()" class="border border-[#1C1C1C]/10 px-3 py-1 text-xs text-[#1C1C1C]/40">已结束</span>
                <CountdownPill v-else :start-time="safeProduct(raw).start_time" />
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Progress :model-value="stockPercent(safeProduct(raw).stock)" />
            </CardContent>
          </RouterLink>
        </Card>
      </div>

      <div v-if="!productStore.loading && productStore.items.length === 0" class="mt-8 border border-[#1C1C1C]/10 p-8 text-center text-[#1C1C1C]/40">
        暂无商品，快去发布第一双吧。
      </div>

      <div v-if="productStore.items.length < productStore.total" class="mt-8 flex justify-center">
        <MagmaButton :loading="productStore.loading" @click="loadMore">加载更多</MagmaButton>
      </div>
    </section>
  </MainLayout>
</template>
