<script setup lang="ts">
import { onMounted, reactive } from "vue"
import api from "@/lib/api"
import { getAdminErrorMessage } from "@/lib/admin"
import { formatPrice } from "@/lib/utils"
import type { AdminStats } from "@/types/admin"

const state = reactive({ loading: false, error: "", stats: null as AdminStats | null })

const cards: Array<{ label: string; key: keyof AdminStats; suffix?: string; isMoney?: boolean }> = [
  { label: "Total Users", key: "total_users", suffix: "人" },
  { label: "Total Orders", key: "total_orders", suffix: "笔" },
  { label: "Revenue", key: "total_revenue_cents", isMoney: true },
  { label: "Products", key: "total_products", suffix: "件" },
  { label: "Pending", key: "pending_orders", suffix: "笔" },
]

const fetchStats = async () => {
  state.loading = true
  state.error = ""
  try { state.stats = await api.get<AdminStats, AdminStats>("/admin/stats") }
  catch (error) { state.error = getAdminErrorMessage(error) }
  finally { state.loading = false }
}

onMounted(fetchStats)
</script>

<template>
  <div class="space-y-8">
    <div>
      <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Overview</p>
      <h1 class="font-serif text-2xl tracking-tight md:text-3xl">数据概览</h1>
    </div>

    <div v-if="state.error" class="border border-dashed border-[#1C1C1C]/20 p-4 text-sm text-[#1C1C1C]/50">{{ state.error }}</div>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
      <div v-for="card in cards" :key="card.key" class="border border-[#1C1C1C]/10 bg-white p-4">
        <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">{{ card.label }}</p>
        <div class="mt-2 text-2xl">
          <template v-if="state.loading || !state.stats">--</template>
          <template v-else-if="card.isMoney">{{ formatPrice(state.stats[card.key] / 100) }}</template>
          <template v-else>{{ state.stats[card.key] }} <span class="text-xs text-[#1C1C1C]/40">{{ card.suffix }}</span></template>
        </div>
      </div>
    </div>
  </div>
</template>
