<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import api, { buildStreamUrl, resolveAssetUrl } from "@/lib/api"
import type { Order, OrderWithPayment } from "@/types/order"
import type { Payment } from "@/types/payment"
import type { Coupon } from "@/types/coupon"
import type { Product } from "@/types/product"
import { formatPrice } from "@/lib/utils"
import { toast } from "vue-sonner"

const route = useRoute()
const router = useRouter()
const data = ref<OrderWithPayment | null>(null)
const loading = ref(false)
const paying = ref(false)
const applyingCoupon = ref(false)
const product = ref<Product | null>(null)
const coupons = ref<Coupon[]>([])
const couponsLoading = ref(false)
const selectedCouponId = ref<number | null>(null)
let orderEventSource: EventSource | null = null
let productEventSource: EventSource | null = null
let pollTimer: number | null = null

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await api.get<OrderWithPayment, OrderWithPayment>(`/orders/${route.params.id}`)
    data.value = res
    selectedCouponId.value = res.coupon?.id ?? null
    if (res.order?.product_id) {
      await fetchProduct(res.order.product_id)
    }
    await fetchCoupons()
    stopRealtimeIfResolved()
  } catch (err: any) {
    toast.error(err?.message || "获取订单失败")
  } finally {
    loading.value = false
  }
}

const stopPolling = () => {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

const closeStreams = () => {
  orderEventSource?.close()
  productEventSource?.close()
  orderEventSource = null
  productEventSource = null
}

const stopAllRealtime = () => {
  stopPolling()
  closeStreams()
}

const startPollingFallback = () => {
  if (pollTimer !== null || !isPendingPayment.value) return
  pollTimer = window.setInterval(async () => {
    if (!isPendingPayment.value) {
      stopAllRealtime()
      return
    }
    await fetchDetail()
    stopRealtimeIfResolved()
  }, 5000)
}

const bindStreams = () => {
  closeStreams()
  stopPolling()

  const accessToken = localStorage.getItem("access_token") || ""
  if (!accessToken || !order.value?.id) return

  orderEventSource = new EventSource(buildStreamUrl(`/stream/orders/${order.value.id}`, accessToken))
  orderEventSource.onmessage = async () => {
    await fetchDetail()
    stopRealtimeIfResolved()
  }
  orderEventSource.onerror = () => {
    orderEventSource?.close()
    orderEventSource = null
    startPollingFallback()
  }

  if (order.value.product_id) {
    productEventSource = new EventSource(buildStreamUrl(`/stream/products/${order.value.product_id}`, accessToken))
    productEventSource.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data) as { event: string; data?: { stock?: number } }
        if (typeof payload.data?.stock === "number" && product.value) {
          product.value = { ...product.value, stock: payload.data.stock }
        }
      } catch {
        // ignore malformed event
      }
    }
    productEventSource.onerror = () => {
      productEventSource?.close()
      productEventSource = null
    }
  }
}

const payment = computed<Payment | undefined>(() => data.value?.payment)
const order = computed<Order | undefined>(() => data.value?.order)
const currentCoupon = computed<Coupon | null>(() => data.value?.coupon || null)
const isPendingPayment = computed(() => order.value?.status === 0 && payment.value?.status === "pending")
const basePrice = computed(() => product.value?.price ?? (payment.value ? payment.value.amount_cents / 100 : 0))
const payableAmount = computed(() => (payment.value ? payment.value.amount_cents / 100 : 0))
const savedAmount = computed(() => {
  const saved = basePrice.value - payableAmount.value
  return saved > 0 ? saved : 0
})
const basePriceCents = computed(() => Math.round(basePrice.value * 100))
const usableCoupons = computed(() => coupons.value.filter((c) => basePriceCents.value >= c.min_spend_cents))
const displayCoupons = computed(() => {
  const list = [...usableCoupons.value]
  if (currentCoupon.value && !list.some((c) => c.id === currentCoupon.value!.id)) {
    list.unshift(currentCoupon.value)
  }
  return list
})
const productCover = computed(
  () => resolveAssetUrl(product.value?.image) || "https://dummyimage.com/400x300/F9F8F6/1C1C1C&text=SneakerFlash"
)

const stopRealtimeIfResolved = () => {
  if (!isPendingPayment.value) {
    stopAllRealtime()
  }
}

const pay = async (status: Payment["status"]) => {
  if (!payment.value) {
    toast.error("暂无支付单")
    return
  }
  paying.value = true
  try {
    await api.post("/payment/callback", {
      payment_id: payment.value.payment_id,
      status,
      notify_data: "mock",
    })
    toast.success(status === "paid" ? "支付成功" : "支付失败")
    await fetchDetail()
    if (isPendingPayment.value) bindStreams()
    else {
      closeStreams()
      stopPolling()
    }
  } catch (err: any) {
    toast.error(err?.message || "支付操作失败")
  } finally {
    paying.value = false
  }
}

