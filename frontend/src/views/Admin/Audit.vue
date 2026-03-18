<script setup lang="ts">
import { onMounted, reactive } from "vue"
import api from "@/lib/api"
import { getAdminErrorMessage } from "@/lib/admin"
import type { AuditLog } from "@/types/admin"
import MagmaButton from "@/components/motion/MagmaButton.vue"

const state = reactive({
  items: [] as AuditLog[],
  total: 0,
  page: 1,
  pageSize: 20,
  loading: false,
  error: "",
  resource: "",
  action: "",
  actorName: "",
})

const fetchLogs = async () => {
  state.loading = true
  state.error = ""
  try {
    const res = await api.get<
      { list: AuditLog[]; total: number },
      { list: AuditLog[]; total: number }
    >("/admin/audit", {
      params: {
        page: state.page,
        page_size: state.pageSize,
        resource: state.resource || undefined,
        action: state.action || undefined,
        actor_name: state.actorName || undefined,
      },
    })
    state.items = res.list || []
    state.total = res.total
  } catch (error) {
    state.error = getAdminErrorMessage(error)
  } finally {
    state.loading = false
  }
}

const applyFilters = () => {
  state.page = 1
  fetchLogs()
}

const onPage = (delta: number) => {
  const next = state.page + delta
  if (next <= 0) return
  state.page = next
  fetchLogs()
}

onMounted(fetchLogs)
</script>

<template>
  <div class="space-y-8">
    <div>
      <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Audit</p>
      <h1 class="font-serif text-2xl tracking-tight md:text-3xl">审计日志</h1>
    </div>

    <div class="grid gap-3 border border-[#1C1C1C]/10 bg-white p-4 md:grid-cols-4">
      <input v-model="state.actorName" type="text" placeholder="操作者" class="border border-[#1C1C1C]/10 bg-transparent px-3 py-2 text-sm outline-none focus:border-[#1C1C1C]" />
      <select v-model="state.resource" class="border border-[#1C1C1C]/10 bg-transparent px-3 py-2 text-sm outline-none focus:border-[#1C1C1C]">
        <option value="">全部资源</option>
        <option value="coupons">coupons</option>
        <option value="risk">risk</option>
        <option value="audit">audit</option>
      </select>
      <select v-model="state.action" class="border border-[#1C1C1C]/10 bg-transparent px-3 py-2 text-sm outline-none focus:border-[#1C1C1C]">
        <option value="">全部动作</option>
        <option value="create">create</option>
        <option value="update">update</option>
        <option value="delete">delete</option>
      </select>
      <MagmaButton class="justify-center" :loading="state.loading" @click="applyFilters">筛选</MagmaButton>
    </div>

    <div v-if="state.error" class="border border-dashed border-[#1C1C1C]/20 p-4 text-sm text-[#1C1C1C]/50">{{ state.error }}</div>

    <div v-if="state.loading" class="space-y-2">
      <div v-for="i in 5" :key="i" class="h-14 animate-pulse bg-[#1C1C1C]/5"></div>
    </div>

    <div v-else-if="state.items.length > 0" class="overflow-x-auto border border-[#1C1C1C]/10 bg-white">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[#1C1C1C]/10">
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">时间</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">操作者</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">资源</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">动作</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">结果</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">请求</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#1C1C1C]/5">
          <tr v-for="log in state.items" :key="log.id" class="align-top hover:bg-[#1C1C1C]/[0.02]">
            <td class="px-4 py-3 text-[#1C1C1C]/40">{{ log.created_at }}</td>
            <td class="px-4 py-3">
              <div>{{ log.actor_name }}</div>
              <div class="text-xs text-[#1C1C1C]/40">{{ log.actor_role }}</div>
            </td>
            <td class="px-4 py-3">{{ log.resource }}<span v-if="log.resource_id" class="text-[#1C1C1C]/40"> #{{ log.resource_id }}</span></td>
            <td class="px-4 py-3">{{ log.action }}</td>
            <td class="px-4 py-3">
              <span class="border px-2 py-0.5 text-xs" :class="log.result === 'success' ? 'border-[#1C1C1C]/10 text-[#1C1C1C]' : 'border-[#1C1C1C]/20 text-[#1C1C1C]/50'">
                {{ log.result }}
              </span>
            </td>
            <td class="px-4 py-3 text-xs text-[#1C1C1C]/50">
              <div>{{ log.request_path }}</div>
              <div>{{ log.request_ip }}</div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else class="border border-[#1C1C1C]/10 p-8 text-center text-sm text-[#1C1C1C]/40">暂无审计日志</div>

    <div v-if="state.total > state.pageSize" class="flex items-center justify-between text-sm text-[#1C1C1C]/40">
      <span>第 {{ state.page }} 页 / 共 {{ Math.ceil(state.total / state.pageSize) }} 页</span>
      <div class="flex gap-3">
        <MagmaButton :disabled="state.page <= 1" @click="onPage(-1)">上一页</MagmaButton>
        <MagmaButton :disabled="state.page * state.pageSize >= state.total" @click="onPage(1)">下一页</MagmaButton>
      </div>
    </div>
  </div>
</template>
