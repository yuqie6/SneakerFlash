<script setup lang="ts">
import { computed } from "vue"
import { RouterLink, RouterView, useRoute } from "vue-router"
import { useUserStore } from "@/stores/userStore"
import {
  LayoutDashboard, Users, ShoppingCart, Ticket, Package, ShieldAlert, ArrowLeft, ScrollText,
} from "lucide-vue-next"

const route = useRoute()
const userStore = useUserStore()
const nav = [
  { to: "/admin", label: "数据概览", icon: LayoutDashboard, exact: true, permission: "stats" },
  { to: "/admin/users", label: "用户管理", icon: Users, permission: "users" },
  { to: "/admin/orders", label: "订单管理", icon: ShoppingCart, permission: "orders" },
  { to: "/admin/coupons", label: "优惠券", icon: Ticket, permission: "coupons" },
  { to: "/admin/products", label: "商品管理", icon: Package, permission: "products" },
  { to: "/admin/risk", label: "风控管理", icon: ShieldAlert, permission: "risk" },
  { to: "/admin/audit", label: "审计日志", icon: ScrollText, permission: "audit" },
]
const visibleNav = computed(() => nav.filter((item) => !item.permission || userStore.hasPermission(item.permission)))
const isActive = (item: (typeof nav)[0]) => (item.exact ? route.path === item.to : route.path.startsWith(item.to) && (item.exact || item.to !== "/admin"))
</script>

<template>
  <div class="flex min-h-screen">
    <aside class="sticky top-0 flex h-screen w-52 shrink-0 flex-col bg-[#1C1C1C] text-white">
      <div class="border-b border-white/10 p-5">
        <p class="font-serif text-lg tracking-tight">Admin</p>
        <p class="mt-1 text-xs text-white/40">SneakerFlash</p>
      </div>
      <nav class="flex-1 space-y-0.5 p-3">
        <RouterLink
          v-for="item in visibleNav"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-3 py-2.5 text-sm transition-colors"
          :class="isActive(item) ? 'bg-white/10 text-white' : 'text-white/50 hover:bg-white/5 hover:text-white'"
        >
          <component :is="item.icon" class="h-4 w-4" />
          {{ item.label }}
        </RouterLink>
      </nav>
      <div class="border-t border-white/10 p-3">
        <RouterLink to="/" class="flex items-center gap-2 px-3 py-2 text-xs text-white/40 transition-colors hover:text-white">
          <ArrowLeft class="h-3 w-3" />
          返回主站
        </RouterLink>
      </div>
    </aside>
    <div class="flex flex-1 flex-col bg-[#F9F8F6]">
      <main class="flex-1 p-8">
        <RouterView />
      </main>
      <footer class="border-t border-[#1C1C1C]/10 px-8 py-6">
        <div class="flex items-center justify-between text-xs text-[#1C1C1C]/40">
          <span class="font-serif">SneakerFlash Admin</span>
          <span>管理后台</span>
        </div>
      </footer>
    </div>
  </div>
</template>