const fetchProduct = async (productId: number) => {
  try {
    const res = await api.get<Product, Product>(`/product/${productId}`)
    product.value = res
  } catch {
    /* ignore */
  }
}

const fetchCoupons = async () => {
  if (!isPendingPayment.value) return
  couponsLoading.value = true
  try {
    const res = await api.get<
      { list: Coupon[]; total: number; page: number; page_size: number },
      { list: Coupon[]; total: number; page: number; page_size: number }
    >("/coupons/mine", { params: { status: "available", page: 1, page_size: 100 } })
    coupons.value = Array.isArray(res.list) ? res.list : []
  } catch (err: any) {
    toast.error(err?.message || "获取优惠券失败")
  } finally {
    couponsLoading.value = false
  }
}

const applyCoupon = async () => {
  if (!order.value) return
  applyingCoupon.value = true
  try {
    const res = await api.post<OrderWithPayment, OrderWithPayment>(`/orders/${order.value.id}/apply-coupon`, {
      coupon_id: selectedCouponId.value,
    })
    data.value = res
    selectedCouponId.value = res.coupon?.id ?? null
    toast.success(selectedCouponId.value ? "已应用优惠券" : "已取消优惠券")
    await fetchCoupons()
    if (isPendingPayment.value) {
      bindStreams()
    } else {
      stopAllRealtime()
    }
  } catch (err: any) {
    toast.error(err?.message || "优惠券应用失败")
  } finally {
    applyingCoupon.value = false
  }
}

const orderStatusText = (s?: number) => {
  switch (s) {
    case 0: return "待支付"
    case 1: return "已支付"
    case 2: return "支付失败"
    case 3: return "已取消"
    default: return "未知"
  }
}

const paymentStatusText = (s?: Payment["status"]) => {
  switch (s) {
    case "pending": return "待支付"
    case "paid": return "已支付"
    case "failed": return "失败"
    case "refunded": return "已退款"
    default: return "未知"
  }
}

onMounted(async () => {
  await fetchDetail()
  if (isPendingPayment.value) {
    bindStreams()
  }
})

onUnmounted(() => {
  closeStreams()
  stopPolling()
})
</script>

