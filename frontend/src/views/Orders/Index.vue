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
  pageSize: 10,
  loading: false,
  status: "",
})

const fetchOrders = async () => {
  state.loading = true
  try {
    const res = await api.get<
      { list: Order[]; total: number; page: number; page_size: number },
      { list: Order[]; total: number; page: number; page_size: number }
    >("/orders", {
      params: {
        page: state.page,
        page_size: state.pageSize,
        status: state.status || undefined,
      },
    })
    state.items = res.list
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
    <section class="mx-auto max-w-6xl px-6 py-16 md:py-24">
      <div class="mb-8 flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Orders</p>
          <h1 class="font-serif text-2xl tracking-tight md:text-3xl">订单中心</h1>
          <p class="mt-1 text-sm text-[#1C1C1C]/40">查看您的全部订单，点击订单可进入详情页完成支付。</p>
        </div>
        <div class="flex items-center gap-4 text-sm">
          <button
            v-for="tab in [
              { label: '全部', value: '' },
              { label: '待支付', value: '0' },
              { label: '已支付', value: '1' },
              { label: '失败', value: '2' },
            ]"
            :key="tab.value"
            class="hover-underline pb-0.5 transition-colors"
            :class="state.status === tab.value ? 'text-[#1C1C1C] font-medium' : 'text-[#1C1C1C]/40'"
            @click="onStatusChange(tab.value)"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>

      <div v-if="state.loading" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <div v-for="i in 6" :key="i" class="h-36 animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>
      </div>

      <div v-else-if="state.items.length === 0" class="border border-[#1C1C1C]/10 p-8 text-center text-[#1C1C1C]/40">
        暂无订单，去抢购或发布商品吧。
      </div>

      <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card v-for="order in state.items" :key="order.id" class="hover:border-[#1C1C1C]/30">
          <CardHeader class="space-y-2">
            <CardTitle class="font-serif text-lg tracking-tight">{{ order.order_num }}</CardTitle>
            <CardDescription class="text-[#1C1C1C]/40">状态：{{ statusText(order.status) }}</CardDescription>
          </CardHeader>
          <CardContent class="flex items-center justify-between">
            <RouterLink
              :to="`/orders/${order.id}`"
              class="hover-underline text-sm text-[#1C1C1C]"
            >
              查看详情 / 支付
            </RouterLink>
            <span class="text-xs text-[#1C1C1C]/40">{{ order.created_at }}</span>
          </CardContent>
        </Card>
      </div>

      <div v-if="state.items.length > 0" class="mt-6 flex items-center justify-between text-sm text-[#1C1C1C]/40">
        <span>第 {{ state.page }} 页 / 共 {{ Math.ceil(state.total / state.pageSize) || 1 }} 页</span>
        <div class="flex gap-3">
          <MagmaButton :disabled="state.page <= 1 || state.loading" @click="onPage(-1)">上一页</MagmaButton>
          <MagmaButton :disabled="state.page * state.pageSize >= state.total || state.loading" @click="onPage(1)">下一页</MagmaButton>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
