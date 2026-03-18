<script setup lang="ts">
import { onMounted, reactive } from "vue"
import api from "@/lib/api"
import { getAdminErrorMessage } from "@/lib/admin"
import { formatPrice } from "@/lib/utils"
import type { AdminUser } from "@/types/admin"
import MagmaButton from "@/components/motion/MagmaButton.vue"

const state = reactive({ items: [] as AdminUser[], total: 0, page: 1, pageSize: 20, loading: false, error: "" })

const fetchUsers = async () => {
  state.loading = true
  state.error = ""
  try {
    const res = await api.get<{ list: AdminUser[]; total: number }, { list: AdminUser[]; total: number }>("/admin/users", { params: { page: state.page, page_size: state.pageSize } })
    state.items = res.list || []
    state.total = res.total
  } catch (error) { state.error = getAdminErrorMessage(error) } finally { state.loading = false }
}

const onPage = (d: number) => { state.page += d; fetchUsers() }
onMounted(fetchUsers)
</script>

<template>
  <div class="space-y-8">
    <div>
      <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Users</p>
      <h1 class="font-serif text-2xl tracking-tight md:text-3xl">用户管理</h1>
    </div>

    <div v-if="state.error" class="border border-dashed border-[#1C1C1C]/20 p-4 text-sm text-[#1C1C1C]/50">{{ state.error }}</div>

    <div v-if="state.loading" class="space-y-2">
      <div v-for="i in 5" :key="i" class="h-12 animate-pulse bg-[#1C1C1C]/5"></div>
    </div>

    <div v-else-if="state.items.length > 0" class="overflow-x-auto border border-[#1C1C1C]/10 bg-white">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[#1C1C1C]/10">
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">ID</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">用户名</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">余额</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">VIP</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">累计消费</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">注册时间</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#1C1C1C]/5">
          <tr v-for="u in state.items" :key="u.id" class="hover:bg-[#1C1C1C]/[0.02]">
            <td class="px-4 py-3 text-[#1C1C1C]/40">{{ u.id }}</td>
            <td class="px-4 py-3">{{ u.username }}</td>
            <td class="px-4 py-3">{{ formatPrice(u.balance) }}</td>
            <td class="px-4 py-3">L{{ u.growth_level }}</td>
            <td class="px-4 py-3">{{ formatPrice(u.total_spent_cents / 100) }}</td>
            <td class="px-4 py-3 text-[#1C1C1C]/40">{{ u.created_at }}</td>
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
