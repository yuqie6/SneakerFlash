<script setup lang="ts">
import { onMounted, computed } from "vue"
import { RouterLink } from "vue-router"
import { Flame, Sparkles } from "lucide-vue-next"
import { useProductStore } from "@/stores/productStore"
import MainLayout from "@/layout/MainLayout.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import CountdownPill from "@/components/CountdownPill.vue"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { formatPrice } from "@/lib/utils"

const productStore = useProductStore()
const pageSize = 12
const currentPage = computed(() => Math.ceil(productStore.items.length / pageSize) || 1)

onMounted(() => {
  productStore.fetchProducts(1, pageSize)
})

const heroProduct = computed(() => productStore.items[0])
const skeletons = computed(() => Array.from({ length: 6 }))

const stockPercent = (stock: number) => Math.max(0, Math.min(100, stock))

const loadMore = () => {
  const nextPage = currentPage.value + 1
  if (productStore.items.length >= productStore.total) return
  productStore.fetchProducts(nextPage, pageSize, true)
}
</script>

<template>
  <MainLayout>
    <section class="relative overflow-hidden border-b border-obsidian-border/60 bg-gradient-to-br from-black via-obsidian-card to-black">
      <div class="pointer-events-none absolute inset-0 opacity-70 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-10 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-0 right-0 h-80 w-80 rounded-full bg-[#ea580c66] blur-3xl"></div>
      </div>
      <div class="relative mx-auto flex max-w-6xl flex-col gap-10 px-6 py-16 md:flex-row md:items-center">
        <div class="flex-1 space-y-6">
          <p class="flex items-center gap-2 text-sm uppercase tracking-[0.35em] text-magma">
            <Flame class="h-4 w-4" />
            Midnight Magma Drop
          </p>
          <h1 class="text-4xl font-bold leading-tight md:text-5xl">
            以熔岩速度，抢下你的下一双限量鞋。
          </h1>
          <p class="text-white/70">
            实时库存，毫秒级下单反馈。黑曜石质感搭配熔岩流光，专为高压场景优化。
          </p>
          <div class="flex flex-wrap gap-3">
            <RouterLink to="/login">
              <MagmaButton>立即登录</MagmaButton>
            </RouterLink>
            <RouterLink to="/register">
              <button class="rounded-full border border-obsidian-border px-6 py-3 text-sm text-white transition hover:border-magma hover:text-magma">
                先去注册
              </button>
            </RouterLink>
          </div>
          <div class="flex flex-wrap gap-4 text-sm text-white/70">
            <div class="flex items-center gap-2 rounded-full border border-obsidian-border/70 px-3 py-2">
              <Sparkles class="h-4 w-4 text-magma" />
              物理动效
            </div>
            <div class="flex items-center gap-2 rounded-full border border-obsidian-border/70 px-3 py-2">
              <Sparkles class="h-4 w-4 text-magma" />
              毫秒反馈
            </div>
          </div>
        </div>
        <div class="flex-1">
          <ParallaxCard v-if="heroProduct" class="glass">
            <div class="relative h-full w-full overflow-hidden rounded-2xl">
              <img :src="heroProduct?.image || '/placeholder.svg'" alt="hero" class="h-full w-full object-cover" />
              <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-black/20 to-transparent"></div>
              <div class="absolute bottom-0 left-0 right-0 p-6">
                <p class="text-sm text-white/70">当前抢购</p>
                <h3 class="text-2xl font-semibold">{{ heroProduct?.name }}</h3>
                <p class="mt-2 text-magma text-lg">{{ formatPrice(heroProduct?.price || 0) }}</p>
              </div>
            </div>
          </ParallaxCard>
        </div>
      </div>
    </section>

    <section class="relative mx-auto max-w-6xl px-6 py-12">
      <div class="mb-6 flex items-center justify-between">
        <div>
          <p class="text-sm uppercase tracking-[0.3em] text-magma">Seckill Hall</p>
          <h2 class="text-2xl font-semibold">抢购大厅</h2>
        </div>
        <p class="text-sm text-white/60">实时库存 · 秒杀预备</p>
      </div>

      <div v-if="productStore.loading" class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <div v-for="(_, idx) in skeletons" :key="idx" class="animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card p-4">
          <div class="h-44 w-full rounded-xl bg-white/5"></div>
          <div class="mt-4 h-4 w-3/4 rounded bg-white/5"></div>
          <div class="mt-2 h-4 w-1/2 rounded bg-white/5"></div>
        </div>
      </div>

      <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <Card v-for="item in productStore.items" :key="item.id" class="overflow-hidden border-obsidian-border/70 bg-obsidian-card/80">
          <RouterLink :to="`/product/${item.id}`">
            <div class="relative overflow-hidden">
              <img :src="item.image || '/placeholder.svg'" alt="" class="h-52 w-full object-cover transition duration-500 hover:scale-105" />
              <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-transparent to-transparent"></div>
              <div class="absolute left-3 top-3 flex items-center gap-2 rounded-full bg-white/10 px-3 py-1 text-xs text-white">
                <span class="h-2 w-2 rounded-full bg-magma animate-pulse-fast"></span>
                {{ new Date(item.start_time).toLocaleString() }}
              </div>
            </div>
            <CardHeader class="space-y-2">
              <CardTitle class="flex items-center justify-between text-lg">
                <span>{{ item.name }}</span>
                <span class="text-magma text-base font-semibold">{{ formatPrice(item.price) }}</span>
              </CardTitle>
              <CardDescription class="flex items-center justify-between text-white/70">
                <span>库存 {{ item.stock }}</span>
                <CountdownPill :start-time="item.start_time" />
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Progress :model-value="stockPercent(item.stock)" class="h-2 bg-obsidian-border" />
            </CardContent>
          </RouterLink>
        </Card>
      </div>

      <div v-if="!productStore.loading && productStore.items.length === 0" class="mt-8 rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-8 text-center text-white/70">
        暂无商品，快去发布第一双吧。
      </div>

      <div v-if="productStore.items.length < productStore.total" class="mt-8 flex justify-center">
        <MagmaButton :loading="productStore.loading" @click="loadMore">加载更多</MagmaButton>
      </div>
    </section>
  </MainLayout>
</template>
