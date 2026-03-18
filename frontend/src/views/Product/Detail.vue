<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed } from "vue"
import { RouterLink, useRoute, useRouter } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
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
const orderPollErrorCount = ref(0)

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
      orderPollErrorCount.value = 0
      seckillResult.value = {
        order_num: res.order.order.order_num,
        order_id: res.order.order.id,
        payment_id: res.order.payment?.payment_id || seckillResult.value?.payment_id || "",
        status: "ready",
      }
      router.push({ name: "order-detail", params: { id: res.order.order.id } })
      if (orderPollTimer.value) clearInterval(orderPollTimer.value)
    } else if (res.status === "failed") {
      orderPollErrorCount.value = 0
      toast.error(res.message || "订单生成失败")
      if (orderPollTimer.value) clearInterval(orderPollTimer.value)
    }
  } catch (err: unknown) {
    orderPollErrorCount.value += 1
    if (orderPollErrorCount.value >= 3) {
      if (orderPollTimer.value) clearInterval(orderPollTimer.value)
      const message = err instanceof Error ? err.message : "订单轮询失败，请稍后到订单页查看"
      toast.error(message)
    }
  }
}

const startOrderPolling = (orderNum: string) => {
  orderPollErrorCount.value = 0
  pollOrderStatus(orderNum)
  if (orderPollTimer.value) clearInterval(orderPollTimer.value)
  orderPollTimer.value = window.setInterval(() => pollOrderStatus(orderNum), 1500)
}

const isLoading = computed(() => status.value === "loading")
const isSoldOut = computed(() => (product.value?.stock || 0) <= 0)
const progressValue = computed(() => Math.max(0, Math.min(100, product.value?.stock || 0)))
const buttonClass = computed(() => ({
  "animate-shake": status.value === "failed",
}))
const cover = computed(
  () => resolveAssetUrl(product.value?.image) || "https://dummyimage.com/900x600/F9F8F6/1C1C1C&text=SneakerFlash"
)

const stockStatus = computed(() => {
  const stock = product.value?.stock || 0
  if (stock === 0) return { text: "已售罄", tone: "border-[#1C1C1C]/30 text-[#1C1C1C]/50" }
  if (stock <= 10) return { text: "库存紧张", tone: "border-[#1C1C1C]/25 text-[#1C1C1C]/70" }
  return { text: "库存充足", tone: "border-[#1C1C1C]/20 text-[#1C1C1C]/70" }
})

const seckillStatus = computed(() => {
  const endTime = product.value?.end_time
  if (endTime && new Date(endTime) < new Date()) {
    return { text: "已结束", tone: "border-[#1C1C1C]/10 text-[#1C1C1C]/40" }
  }
  if (!isStarted.value) return { text: "即将开始", tone: "border-[#1C1C1C]/20 text-[#1C1C1C]/70" }
  if (product.value?.stock === 0) return { text: "已售罄", tone: "border-[#1C1C1C]/10 text-[#1C1C1C]/40" }
  return { text: "抢购中", tone: "border-[#1C1C1C]/20 text-[#1C1C1C]/70" }
})

