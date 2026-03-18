<script setup lang="ts">
import { onMounted, reactive } from "vue"
import api from "@/lib/api"
import { getAdminErrorMessage } from "@/lib/admin"
import type { Order } from "@/types/order"
import MagmaButton from "@/components/motion/MagmaButton.vue"

const state = reactive({ items: [] as Order[], total: 0, page: 1, pageSize: 20, loading: false, error: "", status: "" })

const statusText = (s: number) => { switch (s) { case 0: return "待支付"; case 1: return "已支付"; case 2: return "失败"; default: return "未知" } }
const statusTone = (s: number) => { switch (s) { case 0: return "text-[#1C1C1C]/60"; case 1: return "text-[#1C1C1C]"; default: return "text-[#1C1C1C]/30" } }

const fetchOrders = async () => {
  state.loading = true
  state.error = ""
  try {
    const res = await api.get<{ list: Order[]; total: number }, { list: Order[]; total: number }>("/admin/orders", { params: { page: state.page, page_size: state.pageSize, status: state.status || undefined } })
    state.items = res.list || []
    state.total = res.total
  } catch (error) { state.error = getAdminErrorMessage(error) } finally { state.loading = false }
}

const onStatusChange = (v: string) => { state.page = 1; state.status = v; fetchOrders() }
const onPage = (d: number) => { state.page += d; fetchOrders() }
onMounted(fetchOrders)
</script>

<template>
  <div class="space-y-8">
    <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Orders</p>
        <h1 class="font-serif text-2xl tracking-tight md:text-3xl">订单管理</h1>
      </div>
      <div class="flex gap-4 text-sm">
        <button v-for="tab in [{ l: '全部', v: '' }, { l: '待支付', v: '0' }, { l: '已支付', v: '1' }, { l: '失败', v: '2' }]" :key="tab.v" class="hover-underline pb-0.5" :class="state.status === tab.v ? 'text-[#1C1C1C] font-medium' : 'text-[#1C1C1C]/40'" @click="onStatusChange(tab.v)">{{ tab.l }}</button>
      </div>
    </div>

    <div v-if="state.error" class="border border-dashed border-[#1C1C1C]/20 p-4 text-sm text-[#1C1C1C]/50">{{ state.error }}</div>

    <div v-if="state.loading" class="space-y-2">
      <div v-for="i in 5" :key="i" class="h-12 animate-pulse bg-[#1C1C1C]/5"></div>
    </div>

    <div v-else-if="state.items.length > 0" class="overflow-x-auto border border-[#1C1C1C]/10 bg-white">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[#1C1C1C]/10">
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">订单号</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">用户ID</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">商品ID</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">状态</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">时间</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#1C1C1C]/5">
          <tr v-for="o in state.items" :key="o.id" class="hover:bg-[#1C1C1C]/[0.02]">
            <td class="px-4 py-3 font-mono text-xs">{{ o.order_num }}</td>
            <td class="px-4 py-3 text-[#1C1C1C]/40">{{ o.user_id }}</td>
            <td class="px-4 py-3 text-[#1C1C1C]/40">{{ o.product_id }}</td>
            <td class="px-4 py-3"><span class="border border-[#1C1C1C]/10 px-2 py-0.5 text-xs" :class="statusTone(o.status)">{{ statusText(o.status) }}</span></td>
            <td class="px-4 py-3 text-[#1C1C1C]/40">{{ o.created_at }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else class="border border-[#1C1C1C]/10 p-8 text-center text-sm text-[#1C1C1C]/40">暂无数据</div>

    <div v-if="state.total > state.pageSize" class="flex items-center justify-between text-sm text-[#1C1C1C]/40">
      <span>第 {{ state.page }} 页 / 共 {{ Math.ceil(state.total / state.pageSize) }} 页</span>
      <div class="flex gap-3">
        <MagmaButton :disabled="state.page <= 1" @click="onPage(-1)">上一页</MagmaButton>
        <MagmaButton :disabled="state.page * state.pageSize >= state.total" @click="onPage(1)">下一页</MagmaButton>
      </div>
    </div>
  </div>
</template>
