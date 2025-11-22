<script setup lang="ts">
import { onMounted, computed } from "vue"
import { RouterLink } from "vue-router"
import { useProductStore } from "@/stores/productStore"
import MainLayout from "@/layout/MainLayout.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import CountdownPill from "@/components/CountdownPill.vue"
import { formatPrice } from "@/lib/utils"

const productStore = useProductStore()

onMounted(() => {
  productStore.fetchProducts(1, 12)
})

const heroProduct = computed(() => productStore.items[0])
</script>

<template>
  <MainLayout>
    <section class="relative overflow-hidden border-b border-obsidian-border/60 bg-gradient-to-br from-black via-obsidian-card to-black">
      <div class="mx-auto flex max-w-6xl flex-col gap-10 px-6 py-16 md:flex-row md:items-center">
        <div class="flex-1 space-y-6">
          <p class="text-sm uppercase tracking-[0.4em] text-magma">Midnight Magma Drop</p>
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

    <section class="mx-auto max-w-6xl px-6 py-12">
      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-2xl font-semibold">全部球鞋</h2>
        <p class="text-sm text-white/60">实时库存 · 秒杀预备</p>
      </div>

      <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <ParallaxCard v-for="item in productStore.items" :key="item.id">
          <RouterLink :to="`/product/${item.id}`" class="block space-y-4">
            <div class="relative overflow-hidden rounded-xl">
              <img :src="item.image || '/placeholder.svg'" alt="" class="h-48 w-full object-cover" />
              <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-transparent to-transparent"></div>
              <div class="absolute left-3 top-3 rounded-full bg-white/10 px-3 py-1 text-xs text-white">
                {{ new Date(item.start_time).toLocaleString() }}
              </div>
            </div>
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold">{{ item.name }}</h3>
                <span class="text-magma font-semibold">{{ formatPrice(item.price) }}</span>
              </div>
              <div class="flex items-center justify-between text-sm text-white/60">
                <span>库存 {{ item.stock }}</span>
                <CountdownPill :start-time="item.start_time" />
              </div>
              <div class="h-2 w-full overflow-hidden rounded-full bg-obsidian-border">
                <div class="h-full bg-magma" :style="{ width: `${item.stock > 0 ? 100 : 0}%` }"></div>
              </div>
            </div>
          </RouterLink>
        </ParallaxCard>
      </div>
    </section>
  </MainLayout>
</template>