const isEnded = computed(() => {
  const endTime = product.value?.end_time
  return !!(endTime && new Date(endTime) < new Date())
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
    <section class="mx-auto max-w-6xl px-6 py-16 md:py-24">
      <div v-if="product" class="flex flex-col gap-8">
        <!-- 页面标题 -->
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div class="space-y-1">
            <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Flash Sale</p>
            <h1 class="font-serif text-3xl tracking-tight md:text-5xl">{{ product.name }}</h1>
          </div>
          <div class="flex items-center gap-3">
            <span class="border px-3 py-1 text-xs" :class="seckillStatus.tone">{{ seckillStatus.text }}</span>
            <span class="border px-3 py-1 text-xs" :class="stockStatus.tone">{{ stockStatus.text }}</span>
          </div>
        </div>

        <!-- 四宫格 -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Price</p>
            <div class="mt-2 text-2xl">{{ formatPrice(product.price) }}</div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Stock</p>
            <div class="mt-2 text-xl">{{ product.stock }} <span class="text-xs text-[#1C1C1C]/40">件</span></div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Start</p>
            <div class="mt-2 text-sm">{{ formatDateTime(product.start_time) }}</div>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">End</p>
            <div class="mt-2 text-sm" :class="isEnded ? 'text-[#1C1C1C]/40' : ''">
              {{ product.end_time ? formatDateTime(product.end_time) : '永不过期' }}
            </div>
          </div>
        </div>

        <!-- 主内容：图片 + 秒杀 -->
        <div class="grid gap-8 lg:grid-cols-[1.1fr_0.9fr]">
          <!-- 图片 -->
          <div class="relative">
            <div class="border border-[#1C1C1C]/10">
              <img :src="cover" alt="" class="h-[420px] w-full object-cover lg:h-[480px]" />
            </div>
            <div
              v-if="isLoading"
              class="absolute inset-0 flex items-center justify-center bg-white/80 transition"
            >
              <div class="flex items-center gap-2 text-sm text-[#1C1C1C]/60">
                <span class="h-3 w-3 animate-pulse bg-[#1C1C1C]"></span>
                锁定中，请稍候...
              </div>
            </div>
          </div>

          <!-- 秒杀信息 -->
          <div class="flex flex-col gap-6">
            <Card>
              <CardHeader>
                <CardTitle class="font-serif text-xl tracking-tight">秒杀抢购</CardTitle>
                <CardDescription class="text-[#1C1C1C]/40">限时特惠，手快有手慢无</CardDescription>
              </CardHeader>
              <CardContent class="space-y-5">
                <!-- 库存进度 -->
                <div class="border border-[#1C1C1C]/10 p-5">
                  <div class="flex items-center justify-between text-sm">
                    <span class="text-[#1C1C1C]/60">库存进度</span>
                    <span class="text-xs text-[#1C1C1C]/40">剩余 {{ product.stock }} 件</span>
                  </div>
                  <div class="relative mt-4 h-1 w-full bg-[#1C1C1C]/10">
                    <div
                      class="h-full bg-[#1C1C1C] transition-[width] duration-500 ease-out"
                      :style="`width: ${progressValue}%`"
                    ></div>
                  </div>
                </div>

                <!-- 倒计时 + 按钮 -->
                <div class="border border-[#1C1C1C]/10 p-5">
                  <div class="mb-4 flex items-center justify-between">
                    <span class="text-sm text-[#1C1C1C]/60">{{ isStarted ? "活动进行中" : "距离开始" }}</span>
                    <span class="font-serif text-xl">{{ isStarted ? "立即参与" : formatted }}</span>
                  </div>
                  <MagmaButton
                    class="w-full justify-center"
                    :loading="status === 'loading'"
                    :disabled="!isStarted || status === 'success' || isEnded || isSoldOut"
                    :class="buttonClass"
                    @click="onSeckill"
                  >
                    <span v-if="isEnded">活动已结束</span>
                    <span v-else-if="isSoldOut">已售罄</span>
                    <span v-else-if="buttonState === 'pending'">即将开始 · {{ formatted }}</span>
                    <span v-else-if="buttonState === 'loading'">锁定中...</span>
                    <span v-else-if="buttonState === 'success'">GOT 'EM · {{ resultMsg }}</span>
                    <span v-else-if="buttonState === 'failed'">再试一次 · {{ resultMsg }}</span>
                    <span v-else>立即抢购</span>
                  </MagmaButton>
                  <div
                    v-if="seckillResult"
                    class="mt-4 flex items-center justify-between border border-[#1C1C1C]/20 bg-[#1C1C1C]/5 px-4 py-3 text-sm"
                  >
                    <div>
                      <p class="font-medium">
                        {{ seckillResult.status === "pending" ? "订单生成中" : "已锁定订单，前往支付" }}
                      </p>
                      <p class="text-xs text-[#1C1C1C]/40">订单号：{{ seckillResult.order_num }}</p>
                    </div>
                    <RouterLink
                      v-if="seckillResult.order_id"
                      :to="`/orders/${seckillResult.order_id}`"
                      class="bg-[#1C1C1C] px-3 py-1 text-xs text-white transition-colors hover:bg-[#1C1C1C]/80"
                    >
                      去支付
                    </RouterLink>
                    <span v-else class="text-xs text-[#1C1C1C]/40">正在生成订单...</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            <!-- 购买须知 -->
            <Card>
              <CardHeader>
                <CardTitle class="font-serif text-lg tracking-tight">购买须知</CardTitle>
              </CardHeader>
              <CardContent class="space-y-3 text-sm text-[#1C1C1C]/60">
                <div class="flex items-start gap-3">
                  <span class="mt-1.5 h-1 w-1 bg-[#1C1C1C]/70"></span>
                  <div>
                    <p class="font-medium text-[#1C1C1C]">一键抢购</p>
                    <p class="text-[#1C1C1C]/40">点击立即抢购，成功后自动生成订单</p>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <span class="mt-1.5 h-1 w-1 bg-[#1C1C1C]"></span>
                  <div>
                    <p class="font-medium text-[#1C1C1C]">自动跳转</p>
                    <p class="text-[#1C1C1C]/40">锁定成功后自动跳转到订单详情页完成支付</p>
                  </div>
                </div>
                <div class="flex items-start gap-3">
                  <span class="mt-1.5 h-1 w-1 bg-[#1C1C1C]/40"></span>
                  <div>
                    <p class="font-medium text-[#1C1C1C]">失败可重试</p>
                    <p class="text-[#1C1C1C]/40">库存不足或重复抢购会提示原因，不会生成多余订单</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      <!-- 骨架屏 -->
      <div v-else class="flex flex-col gap-6">
        <div class="h-10 w-1/3 animate-pulse bg-[#1C1C1C]/5"></div>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="i in 4" :key="i" class="h-24 animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>
        </div>
        <div class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <div class="h-[480px] animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>
          <div class="space-y-4">
            <div class="h-64 animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>
            <div class="h-48 animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>
          </div>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
