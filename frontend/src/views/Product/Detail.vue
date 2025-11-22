<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useRoute } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import { useProductStore } from "@/stores/productStore"
import { useSeckill } from "@/composables/useSeckill"
import { useCountDown } from "@/composables/useCountDown"
import { formatPrice } from "@/lib/utils"
import type { Product } from "@/types/product"

const route = useRoute()
const productStore = useProductStore()
const product = ref<Product | null>(null)
const loading = ref(false)
const countdownTarget = ref<Date | string | number>(Date.now())

const { status, resultMsg, executeSeckill } = useSeckill()
const { formatted, isStarted } = useCountDown(countdownTarget)

const canBuy = computed(() => isStarted.value && (product.value?.stock ?? 0) > 0)

const load = async () => {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const detail = await productStore.fetchProductDetail(id)
    product.value = detail
    countdownTarget.value = detail.start_time
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <MainLayout>
    <div class="mx-auto grid max-w-6xl grid-cols-1 gap-10 px-6 py-12 md:grid-cols-2">
      <div class="relative">
        <ParallaxCard class="h-full">
          <div class="relative flex h-full flex-col justify-between overflow-hidden rounded-2xl bg-gradient-to-br from-black via-obsidian-card to-magma/15 p-8">
            <div class="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(249,115,22,0.15),transparent_35%)]"></div>
            <div class="relative z-10 flex h-full flex-col justify-between">
              <div>
                <p class="text-sm uppercase text-white/60">Sneaker</p>
                <h2 class="mt-3 text-3xl font-semibold">{{ product?.name || "加载中" }}</h2>
                <p class="mt-2 text-magma text-lg font-semibold">{{ product ? formatPrice(product.price) : "" }}</p>
              </div>
              <div class="mt-8 space-y-2 text-sm text-white/70">
                <div class="flex items-center justify-between">
                  <span>库存</span>
                  <span class="font-semibold text-white">{{ product?.stock ?? "-" }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span>开售时间</span>
                  <span>{{ product?.start_time }}</span>
                </div>
              </div>
            </div>
          </div>
        </ParallaxCard>
      </div>

      <div class="space-y-6">
        <div>
          <p class="text-xs uppercase text-magma">Product Detail</p>
          <h1 class="mt-2 text-3xl font-semibold">{{ product?.name || "商品详情" }}</h1>
          <p class="mt-3 text-white/70">
            极致流畅的磁吸按钮与物理动效，开售后立即锁定，提前完成准备，提升命中率。
          </p>
        </div>

        <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/70 p-6 backdrop-blur-xl">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-white/60">价格</p>
              <p class="text-2xl font-semibold text-magma">{{ product ? formatPrice(product.price) : "-" }}</p>
            </div>
            <div class="text-right">
              <p class="text-sm text-white/60">状态</p>
              <p class="text-lg font-semibold">
                <span v-if="!isStarted">倒计时 {{ formatted }}</span>
                <span v-else>进行中</span>
              </p>
            </div>
          </div>

          <div class="mt-6 space-y-3">
            <MagmaButton
              class="w-full justify-center"
              :loading="status === 'loading'"
              :disabled="!product || !canBuy || status === 'success'"
              @click="product && executeSeckill(product.id)"
            >
              <span v-if="!isStarted">倒计时 {{ formatted }}</span>
              <span v-else-if="status === 'loading'">正在锁定...</span>
              <span v-else-if="status === 'success'">GOT 'EM</span>
              <span v-else-if="status === 'failed'">重试抢购</span>
              <span v-else>立即抢购</span>
            </MagmaButton>
            <p v-if="status === 'success' || status === 'failed'" class="text-center text-sm text-white/70">
              {{ resultMsg }}
            </p>
            <p v-else class="text-center text-xs text-white/50">需要登录后才能下单，未开始时按钮将显示倒计时。</p>
          </div>
        </div>

        <div v-if="loading" class="text-sm text-white/60">正在加载商品信息...</div>
      </div>
    </div>
  </MainLayout>
</template>
