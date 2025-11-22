<script setup lang="ts">
import { computed, onMounted } from "vue"
import { RouterLink, useRouter } from "vue-router"
import { format } from "date-fns"
import MainLayout from "@/layout/MainLayout.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import { useProductStore } from "@/stores/productStore"
import { formatPrice } from "@/lib/utils"

const productStore = useProductStore()
const heroProduct = computed(() => productStore.items[0])
const router = useRouter()

onMounted(() => {
  productStore.fetchProducts(1, 12)
})

const formatStart = (dateStr?: string) => {
  if (!dateStr) return ""
  return format(new Date(dateStr), "MM月dd日 HH:mm")
}

const goHero = () => {
  if (heroProduct.value) {
    router.push(`/product/${heroProduct.value.id}`)
  }
}
</script>

<template>
  <MainLayout>
    <section class="relative overflow-hidden border-b border-obsidian-border/50 bg-gradient-to-br from-black via-obsidian-card to-obsidian-bg">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(249,115,22,0.18),_transparent_40%)]"></div>
      <div class="relative mx-auto grid max-w-6xl grid-cols-1 items-center gap-10 px-6 py-16 md:grid-cols-2">
        <div class="space-y-6">
          <p class="text-sm uppercase tracking-[0.3em] text-magma">Midnight Magma</p>
          <h1 class="text-4xl font-semibold leading-tight md:text-5xl">
            黑金赛道 · 高燃秒杀
          </h1>
          <p class="text-base text-white/70">
            磁吸动效、物理滚动与毫秒级反馈，抢鞋的每一刻都被放大。锁定心仪款式，提前就位。
          </p>
          <div class="flex items-center gap-4">
            <MagmaButton @click="goHero">立即浏览</MagmaButton>
            <p v-if="heroProduct" class="text-sm text-white/60">
              即将开抢 · {{ formatStart(heroProduct.start_time) }}
            </p>
          </div>
        </div>
        <div class="justify-self-center">
          <ParallaxCard class="w-full max-w-md">
            <div class="aspect-square w-full rounded-2xl bg-gradient-to-br from-black via-obsidian-card to-magma/20 p-6">
              <div class="flex h-full flex-col justify-between">
                <div>
                  <p class="text-sm uppercase text-white/60">Featured</p>
                  <h3 class="mt-2 text-2xl font-semibold">
                    {{ heroProduct?.name || "即刻开抢" }}
                  </h3>
                  <p class="mt-1 text-magma text-sm">
                    {{ heroProduct ? formatPrice(heroProduct.price) : "敬请期待" }}
                  </p>
                </div>
                <RouterLink
                  v-if="heroProduct"
                  :to="`/product/${heroProduct.id}`"
                  class="inline-flex items-center justify-center rounded-full bg-magma-gradient px-4 py-2 text-sm font-semibold text-white"
                >
                  查看详情
                </RouterLink>
              </div>
            </div>
          </ParallaxCard>
        </div>
      </div>
    </section>

    <section class="mx-auto max-w-6xl px-6 py-12">
      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-2xl font-semibold">抢购大厅</h2>
        <p class="text-sm text-white/60">实时同步库存，提前做好准备。</p>
      </div>
      <div v-if="productStore.loading" class="py-10 text-center text-white/60">加载中...</div>
      <div v-else-if="productStore.items.length === 0" class="py-10 text-center text-white/60">
        暂无商品，稍后再试
      </div>
      <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <ParallaxCard v-for="product in productStore.items" :key="product.id" class="group">
          <div class="space-y-4">
            <div class="aspect-[4/3] w-full overflow-hidden rounded-xl bg-obsidian-bg">
              <div
                class="flex h-full items-center justify-center bg-gradient-to-br from-obsidian-card to-black/60 text-white/50 transition duration-500 group-hover:scale-[1.02]"
              >
                {{ product.name }}
              </div>
            </div>
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold">{{ product.name }}</h3>
                <span class="text-magma font-semibold">{{ formatPrice(product.price) }}</span>
              </div>
              <div class="flex items-center justify-between text-xs text-white/60">
                <span>库存：{{ product.stock }}</span>
                <span>开售：{{ formatStart(product.start_time) }}</span>
              </div>
              <RouterLink
                :to="`/product/${product.id}`"
                class="inline-flex items-center justify-center rounded-full border border-magma/60 px-4 py-2 text-sm font-semibold text-white transition hover:bg-magma hover:text-white"
              >
                进入战场
              </RouterLink>
            </div>
          </div>
        </ParallaxCard>
      </div>
    </section>
  </MainLayout>
</template>
