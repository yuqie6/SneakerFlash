<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed } from "vue"
import { RouterLink, useRoute, useRouter } from "vue-router"
import { Clock, Package, Zap, ShoppingBag } from "lucide-vue-next"
import MainLayout from "@/layout/MainLayout.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { useProductStore } from "@/stores/productStore"
import { useSeckill, type SeckillResult } from "@/composables/useSeckill"
import { useCountDown } from "@/composables/useCountDown"
import { formatPrice } from "@/lib/utils"
import type { Product } from "@/types/product"
import type { OrderWithPayment } from "@/types/order"
import api, { resolveAssetUrl } from "@/lib/api"
import { toast } from "vue-sonner"

type OrderPollResponse = {
	status: "pending" | "ready" | "failed"
	order_num: string
	payment_id?: string
	order?: OrderWithPayment
	message?: string
}

const route = useRoute()
const router = useRouter()
const productStore = useProductStore()
const product = ref<Product | null>(null)
const { status, resultMsg, executeSeckill } = useSeckill()
const pollingTimer = ref<number>()
const orderPollTimer = ref<number>()
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
  if (orderPollTimer.value) clearInterval(orderPollTimer.value)
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
  if (!res) return
  seckillResult.value = res
  if (res.order_id) {
    router.push({ name: "order-detail", params: { id: res.order_id } })
    return
  }
  if (res.order_num) {
    startOrderPolling(res.order_num)
  }
}

const pollOrderStatus = async (orderNum: string) => {
  if (!orderNum) return
  try {
    const res = await api.get<OrderPollResponse, OrderPollResponse>(`/orders/poll/${orderNum}`)
    if (res.status === "ready" && res.order?.order?.id) {
      seckillResult.value = {
        order_num: res.order.order.order_num,
        order_id: res.order.order.id,
        payment_id: res.order.payment?.payment_id || seckillResult.value?.payment_id || "",
        status: "ready",
      }
      router.push({ name: "order-detail", params: { id: res.order.order.id } })
      if (orderPollTimer.value) clearInterval(orderPollTimer.value)
    } else if (res.status === "failed") {
      toast.error(res.message || "订单生成失败")
      if (orderPollTimer.value) clearInterval(orderPollTimer.value)
    }
  } catch (err: any) {
    // 静默重试，长时间失败用户可手动刷新
    console.error(err)
  }
}

const startOrderPolling = (orderNum: string) => {
  pollOrderStatus(orderNum)
  if (orderPollTimer.value) clearInterval(orderPollTimer.value)
  orderPollTimer.value = window.setInterval(() => pollOrderStatus(orderNum), 1500)
}

const isLoading = computed(() => status.value === "loading")
const progressValue = computed(() => Math.max(0, Math.min(100, product.value?.stock || 0)))
const buttonClass = computed(() => ({
  "animate-shake": status.value === "failed",
}))
const cover = computed(
  () => resolveAssetUrl(product.value?.image) || "https://dummyimage.com/900x600/0f0f14/ffffff&text=SneakerFlash"
)

const stockStatus = computed(() => {
  const stock = product.value?.stock || 0
  if (stock === 0) return { text: "已售罄", tone: "border-red-500/40 bg-red-500/15 text-red-200" }
  if (stock <= 10) return { text: "库存紧张", tone: "border-amber-500/40 bg-amber-500/15 text-amber-200" }
  return { text: "库存充足", tone: "border-emerald-500/40 bg-emerald-500/15 text-emerald-200" }
})

const seckillStatus = computed(() => {
  if (!isStarted.value) return { text: "即将开始", tone: "border-magma/40 bg-magma/15 text-magma" }
  if (product.value?.stock === 0) return { text: "已结束", tone: "border-white/20 bg-white/5 text-white/50" }
  return { text: "抢购中", tone: "border-emerald-500/40 bg-emerald-500/15 text-emerald-200" }
})

const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return "--"
  const d = new Date(dateStr)
  if (Number.isNaN(d.getTime())) return dateStr
  return d.toLocaleString()
}
</script>

