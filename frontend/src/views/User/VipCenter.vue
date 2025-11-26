<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { Crown, Ticket, Zap } from "lucide-vue-next"
import MainLayout from "@/layout/MainLayout.vue"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import api from "@/lib/api"
import { toast } from "vue-sonner"
import type { VIPProfile } from "@/types/vip"
import type { Coupon, CouponStatus } from "@/types/coupon"
import { formatPrice } from "@/lib/utils"
import { useUserStore } from "@/stores/userStore"

const userStore = useUserStore()

const vipProfile = ref<VIPProfile | null>(null)
const loadingProfile = ref(false)
const buyingPlanId = ref<number | null>(null)

const coupons = ref<Coupon[]>([])
const couponStatus = ref<CouponStatus | "">("available")
const loadingCoupons = ref(false)
const purchasingCoupon = ref(false)
const couponIdInput = ref("")

const growthThresholds = [
  { level: 1, min: 0 },
  { level: 2, min: 100_000 },
  { level: 3, min: 500_000 },
  { level: 4, min: 2_000_000 },
]

const paidPlans = [
  { id: 1, level: 3, duration: 30, priceCents: 3000, title: "L3 熔岩月卡", desc: "直升 VIP3 · 30 天，含月度券发放" },
  { id: 2, level: 4, duration: 90, priceCents: 8000, title: "L4 玄金季卡", desc: "直升 VIP4 · 90 天，享最高等级权益" },
]

const couponTabs: Array<{ label: string; value: CouponStatus | "" }> = [
  { label: "可用", value: "available" },
  { label: "已使用", value: "used" },
  { label: "已过期", value: "expired" },
  { label: "全部", value: "" },
]

const spentYuan = computed(() => (vipProfile.value?.total_spent_cents || 0) / 100)
const growthLevel = computed(() => vipProfile.value?.growth_level || 1)
const paidLevel = computed(() => vipProfile.value?.paid_level || 0)
const effectiveLevel = computed(() => vipProfile.value?.effective_level || growthLevel.value)
const paidExpireText = computed(() => {
  const raw = vipProfile.value?.paid_expired_at
  if (!raw) return "未开通"
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return raw
  return d.toLocaleString()
})

const nextThreshold = computed(() => growthThresholds.find((g) => g.level === growthLevel.value + 1)?.min ?? null)
const gapToNext = computed(() => {
  if (!nextThreshold.value) return 0
  const spent = vipProfile.value?.total_spent_cents || 0
  return Math.max(0, nextThreshold.value - spent)
})

const maxGrowthThreshold = computed(() => growthThresholds[growthThresholds.length - 1]?.min || 0)
const overallGrowthProgress = computed(() => {
  if (!maxGrowthThreshold.value) return 100
  const spent = vipProfile.value?.total_spent_cents || 0
  const progress = (spent / maxGrowthThreshold.value) * 100
  return Math.max(0, Math.min(100, progress))
})
const highestGrowthLevel = computed(() => growthThresholds[growthThresholds.length - 1]?.level ?? growthLevel.value)
const growthMilestones = computed(() => {
  const spent = vipProfile.value?.total_spent_cents || 0
  const max = maxGrowthThreshold.value || 1
  return growthThresholds.map((item) => ({
    ...item,
    percent: Math.min(100, (item.min / max) * 100),
    isCurrent: growthLevel.value === item.level,
    achieved: spent >= item.min,
  }))
})

const statusLabel = (status: CouponStatus) => {
  switch (status) {
    case "available":
      return "可用"
    case "used":
      return "已使用"
    case "expired":
      return "已过期"
    default:
      return status
  }
}

const statusTone = (status: CouponStatus) => {
  switch (status) {
    case "available":
      return "border-emerald-500/40 bg-emerald-500/15 text-emerald-200"
    case "used":
      return "border-amber-500/40 bg-amber-500/10 text-amber-200"
    case "expired":
      return "border-white/10 bg-white/5 text-white/50"
    default:
      return "border-white/10 bg-white/5 text-white"
  }
}

