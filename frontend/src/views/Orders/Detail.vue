<script setup lang="ts">
import { onMounted, ref, computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import api from "@/lib/api"
import type { Order, OrderWithPayment } from "@/types/order"
import type { Payment } from "@/types/payment"
import { formatPrice } from "@/lib/utils"
import { toast } from "vue-sonner"

const route = useRoute()
const router = useRouter()
const data = ref<OrderWithPayment | null>(null)
const loading = ref(false)
const paying = ref(false)

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await api.get<OrderWithPayment, OrderWithPayment>(`/orders/${route.params.id}`)
    data.value = res
  } catch (err: any) {
    toast.error(err?.message || "获取订单失败")
  } finally {
    loading.value = false
  }
}

const payment = computed<Payment | undefined>(() => data.value?.payment)
const order = computed<Order | undefined>(() => data.value?.order)

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
          <CardContent class="space-y-3 text-sm text-white/80">
            <div class="flex items-center justify-between">
              <span>支付单号</span>
              <span class="font-mono">{{ payment?.payment_id || "-" }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>金额</span>
              <span class="text-magma font-semibold">
                {{ payment ? formatPrice(payment.amount_cents / 100) : "-" }}
              </span>
            </div>
            <div class="flex items-center justify-between">
              <span>状态</span>
              <span class="text-magma">{{ paymentStatusText(payment?.status) }}</span>
            </div>
            <div class="flex items-center gap-3 pt-2">
              <MagmaButton class="flex-1 justify-center" :loading="paying" :disabled="!payment || payment.status === 'paid'" @click="pay('paid')">
                模拟支付成功
              </MagmaButton>
              <button
                class="flex-1 rounded-full border border-obsidian-border px-4 py-2 text-sm text-white transition hover:border-magma hover:text-magma disabled:opacity-50"
                :disabled="!payment || payment.status === 'paid' || paying"
                @click="pay('failed')"
              >
                模拟失败
              </button>
            </div>
            <p class="text-xs text-white/60">实际接入时替换为支付网关；回调已做幂等。</p>
          </CardContent>
        </Card>
      </div>
    </section>
  </MainLayout>
</template>
