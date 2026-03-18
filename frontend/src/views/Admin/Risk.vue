<script setup lang="ts">
import { onMounted, reactive } from "vue"
import api from "@/lib/api"
import { getAdminErrorMessage } from "@/lib/admin"
import type { RiskList } from "@/types/admin"
import { Input } from "@/components/ui/input"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import { toast } from "vue-sonner"

const black = reactive<RiskList>({ ip: [], user: [] })
const gray = reactive<RiskList>({ ip: [], user: [] })
const blackForm = reactive({ type: "ip" as "ip" | "user", value: "" })
const grayForm = reactive({ type: "ip" as "ip" | "user", value: "" })
const loading = reactive({ black: false, gray: false })
const error = reactive({ black: "", gray: "" })

const fetchBlacklist = async () => {
  loading.black = true; error.black = ""
  try {
    const res = await api.get<RiskList, RiskList>("/admin/risk/blacklist")
    black.ip = res.ip || []; black.user = res.user || []
  } catch (err) { error.black = getAdminErrorMessage(err) } finally { loading.black = false }
}
const fetchGraylist = async () => {
  loading.gray = true; error.gray = ""
  try {
    const res = await api.get<RiskList, RiskList>("/admin/risk/graylist")
    gray.ip = res.ip || []; gray.user = res.user || []
  } catch (err) { error.gray = getAdminErrorMessage(err) } finally { loading.gray = false }
}

const addBlack = async () => {
  if (!blackForm.value.trim()) return
  try { await api.post("/admin/risk/blacklist", { type: blackForm.type, value: blackForm.value.trim() }); blackForm.value = ""; toast.success("已添加"); fetchBlacklist() }
  catch {
    return
  }
}
const removeBlack = async (type: string, value: string) => {
  try { await api.delete("/admin/risk/blacklist", { data: { type, value } }); toast.success("已移除"); fetchBlacklist() }
  catch {
    return
  }
}
const addGray = async () => {
  if (!grayForm.value.trim()) return
  try { await api.post("/admin/risk/graylist", { type: grayForm.type, value: grayForm.value.trim() }); grayForm.value = ""; toast.success("已添加"); fetchGraylist() }
  catch {
    return
  }
}
const removeGray = async (type: string, value: string) => {
  try { await api.delete("/admin/risk/graylist", { data: { type, value } }); toast.success("已移除"); fetchGraylist() }
  catch {
    return
  }
}

onMounted(() => { fetchBlacklist(); fetchGraylist() })
</script>

<template>
  <div class="space-y-12">
    <div>
      <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Risk Control</p>
      <h1 class="font-serif text-2xl tracking-tight md:text-3xl">风控管理</h1>
    </div>

    <div class="grid gap-8 lg:grid-cols-2">
      <!-- Blacklist -->
      <div class="space-y-4">
        <h2 class="font-serif text-xl tracking-tight">黑名单</h2>
        <p class="text-xs text-[#1C1C1C]/40">命中后直接拒绝请求（Redis SET: risk:ip:black / risk:user:black）</p>
        <div v-if="error.black" class="border border-dashed border-[#1C1C1C]/20 p-3 text-xs text-[#1C1C1C]/50">{{ error.black }}</div>
        <div class="flex gap-2">
          <select v-model="blackForm.type" class="border border-[#1C1C1C]/10 bg-white px-3 py-2 text-sm">
            <option value="ip">IP</option>
            <option value="user">用户ID</option>
          </select>
          <Input v-model="blackForm.value" :placeholder="blackForm.type === 'ip' ? '192.168.1.1' : '用户ID'" class="flex-1" @keyup.enter="addBlack" />
          <MagmaButton @click="addBlack">添加</MagmaButton>
        </div>
        <div v-if="loading.black" class="h-20 animate-pulse bg-[#1C1C1C]/5"></div>
        <template v-else>
          <div v-for="section in [{ label: 'IP', items: black.ip, type: 'ip' }, { label: '用户', items: black.user, type: 'user' }]" :key="section.type" class="space-y-1">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">{{ section.label }}</p>
            <div v-if="section.items.length === 0" class="text-xs text-[#1C1C1C]/30">无</div>
            <div v-for="item in section.items" :key="item" class="flex items-center justify-between border-b border-[#1C1C1C]/5 py-2 text-sm">
              <span class="font-mono text-xs">{{ item }}</span>
              <button class="text-xs text-[#1C1C1C]/30 hover:text-[#1C1C1C]" @click="removeBlack(section.type, item)">移除</button>
            </div>
          </div>
        </template>
      </div>

      <!-- Graylist -->
      <div class="space-y-4">
        <h2 class="font-serif text-xl tracking-tight">灰名单</h2>
        <p class="text-xs text-[#1C1C1C]/40">命中后返回限流响应（Redis SET: risk:ip:gray / risk:user:gray）</p>
        <div v-if="error.gray" class="border border-dashed border-[#1C1C1C]/20 p-3 text-xs text-[#1C1C1C]/50">{{ error.gray }}</div>
        <div class="flex gap-2">
          <select v-model="grayForm.type" class="border border-[#1C1C1C]/10 bg-white px-3 py-2 text-sm">
            <option value="ip">IP</option>
            <option value="user">用户ID</option>
          </select>
          <Input v-model="grayForm.value" :placeholder="grayForm.type === 'ip' ? '192.168.1.1' : '用户ID'" class="flex-1" @keyup.enter="addGray" />
          <MagmaButton @click="addGray">添加</MagmaButton>
        </div>
        <div v-if="loading.gray" class="h-20 animate-pulse bg-[#1C1C1C]/5"></div>
        <template v-else>
          <div v-for="section in [{ label: 'IP', items: gray.ip, type: 'ip' }, { label: '用户', items: gray.user, type: 'user' }]" :key="section.type" class="space-y-1">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">{{ section.label }}</p>
            <div v-if="section.items.length === 0" class="text-xs text-[#1C1C1C]/30">无</div>
            <div v-for="item in section.items" :key="item" class="flex items-center justify-between border-b border-[#1C1C1C]/5 py-2 text-sm">
              <span class="font-mono text-xs">{{ item }}</span>
              <button class="text-xs text-[#1C1C1C]/30 hover:text-[#1C1C1C]" @click="removeGray(section.type, item)">移除</button>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