<template>
  <MainLayout>
    <section class="relative mx-auto max-w-6xl px-6 py-12">
      <div class="pointer-events-none absolute inset-0 opacity-70 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-0 right-0 h-80 w-80 rounded-full bg-[#ea580c55] blur-3xl"></div>
      </div>

      <div v-if="product" class="relative flex flex-col gap-6">
        <!-- 页面标题区域 -->
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div class="space-y-2">
            <p class="flex items-center gap-2 text-sm uppercase tracking-[0.3em] text-magma">
              <Zap class="h-4 w-4" />
              Flash Sale
            </p>
            <h1 class="text-3xl font-semibold">{{ product.name }}</h1>
          </div>
          <div class="flex items-center gap-3">
            <span class="rounded-full border px-3 py-1 text-xs" :class="seckillStatus.tone">
              {{ seckillStatus.text }}
            </span>
            <span class="rounded-full border px-3 py-1 text-xs" :class="stockStatus.tone">
              {{ stockStatus.text }}
            </span>
          </div>
        </div>

        <!-- 信息概览四宫格 -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="text-sm text-white/60">商品价格</p>
            <div class="mt-2 text-2xl font-semibold text-magma">{{ formatPrice(product.price) }}</div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="text-sm text-white/60">剩余库存</p>
            <div class="mt-2 flex items-center gap-2 text-xl font-semibold text-white">
              {{ product.stock }}
              <span class="rounded-full border border-obsidian-border px-2 py-0.5 text-xs text-white/70">件</span>
            </div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="text-sm text-white/60">开抢时间</p>
            <div class="mt-2 text-sm font-semibold text-white">{{ formatDateTime(product.start_time) }}</div>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="text-sm text-white/60">商品 ID</p>
            <div class="mt-2 font-mono text-sm font-semibold text-white/80">#{{ product.id }}</div>
          </div>
        </div>

        <!-- 主要内容区：图片 + 秒杀卡片 -->
        <div class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <!-- 左侧：商品图片 -->
          <div class="relative">
            <ParallaxCard class="glass">
              <img :src="cover" alt="" class="h-[420px] w-full rounded-2xl object-cover lg:h-[480px]" />
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

          <!-- 右侧：秒杀信息 -->
          <div class="flex flex-col gap-4">
            <Card class="border-obsidian-border/70 bg-gradient-to-br from-obsidian-card via-black to-obsidian-card">
              <CardHeader>
                <CardTitle class="flex items-center gap-2 text-xl">
                  <ShoppingBag class="h-5 w-5 text-magma" />
                  秒杀抢购
                </CardTitle>
                <CardDescription>限时特惠，手快有手慢无</CardDescription>
              </CardHeader>
              <CardContent class="space-y-5">
                <!-- 库存进度 -->
                <div class="relative overflow-hidden rounded-2xl border border-obsidian-border/70 bg-gradient-to-r from-white/5 via-black/70 to-black/40 p-5">
                  <div class="pointer-events-none absolute inset-0">
                    <div class="absolute left-10 top-0 h-full w-1/2 bg-magma-glow blur-3xl"></div>
                  </div>
                  <div class="relative space-y-4">
                    <div class="flex items-center justify-between text-sm">
                      <span class="text-white/70">库存进度</span>
                      <span class="rounded-full border border-obsidian-border px-3 py-1 text-xs text-white/60">
                        剩余 {{ product.stock }} 件
                      </span>
                    </div>
                    <div class="relative h-3 w-full overflow-hidden rounded-full bg-white/10">
                      <div
                        class="h-full rounded-full bg-gradient-to-r from-magma to-amber-200 shadow-[0_0_20px_rgba(234,88,12,0.35)] transition-[width] duration-500 ease-out"
                        :style="`width: ${progressValue}%`"
                      >
                        <div class="absolute right-0 top-1/2 h-4 w-4 -translate-y-1/2 rounded-full border border-white/60 bg-white shadow-[0_0_15px_rgba(255,255,255,0.7)]"></div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 倒计时 + 抢购按钮 -->
                <div class="rounded-2xl border border-obsidian-border/80 bg-black/30 p-5">
                  <div class="mb-4 flex items-center justify-between">
                    <div class="flex items-center gap-2 text-sm text-white/70">
                      <Clock class="h-4 w-4" />
                      <span>{{ isStarted ? "活动进行中" : "距离开始" }}</span>
                    </div>
                    <span class="text-xl font-semibold text-magma">{{ isStarted ? "立即参与" : formatted }}</span>
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
                    class="mt-4 flex items-center justify-between rounded-xl border border-emerald-500/40 bg-emerald-500/10 px-4 py-3 text-sm"
                  >
                    <div>
                      <p class="font-semibold text-emerald-300">
                        {{ seckillResult.status === "pending" ? "订单生成中" : "已锁定订单，前往支付" }}
                      </p>
                      <p class="text-xs text-white/60">订单号：{{ seckillResult.order_num }}</p>
                    </div>
                    <RouterLink
                      v-if="seckillResult.order_id"
                      :to="`/orders/${seckillResult.order_id}`"
                      class="rounded-full bg-emerald-500 px-3 py-1 text-xs font-semibold text-black transition hover:bg-emerald-400"
                    >
                      去支付
                    </RouterLink>
                    <span v-else class="text-xs text-emerald-200">正在生成订单...</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            <!-- 购买须知 -->
            <Card class="border-obsidian-border/70 bg-obsidian-card/80">
              <CardHeader>
                <CardTitle class="flex items-center gap-2 text-lg">
                  <Package class="h-5 w-5 text-magma" />
                  购买须知
                </CardTitle>
              </CardHeader>
              <CardContent class="space-y-3 text-sm text-white/70">
                <div class="flex items-start gap-3">
                  <span class="mt-1 h-2 w-2 rounded-full bg-magma"></span>
                  <div>
                    <p class="font-semibold text-white">一键抢购</p>
                    <p class="text-white/60">点击立即抢购会实时扣减库存并创建订单</p>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <span class="mt-1 h-2 w-2 rounded-full bg-emerald-400"></span>
                  <div>
                    <p class="font-semibold text-white">自动跳转</p>
                    <p class="text-white/60">锁定成功后自动跳转到订单详情页完成支付</p>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <span class="mt-1 h-2 w-2 rounded-full bg-white/60"></span>
                  <div>
                    <p class="font-semibold text-white">失败可重试</p>
                    <p class="text-white/60">库存不足或重复抢购会提示原因，不会生成多余订单</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      <!-- 加载骨架屏 -->
      <div v-else class="relative flex flex-col gap-6">
        <div class="h-10 w-1/3 animate-pulse rounded-lg bg-white/10"></div>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="i in 4" :key="i" class="h-24 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
        </div>
        <div class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <div class="h-[480px] animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
          <div class="space-y-4">
            <div class="h-64 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
            <div class="h-48 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
          </div>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
