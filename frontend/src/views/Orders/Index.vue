<script setup lang="ts">
import { onMounted, reactive } from "vue"
import { RouterLink } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import api from "@/lib/api"
import type { Order } from "@/types/order"
import { toast } from "vue-sonner"

const state = reactive({
  items: [] as Order[],
  total: 0,
  page: 1,
  size: 10,
  loading: false,
  status: "",
})

const fetchOrders = async () => {
  state.loading = true
  try {
    const res = await api.get<
      { items: Order[]; total: number; page: number; size: number },
      { items: Order[]; total: number; page: number; size: number }
    >("/orders", {
      params: {
        page: state.page,
        size: state.size,
        status: state.status || undefined,
      },
    })
    state.items = res.items
    state.total = res.total
  } catch (err: any) {
    toast.error(err?.message || "获取订单失败")
  } finally {
    state.loading = false
  }
}

const onStatusChange = (value: string) => {
  state.page = 1
  state.status = value
  fetchOrders()
}

const onPage = (delta: number) => {
  const next = state.page + delta
  if (next <= 0) return
  state.page = next
  fetchOrders()
}

const statusText = (s: number) => {
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

onMounted(fetchOrders)
</script>

<template>
  <MainLayout>
    <section class="relative mx-auto max-w-6xl px-6 py-12">
      <div class="pointer-events-none absolute inset-0 opacity-60 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-10 right-0 h-80 w-80 rounded-full bg-[#ea580c55] blur-3xl"></div>
      </div>
      <div class="relative mb-6 flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <p class="text-sm uppercase tracking-[0.3em] text-magma">Orders</p>
          <h1 class="text-2xl font-semibold">订单中心</h1>
          <p class="text-sm text-white/70">查看您的全部订单，点击订单可进入详情页完成支付。</p>
        </div>
        <div class="flex items-center gap-3 text-sm">
          <button
            class="rounded-full border border-obsidian-border px-3 py-1 text-white/80 transition hover:border-magma hover:text-magma"
            :class="{ 'border-magma text-magma': state.status === '' }"
            @click="onStatusChange('')"
          >
            全部
          </button>
          <button
            class="rounded-full border border-obsidian-border px-3 py-1 text-white/80 transition hover:border-magma hover:text-magma"
            :class="{ 'border-magma text-magma': state.status === '0' }"
            @click="onStatusChange('0')"
          >
            待支付
          </button>
          <button
            class="rounded-full border border-obsidian-border px-3 py-1 text-white/80 transition hover:border-magma hover:text-magma"
            :class="{ 'border-magma text-magma': state.status === '1' }"
            @click="onStatusChange('1')"
          >
            已支付
          </button>
          <button
            class="rounded-full border border-obsidian-border px-3 py-1 text-white/80 transition hover:border-magma hover:text-magma"
            :class="{ 'border-magma text-magma': state.status === '2' }"
            @click="onStatusChange('2')"
          >
            失败
          </button>
        </div>
      </div>

      <div v-if="state.loading" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <div v-for="i in 6" :key="i" class="h-36 animate-pulse rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
      </div>

      <div v-else-if="state.items.length === 0" class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-8 text-center text-white/70">
        暂无订单，去抢购或发布商品吧。
      </div>

      <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card v-for="order in state.items" :key="order.id" class="border-obsidian-border/70 bg-obsidian-card/80">
          <CardHeader class="space-y-2">
            <CardTitle class="text-lg">订单号：{{ order.order_num }}</CardTitle>
            <CardDescription class="text-white/70">状态：{{ statusText(order.status) }}</CardDescription>
          </CardHeader>
          <CardContent class="flex items-center justify-between">
            <RouterLink
              :to="`/orders/${order.id}`"
              class="text-sm text-magma underline decoration-magma/60 underline-offset-4"
            >
              查看详情 / 支付
            </RouterLink>
            <span class="text-xs text-white/50">{{ order.created_at }}</span>
          </CardContent>
        </Card>
      </div>

      <div v-if="state.items.length > 0" class="mt-6 flex items-center justify-between text-sm text-white/70">
        <span>第 {{ state.page }} 页 / 共 {{ Math.ceil(state.total / state.size) || 1 }} 页</span>
        <div class="flex gap-3">
          <MagmaButton :disabled="state.page <= 1 || state.loading" @click="onPage(-1)">上一页</MagmaButton>
          <MagmaButton :disabled="state.page * state.size >= state.total || state.loading" @click="onPage(1)">下一页</MagmaButton>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