const couponSourceLabel = (from?: string) => {
  switch (from) {
    case "vip_month":
      return "VIP 月度配额"
    case "purchase":
      return "购买"
    case "system":
      return "系统发放"
    default:
      return from || "系统发放"
  }
}

const couponValueText = (coupon: Coupon) => {
  if (coupon.type === "discount") {
    const value = (coupon.discount_rate / 10).toFixed(1).replace(/\.0$/, "")
    return `${value}折`
  }
  return formatPrice(coupon.amount_cents / 100)
}

const couponRuleText = (coupon: Coupon) => {
  if (coupon.type === "discount") {
    return coupon.min_spend_cents > 0 ? `满 ${formatPrice(coupon.min_spend_cents / 100)} 可用` : "无门槛"
  }
  return `满 ${formatPrice(coupon.min_spend_cents / 100)} 减`
}

const fetchVipProfile = async () => {
  loadingProfile.value = true
  try {
    const res = await api.get<VIPProfile, VIPProfile>("/vip/profile")
    vipProfile.value = res
  } catch (err: any) {
    toast.error(err?.message || "获取 VIP 信息失败")
  } finally {
    loadingProfile.value = false
  }
}

const fetchCoupons = async () => {
  loadingCoupons.value = true
  try {
    const res = await api.get<Coupon[], Coupon[]>("/coupons/mine", {
      params: { status: couponStatus.value || undefined },
    })
    coupons.value = Array.isArray(res) ? res : []
  } catch (err: any) {
    toast.error(err?.message || "获取优惠券失败")
  } finally {
    loadingCoupons.value = false
  }
}

const purchaseVip = async (planId: number) => {
  buyingPlanId.value = planId
  try {
    const res = await api.post<VIPProfile, VIPProfile>("/vip/purchase", { plan_id: planId })
    vipProfile.value = res
    toast.success("付费 VIP 已生效")
    await fetchCoupons()
  } catch (err: any) {
    toast.error(err?.message || "购买失败")
  } finally {
    buyingPlanId.value = null
  }
}

const purchaseCoupon = async () => {
  const id = Number(couponIdInput.value)
  if (!Number.isFinite(id) || id <= 0) {
    toast.error("请输入有效的券 ID")
    return
  }
  purchasingCoupon.value = true
  try {
    await api.post<Coupon, Coupon>("/coupons/purchase", { coupon_id: id })
    toast.success("优惠券已到账")
    couponIdInput.value = ""
    await fetchCoupons()
  } catch (err: any) {
    toast.error(err?.message || "购买优惠券失败")
  } finally {
    purchasingCoupon.value = false
  }
}

const formatDate = (val?: string) => {
  if (!val) return "--"
  const d = new Date(val)
  if (Number.isNaN(d.getTime())) return val
  return d.toLocaleDateString()
}

watch(couponStatus, () => {
  fetchCoupons()
})

onMounted(() => {
  if (!userStore.profile) {
    userStore.fetchProfile()
  }
  fetchVipProfile()
  fetchCoupons()
})
</script>

