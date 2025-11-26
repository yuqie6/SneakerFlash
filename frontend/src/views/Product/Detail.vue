<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed } from "vue"
import { RouterLink, useRoute, useRouter } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { useProductStore } from "@/stores/productStore"
import { useSeckill, type SeckillResult } from "@/composables/useSeckill"
import { useCountDown } from "@/composables/useCountDown"
import { formatPrice } from "@/lib/utils"
import type { Product } from "@/types/product"
import { resolveAssetUrl } from "@/lib/api"

const route = useRoute()
const router = useRouter()
const productStore = useProductStore()
const product = ref<Product | null>(null)
const { status, resultMsg, executeSeckill } = useSeckill()
const pollingTimer = ref<number>()
const seckillResult = ref<SeckillResult | null>(null)

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

const onSeckill = async () => {
  if (!product.value) return
  const res = await executeSeckill(product.value.id)
  if (res?.order_id) {
    seckillResult.value = res
    router.push({ name: "order-detail", params: { id: res.order_id } })
  }
}

const isLoading = computed(() => status.value === "loading")
const progressValue = computed(() => Math.max(0, Math.min(100, product.value?.stock || 0)))
const buttonClass = computed(() => ({
  "animate-shake": status.value === "failed",
}))
const cover = computed(
  () => resolveAssetUrl(product.value?.image) || "https://dummyimage.com/900x600/0f0f14/ffffff&text=SneakerFlash"
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

        <div class="flex flex-col gap-4">
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
                <div
                  v-if="seckillResult"
                  class="mt-3 flex items-center justify-between rounded-xl border border-emerald-500/40 bg-emerald-500/10 px-4 py-3 text-sm text-white/80"
                >
                  <div>
                    <p class="font-semibold text-emerald-300">已锁定订单，前往支付</p>
                    <p class="text-xs text-white/60">订单号：{{ seckillResult.order_num }}</p>
                  </div>
                  <RouterLink
                    :to="`/orders/${seckillResult.order_id}`"
                    class="rounded-full bg-emerald-500 px-3 py-1 text-xs font-semibold text-black transition hover:bg-emerald-400"
                  >
                    去支付
                  </RouterLink>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card class="border-obsidian-border/80 bg-obsidian-card/80">
            <CardHeader class="space-y-1">
              <p class="text-sm uppercase tracking-[0.3em] text-magma">Payment</p>
              <CardTitle class="text-2xl font-semibold">支付指引</CardTitle>
              <CardDescription>抢购成功后自动生成支付单并跳转到支付页。</CardDescription>
            </CardHeader>
            <CardContent class="space-y-3 text-sm text-white/70">
              <div class="flex items-start gap-3">
                <span class="mt-1 h-2 w-2 rounded-full bg-magma"></span>
                <div>
                  <p class="font-semibold text-white">一键抢购</p>
                  <p class="text-white/60">点击“立即抢购”会实时扣减库存并创建订单/支付单。</p>
                </div>
              </div>
              <div class="flex items-start gap-3">
                <span class="mt-1 h-2 w-2 rounded-full bg-emerald-400"></span>
                <div>
                  <p class="font-semibold text-white">跳转支付页</p>
                  <p class="text-white/60">锁定成功后跳转到订单详情，支持立即支付或稍后在订单中心查看。</p>
                </div>
              </div>
              <div class="flex items-start gap-3">
                <span class="mt-1 h-2 w-2 rounded-full bg-white/60"></span>
                <div>
                  <p class="font-semibold text-white">失败可重试</p>
                  <p class="text-white/60">库存不足或重复抢购会提示原因，不会生成多余订单。</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <div v-else class="relative grid gap-4 rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-6 text-white/70">
        <div class="h-6 w-1/2 animate-pulse rounded bg-white/10"></div>
        <div class="h-4 w-full animate-pulse rounded bg-white/10"></div>
        <div class="h-4 w-2/3 animate-pulse rounded bg-white/10"></div>
      </div>
    </section>
  </MainLayout>
</template>
