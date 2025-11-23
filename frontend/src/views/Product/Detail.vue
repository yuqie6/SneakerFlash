<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed, watch } from "vue"
import { useRoute } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { useProductStore } from "@/stores/productStore"
import { useSeckill } from "@/composables/useSeckill"
import { useCountDown } from "@/composables/useCountDown"
import { formatPrice } from "@/lib/utils"
import type { Product } from "@/types/product"
import { resolveAssetUrl } from "@/lib/api"
import { toast } from "vue-sonner"

const route = useRoute()
const productStore = useProductStore()
const product = ref<Product | null>(null)
const { status, resultMsg, executeSeckill } = useSeckill()
const pollingTimer = ref<number>()

const load = async () => {
  const id = Number(route.params.id)
  product.value = await productStore.fetchProductDetail(id)
}

onMounted(load)
onMounted(() => {
  pollingTimer.value = window.setInterval(async () => {
    if (!product.value) return
    try {
      const latest = await productStore.fetchProductDetail(product.value.id, true)
      product.value = latest
    } catch {
      /* ignore */
    }
  }, 5000)
})

onBeforeUnmount(() => {
  if (pollingTimer.value) clearInterval(pollingTimer.value)
})

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

const isLoading = computed(() => status.value === "loading")
const progressValue = computed(() => Math.max(0, Math.min(100, (product.value?.stock || 0))))
const buttonClass = computed(() => ({
  "animate-shake": status.value === "failed",
}))

const cover = computed(() => resolveAssetUrl(product.value?.image) || "https://dummyimage.com/900x600/0f0f14/ffffff&text=SneakerFlash")

watch(
  () => status.value,
  (val) => {
    if (val === "failed") {
      toast.error(resultMsg.value || "抢购失败")
    }
  }
)
</script>

<template>
  <MainLayout>
    <section class="relative mx-auto max-w-6xl px-6 py-12">
      <div class="pointer-events-none absolute inset-0 opacity-70 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-0 right-0 h-80 w-80 rounded-full bg-[#ea580c55] blur-3xl"></div>
      </div>
      <div v-if="product" class="relative grid gap-10 lg:grid-cols-[1.1fr_0.9fr]">
        <div class="relative">
          <ParallaxCard class="glass">
            <img :src="cover" alt="" class="h-[460px] w-full rounded-2xl object-cover" />
          </ParallaxCard>
          <div
            v-if="isLoading"
            class="absolute inset-0 flex items-center justify-center rounded-2xl bg-black/60 backdrop-blur-sm transition"
          >
            <div class="flex items-center gap-2 text-sm text-white/80">
              <span class="h-3 w-3 animate-pulse rounded-full bg-magma"></span>
              锁定中，请稍候...
            </div>
          </div>
        </div>

        <Card class="relative overflow-hidden border-obsidian-border/80 bg-obsidian-card/80">
          <CardHeader class="space-y-3">
            <p class="text-sm uppercase tracking-[0.3em] text-magma">Seckill</p>
            <CardTitle class="text-3xl font-semibold">{{ product.name }}</CardTitle>
            <CardDescription class="text-lg text-magma">{{ formatPrice(product.price) }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-6">
            <div class="space-y-2">
              <div class="flex items-center justify-between text-sm text-white/70">
                <span>库存</span>
                <span>{{ product.stock }}</span>
              </div>
              <Progress :model-value="progressValue" class="h-2 bg-obsidian-border" />
            </div>

            <div class="rounded-2xl border border-obsidian-border/80 bg-black/30 p-6">
              <div class="mb-3 flex items-center justify-between text-sm text-white/70">
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
                :class="buttonClass"
              >
                <span v-if="buttonState === 'pending'">即将开始 · {{ formatted }}</span>
                <span v-else-if="buttonState === 'loading'">锁定中...</span>
                <span v-else-if="buttonState === 'success'">GOT 'EM · {{ resultMsg }}</span>
                <span v-else-if="buttonState === 'failed'">再试一次 · {{ resultMsg }}</span>
                <span v-else>立即抢购</span>
              </MagmaButton>
            </div>
          </CardContent>
        </Card>
      </div>

      <div v-else class="relative grid gap-4 rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-6 text-white/70">
        <div class="h-6 w-1/2 animate-pulse rounded bg-white/10"></div>
        <div class="h-4 w-full animate-pulse rounded bg-white/10"></div>
        <div class="h-4 w-2/3 animate-pulse rounded bg-white/10"></div>
      </div>
    </section>
  </MainLayout>
</template>
