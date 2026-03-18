<script setup lang="ts">
import { onMounted, computed, ref } from "vue"
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
const showEnded = ref(false)

const isLoggedIn = computed(() => !!userStore.accessToken)
const displayName = computed(() => userStore.profile?.username || "尊贵用户")
const heroAvatar = computed(() => {
  const name = userStore.profile?.username || "guest"
  return resolveAssetUrl(userStore.profile?.avatar) || `https://api.dicebear.com/7.x/shapes/svg?seed=${encodeURIComponent(name)}`
})

onMounted(() => {
  productStore.fetchProducts(1, pageSize)
})

const stockPercent = (stock?: number) => {
  const n = Number(stock)
  if (!Number.isFinite(n) || n < 0) return 0
  return Math.min(100, n)
}

const productCover = (src?: string) => resolveAssetUrl(src) || placeholderImg

const liveProducts = computed(() => productStore.items.filter((p) => {
  if (!p.id) return false
  const now = new Date()
  if (new Date(p.start_time) > now) return false
  if (p.stock <= 0) return false
  if (p.end_time && new Date(p.end_time) < now) return false
  return true
}))

const upcomingProducts = computed(() => productStore.items.filter((p) => {
  if (!p.id) return false
  return new Date(p.start_time) > new Date()
}))

const endedProducts = computed(() => productStore.items.filter((p) => {
  if (!p.id) return false
  const now = new Date()
  if (new Date(p.start_time) > now) return false
  if (p.end_time && new Date(p.end_time) < now) return true
  if (p.stock <= 0) return true
  return false
}))

const sections = computed(() => {
  const s: Array<{ key: string; label: string; title: string; items: Product[] }> = []
  if (liveProducts.value.length > 0) s.push({ key: "live", label: "Live Now", title: "正在抢购", items: liveProducts.value })
  if (upcomingProducts.value.length > 0) s.push({ key: "upcoming", label: "Coming Soon", title: "即将开始", items: upcomingProducts.value })
  if (endedProducts.value.length > 0) s.push({ key: "ended", label: "Ended", title: "已结束", items: endedProducts.value })
  return s
})

const heroProduct = computed(() => {
  return liveProducts.value.find((p) => p.id) || upcomingProducts.value.find((p) => p.id) || productStore.items.find((p) => p.id) || null
})

const loadMore = () => {
  const currentPage = Math.ceil(productStore.items.length / pageSize) || 1
  if (productStore.items.length >= productStore.total) return
  productStore.fetchProducts(currentPage + 1, pageSize, true)
}
</script>

<template>
  <MainLayout>
    <!-- Hero -->
    <section class="border-b border-[#1C1C1C]/10 bg-white">
      <div class="mx-auto flex max-w-6xl flex-col gap-10 px-6 py-16 md:flex-row md:items-center md:py-24">
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
            <img :src="productCover(heroProduct.image)" :alt="heroProduct.name" class="h-[360px] w-full object-cover" />
            <div class="border-t border-[#1C1C1C]/10 p-6">
              <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">当前抢购</p>
              <h3 class="mt-1 font-serif text-2xl tracking-tight">{{ heroProduct.name }}</h3>
              <p class="mt-2 text-lg">{{ formatPrice(heroProduct.price) }}</p>
            </div>
          </div>
          <div v-else class="border border-[#1C1C1C]/10 p-6 text-[#1C1C1C]/40">
            暂无商品，去发布新品或刷新列表。
          </div>
        </div>
      </div>
    </section>

    <!-- Product Sections -->
    <section class="mx-auto max-w-6xl px-6 py-16 md:py-24">
      <div v-if="productStore.loading">
        <div class="mb-8 flex items-end justify-between border-b border-[#1C1C1C]/10 pb-4">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Seckill Hall</p>
            <h2 class="font-serif text-2xl tracking-tight md:text-3xl">抢购大厅</h2>
          </div>
        </div>
        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <div v-for="i in 6" :key="i" class="animate-pulse border border-[#1C1C1C]/10 p-4">
            <div class="h-44 w-full bg-[#1C1C1C]/5"></div>
            <div class="mt-4 h-4 w-3/4 bg-[#1C1C1C]/5"></div>
            <div class="mt-2 h-4 w-1/2 bg-[#1C1C1C]/5"></div>
          </div>
        </div>
      </div>

      <div v-else-if="productStore.items.length === 0" class="border border-[#1C1C1C]/10 p-8 text-center text-[#1C1C1C]/40">
        暂无商品，快去发布第一双吧。
      </div>

      <div v-else class="space-y-16">
        <div v-for="section in sections" :key="section.key">
          <div class="mb-8 flex items-end justify-between border-b border-[#1C1C1C]/10 pb-4">
            <div>
              <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">{{ section.label }}</p>
              <h2 class="font-serif text-2xl tracking-tight md:text-3xl">{{ section.title }}</h2>
            </div>
            <div class="flex items-center gap-3">
              <p v-if="section.key === 'live'" class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">限量发售 · 先到先得</p>
              <button
                v-if="section.key === 'ended'"
                class="hover-underline text-xs text-[#1C1C1C]/40 transition-colors hover:text-[#1C1C1C]"
                @click="showEnded = !showEnded"
              >
                {{ showEnded ? "收起" : `展开 ${endedProducts.length} 件` }}
              </button>
            </div>
          </div>

          <div v-if="section.key === 'ended' && !showEnded" class="border border-dashed border-[#1C1C1C]/10 py-6 text-center text-sm text-[#1C1C1C]/40">
            {{ endedProducts.length }} 件商品已结束
          </div>

          <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            <Card
              v-for="item in section.items"
              :key="item.id"
              class="group overflow-hidden hover:border-[#1C1C1C]/30"
            >
              <RouterLink :to="`/product/${item.id}`">
                <div class="relative overflow-hidden">
                  <img
                    :src="productCover(item.image)"
                    :alt="item.name"
                    class="h-52 w-full object-cover"
                  />
                  <div class="absolute left-3 top-3 border border-[#1C1C1C]/20 bg-white/90 px-3 py-1 text-xs text-[#1C1C1C]/60">
                    {{ new Date(item.start_time).toLocaleString() }}
                  </div>
                </div>
                <CardHeader class="space-y-2">
                  <CardTitle class="flex items-center justify-between text-lg">
                    <span class="font-serif tracking-tight">{{ item.name }}</span>
                    <span class="text-base">{{ formatPrice(item.price) }}</span>
                  </CardTitle>
                  <CardDescription class="flex items-center justify-between text-[#1C1C1C]/60">
                    <span>库存 {{ item.stock }}</span>
                    <span v-if="item.end_time && new Date(item.end_time!) < new Date()" class="border border-[#1C1C1C]/10 px-3 py-1 text-xs text-[#1C1C1C]/40">已结束</span>
                    <CountdownPill v-else :start-time="item.start_time" />
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Progress :model-value="stockPercent(item.stock)" />
                </CardContent>
              </RouterLink>
            </Card>
          </div>
        </div>
      </div>

      <div v-if="productStore.items.length < productStore.total" class="mt-8 flex justify-center">
        <MagmaButton :loading="productStore.loading" @click="loadMore">加载更多</MagmaButton>
      </div>
    </section>
  </MainLayout>
</template>
