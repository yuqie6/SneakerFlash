<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed, watch } from "vue"
import { RouterLink, useRoute, useRouter } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import ParallaxCard from "@/components/motion/ParallaxCard.vue"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { useProductStore } from "@/stores/productStore"
import { useUserStore } from "@/stores/userStore"
import { useSeckill } from "@/composables/useSeckill"
import { useCountDown } from "@/composables/useCountDown"
import { formatPrice } from "@/lib/utils"
import type { Product } from "@/types/product"
import type { Coupon } from "@/types/coupon"
import type { OrderWithPayment } from "@/types/order"
import api, { resolveAssetUrl } from "@/lib/api"
import { toast } from "vue-sonner"

const route = useRoute()
const router = useRouter()
const productStore = useProductStore()
const userStore = useUserStore()
const product = ref<Product | null>(null)
const { status, resultMsg, executeSeckill } = useSeckill()
const pollingTimer = ref<number>()
const availableCoupons = ref<Coupon[]>([])
const couponsLoading = ref(false)
const selectedCouponId = ref<number | null>(null)
const orderResult = ref<OrderWithPayment | null>(null)
const orderCreating = ref(false)

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

const productPriceCents = computed(() => Math.round((product.value?.price || 0) * 100))
const usableCoupons = computed(() => availableCoupons.value.filter((c) => productPriceCents.value >= c.min_spend_cents))
const selectedCoupon = computed(() => usableCoupons.value.find((c) => c.id === selectedCouponId.value) || null)
const payableAmount = computed(() => (orderResult.value?.payment?.amount_cents || 0) / 100)
const savedAmount = computed(() => {
  if (!product.value) return 0
  const saved = product.value.price - payableAmount.value
  return saved > 0 ? saved : 0
})

const couponText = (coupon: Coupon) => {
  if (coupon.type === "discount") {
    const value = (coupon.discount_rate / 10).toFixed(1).replace(/\.0$/, "")
    return `${value}折`
  }
  return formatPrice(coupon.amount_cents / 100)
}

const couponRule = (coupon: Coupon) => {
  if (coupon.min_spend_cents <= 0) return "无门槛"
  return `满 ${formatPrice(coupon.min_spend_cents / 100)} 可用`
}

const fetchAvailableCoupons = async () => {
  if (!userStore.accessToken) return
  couponsLoading.value = true
  try {
    const res = await api.get<Coupon[], Coupon[]>("/coupons/mine", { params: { status: "available" } })
    availableCoupons.value = Array.isArray(res) ? res : []
  } catch (err: any) {
    toast.error(err?.message || "获取优惠券失败")
  } finally {
    couponsLoading.value = false
  }
}

const createOrder = async () => {
  if (!product.value) return
  if (!userStore.accessToken) {
    toast.error("请先登录")
    router.push({ name: "login", query: { redirect: route.fullPath } })
    return
  }
  const payload: { product_id: number; coupon_id?: number } = { product_id: product.value.id }
  if (selectedCoupon.value) {
    payload.coupon_id = selectedCoupon.value.id
  }
  orderCreating.value = true
  try {
    const res = await api.post<OrderWithPayment, OrderWithPayment>("/orders", payload)
    orderResult.value = res
    await fetchAvailableCoupons()
    toast.success("下单成功", { description: "支付单已生成，可前往订单中心查看" })
  } catch (err: any) {
    toast.error(err?.message || "下单失败")
  } finally {
    orderCreating.value = false
  }
}

watch(
  () => userStore.accessToken,
  (token) => {
    if (token) fetchAvailableCoupons()
  },
  { immediate: true }
)

watch(
  usableCoupons,
  (list) => {
    if (!list.length) {
      selectedCouponId.value = null
      return
    }
    if (!list.some((c) => c.id === selectedCouponId.value)) {
      selectedCouponId.value = list[0]?.id ?? null
    }
  },
  { immediate: true }
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
              </div>
            </CardContent>
          </Card>

          <Card class="border-obsidian-border/80 bg-obsidian-card/80">
            <CardHeader class="space-y-1">
              <p class="text-sm uppercase tracking-[0.3em] text-magma">Order & Coupon</p>
              <CardTitle class="text-2xl font-semibold">下单 / 优惠券</CardTitle>
              <CardDescription>非秒杀下单演示，可选一张优惠券，生成支付单。</CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="flex items-center justify-between text-sm text-white/70">
                <span>标价</span>
                <span class="text-lg font-semibold text-magma">{{ formatPrice(product.price) }}</span>
              </div>
              <div class="space-y-2 rounded-2xl border border-obsidian-border/70 bg-black/30 p-4">
                <div class="flex items-center justify-between text-sm text-white/70">
                  <span>选择优惠券</span>
                  <span class="text-xs text-white/60">{{ couponsLoading ? "加载中..." : `可用 ${usableCoupons.length} 张` }}</span>
                </div>
                <select
                  v-model="selectedCouponId"
                  :disabled="couponsLoading || !userStore.accessToken || usableCoupons.length === 0"
                  class="w-full rounded-lg border border-obsidian-border/70 bg-black/50 px-3 py-2 text-sm text-white outline-none transition focus:border-magma disabled:opacity-50"
                >
                  <option :value="null">不使用优惠券</option>
                  <option v-for="coupon in usableCoupons" :key="coupon.id" :value="coupon.id">
                    {{ couponText(coupon) }} · {{ couponRule(coupon) }}
                  </option>
                </select>
                <p class="text-xs text-white/60">仅显示满足门槛的可用券，未登录则不可用。</p>
              </div>
              <div class="flex items-center justify-between text-sm text-white/70">
                <span>应付金额</span>
                <span class="text-xl font-semibold text-magma">
                  {{ formatPrice(orderResult ? payableAmount : product.price) }}
                </span>
              </div>
              <div v-if="orderResult && savedAmount > 0" class="flex items-center justify-between text-xs text-white/60">
                <span>已优惠</span>
                <span class="text-emerald-300">{{ formatPrice(savedAmount) }}</span>
              </div>
              <MagmaButton class="w-full justify-center" :loading="orderCreating" @click="createOrder">
                创建订单（非秒杀）
              </MagmaButton>
              <div
                v-if="orderResult"
                class="space-y-2 rounded-2xl border border-obsidian-border/70 bg-black/30 p-3 text-xs text-white/70"
              >
                <div class="flex items-center justify-between">
                  <span>订单号</span>
                  <span class="font-mono">{{ orderResult.order.order_num }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span>支付单</span>
                  <span class="font-mono">{{ orderResult.payment?.payment_id }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span>支付金额</span>
                  <span class="text-magma font-semibold">{{ formatPrice(payableAmount) }}</span>
                </div>
                <RouterLink
                  :to="`/orders/${orderResult.order.id}`"
                  class="inline-flex items-center gap-1 text-magma underline decoration-magma/60 underline-offset-4"
                >
                  查看订单
                </RouterLink>
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
