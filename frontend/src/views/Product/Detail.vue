<script setup lang="ts">
import { onMounted, ref, computed } from "vue"
import { useRoute } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import { useProductStore } from "@/stores/productStore"
import { useSeckill } from "@/composables/useSeckill"
import { useCountDown } from "@/composables/useCountDown"
import { formatPrice } from "@/lib/utils"
import type { Product } from "@/types/product"

const route = useRoute()
const productStore = useProductStore()
const product = ref<Product | null>(null)
const { status, resultMsg, executeSeckill } = useSeckill()

const load = async () => {
  const id = Number(route.params.id)
  product.value = await productStore.fetchProductDetail(id)
}

onMounted(load)

const { formatted, isStarted } = useCountDown(computed(() => product.value?.start_time || Date.now()))

const buttonState = computed(() => {
  if (!isStarted.value) return "pending"
  if (status.value === "loading") return "loading"
  if (status.value === "success") return "success"
  if (status.value === "failed") return "failed"
  return "active"
})

const onSeckill = () => {
  if (!product.value) return
  executeSeckill(product.value.id)
}
</script>

<template>
  <MainLayout>
    <section v-if="product" class="mx-auto max-w-6xl px-6 py-12">
      <div class="grid gap-10 lg:grid-cols-[1.1fr_0.9fr]">
        <ParallaxCard class="glass">
          <img :src="product.image || '/placeholder.svg'" alt="" class="h-[420px] w-full rounded-2xl object-cover" />
        </ParallaxCard>
        <div class="space-y-6">
          <div>
            <p class="text-sm uppercase tracking-[0.3em] text-magma">Seckill</p>
            <h1 class="text-3xl font-semibold">{{ product.name }}</h1>
          </div>
          <p class="text-2xl font-bold text-magma">{{ formatPrice(product.price) }}</p>
          <div class="space-y-2">
            <div class="flex items-center justify-between text-sm text-white/70">
              <span>库存</span>
              <span>{{ product.stock }}</span>
            </div>
            <div class="h-2 w-full overflow-hidden rounded-full bg-obsidian-border">
              <div class="h-full bg-magma" :style="{ width: `${product.stock > 0 ? 100 : 0}%` }"></div>
            </div>
          </div>

          <div class="rounded-2xl border border-obsidian-border/80 bg-obsidian-card/80 p-6">
            <div class="mb-4 flex items-center justify-between text-sm text-white/70">
              <span>开抢时间</span>
              <span>{{ new Date(product.start_time).toLocaleString() }}</span>
            </div>
            <div class="mb-4 flex items-center gap-3 text-lg font-semibold">
              <span v-if="!isStarted" class="text-white/70">距离开始</span>
              <span class="text-magma">{{ isStarted ? "进行中" : formatted }}</span>
            </div>
            <MagmaButton
              class="w-full justify-center"
              :loading="status === 'loading'"
              :disabled="!isStarted || status === 'success'"
              @click="onSeckill"
            >
              <span v-if="buttonState === 'pending'">即将开始 · {{ formatted }}</span>
              <span v-else-if="buttonState === 'loading'">锁定中...</span>
              <span v-else-if="buttonState === 'success'">GOT 'EM · {{ resultMsg }}</span>
              <span v-else-if="buttonState === 'failed'">再试一次 · {{ resultMsg }}</span>
              <span v-else>立即抢购</span>
            </MagmaButton>
          </div>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
