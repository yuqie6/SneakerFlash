<script setup lang="ts">
import { onMounted, ref, computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import api from "@/lib/api"
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
  } catch (err: any) {
    toast.error(err?.message || "获取订单失败")
  } finally {
    loading.value = false
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
  if (currentCoupon.value && !list.some((c) => c.id === currentCoupon.value.id)) {
    list.unshift(currentCoupon.value)
  }
  return list
})

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
    const res = await api.get<Coupon[], Coupon[]>("/coupons/mine", { params: { status: "available" } })
    coupons.value = Array.isArray(res) ? res : []
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
  } catch (err: any) {
    toast.error(err?.message || "优惠券应用失败")
  } finally {
    applyingCoupon.value = false
  }
}

const orderStatusText = (s?: number) => {
  switch (s) {
    case 0:
      return "待支付"
    case 1:
      return "已支付"
    case 2:
      return "支付失败"
    default:
      return "未知"
  }
}

const paymentStatusText = (s?: Payment["status"]) => {
  switch (s) {
    case "pending":
      return "待支付"
    case "paid":
      return "已支付"
    case "failed":
      return "失败"
    case "refunded":
      return "已退款"
    default:
      return "未知"
  }
}

onMounted(fetchDetail)
</script>

<template>
  <MainLayout>
    <section class="relative mx-auto max-w-4xl px-6 py-12">
      <div class="pointer-events-none absolute inset-0 opacity-60 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-10 right-0 h-80 w-80 rounded-full bg-[#ea580c55] blur-3xl"></div>
      </div>

      <div class="relative flex items-center justify-between">
        <div>
          <p class="text-sm uppercase tracking-[0.3em] text-magma">Order Detail</p>
          <h1 class="text-2xl font-semibold">订单详情</h1>
        </div>
        <button class="text-sm text-white/60 underline underline-offset-4" @click="router.back()">返回</button>
      </div>

      <div v-if="loading" class="mt-6 h-40 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>

      <div v-else-if="!data" class="mt-6 rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-6 text-white/70">
        未找到订单。
      </div>

      <div v-else class="relative mt-6 grid gap-6 md:grid-cols-2">
        <Card class="border-obsidian-border/70 bg-obsidian-card/80">
          <CardHeader>
            <CardTitle class="text-lg">订单信息</CardTitle>
            <CardDescription>订单状态与基础信息</CardDescription>
          </CardHeader>
          <CardContent class="space-y-3 text-sm text-white/80">
            <div class="flex items-center justify-between">
              <span>订单号</span>
              <span class="font-mono">{{ order?.order_num }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>状态</span>
              <span class="text-magma">{{ orderStatusText(order?.status) }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>商品 ID</span>
              <span>{{ order?.product_id }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>创建时间</span>
              <span class="text-white/60">{{ order?.created_at }}</span>
            </div>
          </CardContent>
        </Card>

        <Card class="border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card">
          <CardHeader>
            <CardTitle class="text-lg">支付信息</CardTitle>
            <CardDescription>演示支付回调，可模拟成功/失败。</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4 text-sm text-white/80">
            <div class="flex items-center justify-between">
              <span>支付单号</span>
              <span class="font-mono">{{ payment?.payment_id || "-" }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>支付状态</span>
              <span class="text-magma">{{ paymentStatusText(payment?.status) }}</span>
            </div>

            <div class="space-y-2 rounded-2xl border border-obsidian-border/70 bg-black/40 p-4">
              <div class="flex items-center justify-between">
                <span>商品原价</span>
                <span class="text-white">{{ formatPrice(basePrice) }}</span>
              </div>
              <div class="space-y-2">
                <div class="flex items-center justify-between text-xs text-white/60">
                  <span>选择优惠券</span>
                  <span>{{ couponsLoading ? "加载中..." : `可用 ${usableCoupons.length} 张` }}</span>
                </div>
                <select
                  v-model="selectedCouponId"
                  :disabled="!isPendingPayment || couponsLoading || displayCoupons.length === 0"
                  class="w-full rounded-lg border border-obsidian-border/70 bg-black/60 px-3 py-2 text-sm text-white outline-none transition focus:border-magma disabled:opacity-50"
                >
                  <option :value="null">不使用优惠券</option>
                  <option v-for="coupon in displayCoupons" :key="coupon.id" :value="coupon.id">
                    {{ coupon.title }} · {{ coupon.type === 'discount' ? coupon.discount_rate / 10 + '折' : '-' + formatPrice(coupon.amount_cents / 100) }}
                  </option>
                </select>
                <div class="flex items-center gap-2 text-xs text-white/60">
                  <MagmaButton
                    class="flex-1 justify-center"
                    size="sm"
                    :loading="applyingCoupon"
                    :disabled="!isPendingPayment || (!selectedCouponId && !currentCoupon)"
                    @click="applyCoupon"
                  >
                    应用
                  </MagmaButton>
                  <button
                    class="flex-1 rounded-full border border-obsidian-border px-3 py-2 text-xs text-white transition hover:border-magma hover:text-magma disabled:opacity-50"
                    :disabled="!isPendingPayment || (!currentCoupon && !selectedCouponId) || applyingCoupon"
                    @click="() => { selectedCouponId = null; applyCoupon() }"
                  >
                    不使用优惠券
                  </button>
                </div>
              </div>
              <div class="flex items-center justify-between">
                <span>应付金额</span>
                <span class="text-2xl font-semibold text-magma">{{ formatPrice(payableAmount) }}</span>
              </div>
              <div v-if="savedAmount > 0" class="flex items-center justify-between text-xs text-emerald-300">
                <span>已优惠</span>
                <span>-{{ formatPrice(savedAmount) }}</span>
              </div>
            </div>

            <div class="flex items-center gap-3 pt-2">
              <MagmaButton class="flex-1 justify-center" :loading="paying" :disabled="!isPendingPayment || paying" @click="pay('paid')">
                模拟支付成功
              </MagmaButton>
              <button
                class="flex-1 rounded-full border border-obsidian-border px-4 py-2 text-sm text-white transition hover:border-magma hover:text-magma disabled:opacity-50"
                :disabled="!isPendingPayment || paying"
                @click="pay('failed')"
              >
                模拟失败
              </button>
            </div>
            <p class="text-xs text-white/60">实际接入时替换为支付网关；回调已做幂等，优惠券仅在待支付时可应用。</p>
          </CardContent>
        </Card>
      </div>
    </section>
  </MainLayout>
</template>
