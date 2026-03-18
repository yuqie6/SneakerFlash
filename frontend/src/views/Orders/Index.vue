<script setup lang="ts">
import { onMounted, reactive } from "vue"
import { RouterLink } from "vue-router"
import MainLayout from "@/layout/MainLayout.vue"
import { Card } from "@/components/ui/card"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import { useProductStore } from "@/stores/productStore"
import api, { resolveAssetUrl } from "@/lib/api"
import { formatPrice } from "@/lib/utils"
import type { Order } from "@/types/order"
import { toast } from "vue-sonner"

const productStore = useProductStore()

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
    const ids = [...new Set(res.list.map((o) => o.product_id).filter(Boolean))]
    await Promise.allSettled(ids.map((id) => productStore.fetchProductDetail(id)))
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
    case 0: return "待支付"
    case 1: return "已支付"
    case 2: return "支付失败"
    case 3: return "已取消"
    default: return "未知"
  }
}
const statusTone = (s: number) => {
  switch (s) {
    case 0: return "text-[#1C1C1C]/60"
    case 1: return "text-[#1C1C1C]"
    case 3: return "text-[#1C1C1C]/50"
    default: return "text-[#1C1C1C]/30"
  }
}
const productCover = (src?: string) => resolveAssetUrl(src) || "https://dummyimage.com/120x120/F9F8F6/1C1C1C&text=·"

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
              { label: '已取消', value: '3' },
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

      <div v-if="state.loading" class="grid gap-4 md:grid-cols-2">
        <div v-for="i in 4" :key="i" class="h-28 animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>
      </div>

      <div v-else-if="state.items.length === 0" class="border border-[#1C1C1C]/10 p-8 text-center text-[#1C1C1C]/40">
        暂无订单，去抢购或发布商品吧。
      </div>

      <div v-else class="grid gap-4 md:grid-cols-2">
        <Card v-for="order in state.items" :key="order.id" class="overflow-hidden hover:border-[#1C1C1C]/30">
          <RouterLink :to="`/orders/${order.id}`" class="flex gap-4 p-4">
            <img
              :src="productCover(productStore.detail(order.product_id)?.image)"
              :alt="productStore.detail(order.product_id)?.name || ''"
              class="h-20 w-20 shrink-0 border border-[#1C1C1C]/10 object-cover"
            />
            <div class="flex min-w-0 flex-1 flex-col justify-between">
              <div>
                <h3 class="truncate font-serif text-lg tracking-tight">{{ productStore.detail(order.product_id)?.name || "商品加载中..." }}</h3>
                <p class="mt-1 text-lg">{{ formatPrice(productStore.detail(order.product_id)?.price || 0) }}</p>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-xs text-[#1C1C1C]/40">{{ order.order_num }} · {{ order.created_at }}</span>
                <span class="shrink-0 border border-[#1C1C1C]/10 px-2 py-0.5 text-xs" :class="statusTone(order.status)">{{ statusText(order.status) }}</span>
              </div>
            </div>
          </RouterLink>
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