<template>
  <MainLayout>
    <section class="relative mx-auto max-w-6xl px-6 py-12">
      <div class="pointer-events-none absolute inset-0 opacity-70 [mask-image:radial-gradient(ellipse_at_center,white,transparent)]">
        <div class="absolute -left-10 top-0 h-64 w-64 rounded-full bg-magma-glow blur-3xl"></div>
        <div class="absolute bottom-0 right-0 h-80 w-80 rounded-full bg-[#ea580c66] blur-3xl"></div>
      </div>

      <div class="relative flex flex-col gap-6">
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div class="space-y-2">
            <p class="flex items-center gap-2 text-sm uppercase tracking-[0.3em] text-magma">
              <Crown class="h-4 w-4" />
              VIP & Coupons
            </p>
            <h1 class="text-3xl font-semibold">权益中心</h1>
            <p class="text-sm text-white/70">成长等级 + 付费 VIP，自动发券与购买体验，保持黑金质感。</p>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 px-6 py-4 shadow-lg shadow-magma-glow/20">
            <div class="text-sm text-white/60">有效等级</div>
            <div class="mt-1 text-3xl font-semibold">VIP L{{ effectiveLevel }}</div>
            <p class="mt-1 text-xs text-white/60">
              累计消费 {{ formatPrice(spentYuan) }}
            </p>
          </div>
        </div>

        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="text-sm text-white/60">成长等级</p>
            <div class="mt-2 flex items-center gap-2 text-xl font-semibold text-white">
              VIP L{{ growthLevel }}
              <span class="rounded-full border border-obsidian-border px-2 py-0.5 text-xs text-white/70">永久</span>
            </div>
            <p class="mt-1 text-xs text-white/50">按累计实付计算</p>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="text-sm text-white/60">付费 VIP</p>
            <div class="mt-2 flex items-center gap-2 text-xl font-semibold text-white">
              {{ paidLevel > 0 ? `VIP L${paidLevel}` : "未开通" }}
            </div>
            <p class="mt-1 text-xs text-white/60">到期 {{ paidExpireText }}</p>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="text-sm text-white/60">有效等级</p>
            <div class="mt-2 text-xl font-semibold text-magma">VIP L{{ effectiveLevel }}</div>
            <p class="mt-1 text-xs text-white/60">权益按更高等级生效</p>
          </div>
          <div class="rounded-2xl border border-obsidian-border/70 bg-obsidian-card/80 p-4">
            <p class="text-sm text-white/60">累计实付</p>
            <div class="mt-2 text-xl font-semibold text-white">{{ formatPrice(spentYuan) }}</div>
            <p class="mt-1 text-xs text-white/60">含已支付订单</p>
          </div>
        </div>

        <div class="grid gap-6 lg:grid-cols-[1.4fr_0.8fr]">
          <Card class="border-obsidian-border/70 bg-gradient-to-br from-obsidian-card via-black to-obsidian-card">
            <CardHeader>
              <CardTitle class="flex items-center gap-2 text-xl">
                <Zap class="h-5 w-5 text-magma" />
                成长进度
              </CardTitle>
              <CardDescription>按累计实付成长，L4 以上无需再冲刺。</CardDescription>
            </CardHeader>
            <CardContent class="space-y-5">
              <div class="flex flex-wrap items-center justify-between gap-3 text-sm text-white/70">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="rounded-full border border-magma/50 bg-magma/10 px-3 py-1 text-xs text-magma">
                    累计 {{ formatPrice(spentYuan) }}
                  </span>
                  <span class="rounded-full border border-obsidian-border px-3 py-1 text-xs text-white/60">
                    当前 L{{ growthLevel }}
                  </span>
                </div>
                <div class="text-xs text-white/60">
                  <template v-if="nextThreshold">
                    距 L{{ growthLevel + 1 }} 还差 <span class="text-magma">{{ formatPrice(gapToNext / 100) }}</span>
                  </template>
                  <span v-else>已达最高等级</span>
                </div>
              </div>

              <div class="relative overflow-hidden rounded-2xl border border-obsidian-border/70 bg-gradient-to-r from-white/5 via-black/70 to-black/40 p-5">
                <div class="pointer-events-none absolute inset-0">
                  <div class="absolute left-10 top-0 h-full w-1/2 bg-magma-glow blur-3xl"></div>
                  <div class="absolute -right-8 -top-10 h-36 w-36 rounded-full bg-magma/25 blur-3xl"></div>
                </div>
                <div class="relative space-y-4">
                  <div class="relative h-3 w-full overflow-visible rounded-full bg-white/10">
                    <div
                      class="absolute inset-0 rounded-full bg-[radial-gradient(circle_at_10%_50%,rgba(249,115,22,0.18),transparent_45%),radial-gradient(circle_at_90%_50%,rgba(234,88,12,0.16),transparent_35%)]"
                    ></div>
                  <div
                    class="relative h-full rounded-full bg-gradient-to-r from-magma to-amber-200 shadow-[0_0_30px_rgba(234,88,12,0.35)] transition-[width] duration-500 ease-out"
                    :style="`width: ${overallGrowthProgress}%`"
                  >
                    <div class="absolute right-0 top-1/2 h-4 w-4 -translate-y-1/2 rounded-full border border-white/60 bg-white shadow-[0_0_15px_rgba(255,255,255,0.7)]"></div>
                  </div>
                  </div>
                  <div class="flex items-center justify-between text-xs text-white/60">
                    <span>总进度 {{ overallGrowthProgress.toFixed(1).replace(/\.0$/, "") }}%</span>
                    <span>最高 L{{ highestGrowthLevel }}</span>
                  </div>
                  <div class="grid gap-2 text-xs text-white/60 sm:grid-cols-2 lg:grid-cols-4">
                    <div
                      v-for="milestone in growthMilestones"
                      :key="`card-${milestone.level}`"
                      class="rounded-lg border border-obsidian-border/70 bg-black/30 px-3 py-2 transition"
                      :class="[
                        milestone.isCurrent
                          ? 'border-magma/80 shadow-[0_0_12px_rgba(249,115,22,0.35)]'
                          : milestone.achieved
                            ? 'border-obsidian-border text-white/70'
                            : 'border-obsidian-border/70 text-white/60'
                      ]"
                    >
                      <div class="flex items-center justify-between text-white">
                        <span class="font-semibold">L{{ milestone.level }}</span>
                        <span
                          class="rounded-full px-2 py-0.5 text-[11px]"
                          :class="milestone.isCurrent ? 'bg-magma/20 text-magma' : milestone.achieved ? 'bg-white/10 text-white/70' : 'bg-white/5 text-white/60'"
                        >
                          {{ milestone.isCurrent ? "当前" : milestone.achieved ? "已达成" : "未达成" }}
                        </span>
                      </div>
                      <p class="mt-1 text-[11px]">≥ {{ formatPrice(milestone.min / 100) }}</p>
                      <p v-if="milestone.isCurrent && nextThreshold" class="mt-1 text-[11px] text-magma">
                        距 L{{ growthLevel + 1 }} {{ formatPrice(gapToNext / 100) }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card class="border-obsidian-border/70 bg-obsidian-card/80">
            <CardHeader>
              <CardTitle class="text-xl">付费 VIP 套餐</CardTitle>
              <CardDescription>即刻直升，生效期间取成长等级与付费等级最大值。</CardDescription>
            </CardHeader>
            <CardContent class="space-y-3">
              <div
                v-for="plan in paidPlans"
                :key="plan.id"
                class="flex items-center justify-between rounded-xl border border-obsidian-border/60 bg-black/30 px-4 py-3"
              >
                <div class="space-y-1">
                  <p class="text-sm text-white/70">{{ plan.title }}</p>
                  <p class="text-lg font-semibold text-white">VIP L{{ plan.level }} · {{ plan.duration }} 天</p>
                  <p class="text-xs text-white/60">{{ plan.desc }}</p>
                </div>
                <div class="flex flex-col items-end gap-2 text-right">
                  <div class="text-lg font-semibold text-magma">{{ formatPrice(plan.priceCents / 100) }}</div>
                  <MagmaButton
                    class="px-4 py-2"
                    :loading="buyingPlanId === plan.id"
                    :disabled="loadingProfile"
                    @click="purchaseVip(plan.id)"
                  >
                    购买
                  </MagmaButton>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <div class="grid gap-6 lg:grid-cols-[1.6fr_0.8fr]">
          <Card class="border-obsidian-border/70 bg-obsidian-card/80">
            <CardHeader class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
              <div>
                <CardTitle class="flex items-center gap-2 text-xl">
                  <Ticket class="h-5 w-5 text-magma" />
                  我的优惠券
                </CardTitle>
                <CardDescription>登录时自动发放当月 VIP 配额券，可按状态筛选。</CardDescription>
              </div>
              <div class="flex flex-wrap gap-2 text-sm">
                <button
                  v-for="tab in couponTabs"
                  :key="tab.label"
                  class="rounded-full border px-3 py-1 transition"
                  :class="[
                    couponStatus === tab.value ? 'border-magma text-magma bg-magma/10' : 'border-obsidian-border text-white/70 hover:border-magma hover:text-magma',
                  ]"
                  @click="couponStatus = tab.value"
                >
                  {{ tab.label }}
                </button>
              </div>
            </CardHeader>
            <CardContent>
              <div v-if="loadingCoupons" class="space-y-3">
                <div v-for="i in 3" :key="i" class="h-20 animate-pulse rounded-xl border border-obsidian-border/70 bg-obsidian-card/80"></div>
              </div>
              <div v-else-if="coupons.length === 0" class="rounded-xl border border-obsidian-border/70 bg-black/30 p-6 text-sm text-white/70">
                暂无符合条件的优惠券。
              </div>
              <div v-else class="space-y-3">
                <div
                  v-for="coupon in coupons"
                  :key="coupon.id"
                  class="rounded-xl border border-obsidian-border/70 bg-black/30 p-4 transition hover:border-magma/60 hover:shadow-lg hover:shadow-magma-glow/10"
                >
                  <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <div class="flex items-center gap-2">
                        <h3 class="text-lg font-semibold">{{ coupon.title }}</h3>
                        <span class="rounded-full border px-2 py-0.5 text-xs text-white/70">
                          {{ coupon.type === "discount" ? "折扣券" : "满减券" }}
                        </span>
                      </div>
                      <p class="text-xs text-white/60">{{ coupon.description || "系统发放" }}</p>
                    </div>
                    <span class="rounded-full border px-3 py-1 text-xs" :class="statusTone(coupon.status)">
                      {{ statusLabel(coupon.status) }}
                    </span>
                  </div>
                  <div class="mt-3 flex flex-wrap items-center gap-4 text-sm">
                    <div class="text-2xl font-semibold text-magma">{{ couponValueText(coupon) }}</div>
                    <span class="text-white/70">{{ couponRuleText(coupon) }}</span>
                    <span class="rounded-full border border-obsidian-border px-3 py-1 text-xs text-white/70">
                      有效期 {{ formatDate(coupon.valid_from) }} - {{ formatDate(coupon.valid_to) }}
                    </span>
                    <span class="rounded-full border border-obsidian-border px-3 py-1 text-xs text-white/60">
                      来源 {{ couponSourceLabel(coupon.obtained_from) }}
                    </span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card class="border-obsidian-border/70 bg-gradient-to-b from-obsidian-card via-black to-obsidian-card">
            <CardHeader>
              <CardTitle class="text-xl">购买优惠券</CardTitle>
              <CardDescription>支持后台标记可购买的券，模拟支付后直接入账。</CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="space-y-2 text-sm">
                <label class="text-white/70">券 ID</label>
                <input
                  v-model="couponIdInput"
                  type="number"
                  min="1"
                  placeholder="输入后台配置的券模板 ID"
                  class="w-full rounded-lg border border-obsidian-border/60 bg-black/40 px-3 py-2 text-sm text-white outline-none transition focus:border-magma"
                />
                <p class="text-xs text-white/50">VIP 每月自动发券无需购买；仅对可购买券有效。</p>
              </div>
              <MagmaButton class="w-full justify-center" :loading="purchasingCoupon" @click="purchaseCoupon">
                购买优惠券
              </MagmaButton>
              <div class="rounded-xl border border-obsidian-border/60 bg-black/30 p-3 text-xs text-white/70">
                <p>提示：</p>
                <p>· 购买成功后自动刷新列表，可在下单时选择使用。</p>
                <p>· 当月 VIP 配额会在进入本页/登录时自动发放。</p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </section>
  </MainLayout>
</template>