<template>
  <MainLayout>
    <section class="mx-auto max-w-4xl px-6 py-16 md:py-24">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Order Detail</p>
          <h1 class="font-serif text-2xl tracking-tight md:text-3xl">订单详情</h1>
        </div>
        <button class="hover-underline text-sm text-[#1C1C1C]/60" @click="router.back()">返回</button>
      </div>

      <div v-if="loading" class="mt-6 h-40 animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>

      <div v-else-if="!data" class="mt-6 border border-[#1C1C1C]/10 p-6 text-[#1C1C1C]/40">未找到订单。</div>

      <div v-else class="mt-8 flex flex-col gap-6">
        <!-- 商品信息 -->
        <Card class="overflow-hidden">
          <div class="grid gap-6 md:grid-cols-[280px_1fr]">
            <div class="m-4 mr-0 md:m-6">
              <img :src="productCover" alt="" class="h-[200px] w-full object-cover md:h-[240px]" />
            </div>
            <div class="flex flex-col justify-center px-4 pb-4 md:py-6 md:pr-6">
              <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Product</p>
              <h2 class="mt-2 font-serif text-2xl tracking-tight">{{ product?.name || '商品加载中...' }}</h2>
              <p class="mt-2 text-2xl">{{ formatPrice(product?.price || 0) }}</p>
              <div class="mt-4 flex flex-wrap gap-4 text-sm text-[#1C1C1C]/40">
                <span>商品 ID: {{ order?.product_id }}</span>
                <span>库存: {{ product?.stock ?? '-' }}</span>
              </div>
            </div>
          </div>
        </Card>

        <div class="grid gap-6 md:grid-cols-2">
          <!-- 订单信息 -->
          <Card>
            <CardHeader>
              <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Order</p>
              <CardTitle class="font-serif text-lg tracking-tight">订单信息</CardTitle>
              <CardDescription class="text-[#1C1C1C]/40">订单状态与基础信息</CardDescription>
            </CardHeader>
            <CardContent class="space-y-3 text-sm">
              <div class="flex items-center justify-between">
                <span class="text-[#1C1C1C]/60">订单号</span>
                <span class="text-xs tracking-[0.12em] text-[#1C1C1C]/60">{{ order?.order_num }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-[#1C1C1C]/60">状态</span>
                <span class="border px-2 py-0.5 text-xs" :class="order?.status === 1 ? 'border-[#1C1C1C]/20 text-[#1C1C1C]/70' : order?.status === 2 || order?.status === 3 ? 'border-[#1C1C1C]/30 text-[#1C1C1C]/50' : 'border-[#1C1C1C]/20 text-[#1C1C1C]/70'">
                  {{ orderStatusText(order?.status) }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-[#1C1C1C]/60">创建时间</span>
                <span class="text-[#1C1C1C]/40">{{ order?.created_at }}</span>
              </div>
            </CardContent>
          </Card>

          <!-- 支付信息 -->
          <Card>
            <CardHeader>
              <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Payment</p>
              <CardTitle class="font-serif text-lg tracking-tight">支付信息</CardTitle>
              <CardDescription class="text-[#1C1C1C]/40">确认金额并完成支付</CardDescription>
            </CardHeader>
            <CardContent class="space-y-4 text-sm">
              <div class="flex items-center justify-between">
                <span class="text-[#1C1C1C]/60">支付单号</span>
                <span class="text-xs tracking-[0.12em] text-[#1C1C1C]/60">{{ payment?.payment_id || "-" }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-[#1C1C1C]/60">支付状态</span>
                <span class="text-[#1C1C1C]/70">{{ paymentStatusText(payment?.status) }}</span>
              </div>

              <div class="space-y-2 border border-[#1C1C1C]/10 p-4">
                <div class="flex items-center justify-between">
                  <span class="text-[#1C1C1C]/60">商品原价</span>
                  <span>{{ formatPrice(basePrice) }}</span>
                </div>
                <div class="space-y-2">
                  <div class="flex items-center justify-between text-xs text-[#1C1C1C]/40">
                    <span>选择优惠券</span>
                    <span>{{ couponsLoading ? "加载中..." : `可用 ${usableCoupons.length} 张` }}</span>
                  </div>
                  <select
                    v-model="selectedCouponId"
                    :disabled="!isPendingPayment || couponsLoading || displayCoupons.length === 0"
                    class="w-full border border-[#1C1C1C]/10 bg-transparent px-3 py-2 text-sm outline-none transition-colors focus:border-[#1C1C1C] disabled:opacity-50"
                  >
                    <option :value="null">不使用优惠券</option>
                    <option v-for="coupon in displayCoupons" :key="coupon.id" :value="coupon.id">
                      {{ coupon.title }} · {{ coupon.type === 'discount' ? coupon.discount_rate / 10 + '折' : '-' + formatPrice(coupon.amount_cents / 100) }}
                    </option>
                  </select>
                  <div class="flex items-center gap-2 text-xs">
                    <MagmaButton
                      class="flex-1 justify-center py-2"
                      :loading="applyingCoupon"
                      :disabled="!isPendingPayment || (!selectedCouponId && !currentCoupon)"
                      @click="applyCoupon"
                    >
                      应用
                    </MagmaButton>
                    <button
                      class="flex-1 border border-[#1C1C1C]/20 px-3 py-2 text-xs transition-colors hover:border-[#1C1C1C] disabled:opacity-50"
                      :disabled="!isPendingPayment || (!currentCoupon && !selectedCouponId) || applyingCoupon"
                      @click="() => { selectedCouponId = null; applyCoupon() }"
                    >
                      不使用优惠券
                    </button>
                  </div>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-[#1C1C1C]/60">应付金额</span>
                  <span class="text-2xl">{{ formatPrice(payableAmount) }}</span>
                </div>
                <div v-if="savedAmount > 0" class="flex items-center justify-between text-xs text-[#1C1C1C]/60">
                  <span>已优惠</span>
                  <span>-{{ formatPrice(savedAmount) }}</span>
                </div>
              </div>

              <div class="flex items-center gap-3 pt-2">
                <MagmaButton class="flex-1 justify-center" :loading="paying" :disabled="!isPendingPayment || paying" @click="pay('paid')">
                  确认支付
                </MagmaButton>
                <button
                  class="flex-1 border border-[#1C1C1C]/20 px-4 py-2 text-sm transition-colors hover:border-[#1C1C1C] disabled:opacity-50"
                  :disabled="!isPendingPayment || paying"
                  @click="pay('failed')"
                >
                  标记支付失败
                </button>
              </div>
              <p class="text-xs text-[#1C1C1C]/40">优惠券仅在待支付状态下可使用，支付后将自动发货。</p>
            </CardContent>
          </Card>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
