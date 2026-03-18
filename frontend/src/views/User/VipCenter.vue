<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
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
  { id: 1, level: 3, duration: 30, priceCents: 3000, title: "L3 月卡", desc: "直升 VIP3 · 30 天，含月度券发放" },
  { id: 2, level: 4, duration: 90, priceCents: 8000, title: "L4 季卡", desc: "直升 VIP4 · 90 天，享最高等级权益" },
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
  return Math.max(0, Math.min(100, (spent / maxGrowthThreshold.value) * 100))
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
  switch (status) { case "available": return "可用"; case "used": return "已使用"; case "expired": return "已过期"; default: return status }
}
const statusTone = (status: CouponStatus) => {
  switch (status) {
    case "available": return "border-[#1C1C1C]/20 text-[#1C1C1C]/70"
    case "used": return "border-[#1C1C1C]/15 text-[#1C1C1C]/50"
    case "expired": return "border-[#1C1C1C]/10 text-[#1C1C1C]/30"
    default: return "border-[#1C1C1C]/10 text-[#1C1C1C]"
  }
}
const couponSourceLabel = (from?: string) => {
  switch (from) { case "vip_month": return "VIP 月度配额"; case "purchase": return "购买"; case "system": return "系统发放"; default: return from || "系统发放" }
}
const couponValueText = (coupon: Coupon) => {
  if (coupon.type === "discount") { return `${(coupon.discount_rate / 10).toFixed(1).replace(/\.0$/, "")}折` }
  return formatPrice(coupon.amount_cents / 100)
}
const couponRuleText = (coupon: Coupon) => {
  if (coupon.type === "discount") return coupon.min_spend_cents > 0 ? `满 ${formatPrice(coupon.min_spend_cents / 100)} 可用` : "无门槛"
  return `满 ${formatPrice(coupon.min_spend_cents / 100)} 减`
}

const fetchVipProfile = async () => {
  loadingProfile.value = true
  try { const res = await api.get<VIPProfile, VIPProfile>("/vip/profile"); vipProfile.value = res }
  catch (err: any) { toast.error(err?.message || "获取 VIP 信息失败") }
  finally { loadingProfile.value = false }
}
const fetchCoupons = async () => {
  loadingCoupons.value = true
  try {
    const res = await api.get<{ list: Coupon[]; total: number; page: number; page_size: number }, { list: Coupon[]; total: number; page: number; page_size: number }>("/coupons/mine", { params: { status: couponStatus.value || undefined, page: 1, page_size: 100 } })
    coupons.value = Array.isArray(res.list) ? res.list : []
  } catch (err: any) { toast.error(err?.message || "获取优惠券失败") }
  finally { loadingCoupons.value = false }
}
const purchaseVip = async (planId: number) => {
  buyingPlanId.value = planId
  try { const res = await api.post<VIPProfile, VIPProfile>("/vip/purchase", { plan_id: planId }); vipProfile.value = res; toast.success("付费 VIP 已生效"); await fetchCoupons() }
  catch (err: any) { toast.error(err?.message || "购买失败") }
  finally { buyingPlanId.value = null }
}
const purchaseCoupon = async () => {
  const id = Number(couponIdInput.value)
  if (!Number.isFinite(id) || id <= 0) { toast.error("请输入有效的券 ID"); return }
  purchasingCoupon.value = true
  try { await api.post<Coupon, Coupon>("/coupons/purchase", { coupon_id: id }); toast.success("优惠券已到账"); couponIdInput.value = ""; await fetchCoupons() }
  catch (err: any) { toast.error(err?.message || "购买优惠券失败") }
  finally { purchasingCoupon.value = false }
}
const formatDate = (val?: string) => { if (!val) return "--"; const d = new Date(val); if (Number.isNaN(d.getTime())) return val; return d.toLocaleDateString() }

watch(couponStatus, () => { fetchCoupons() })
onMounted(() => { if (!userStore.profile) userStore.fetchProfile(); fetchVipProfile(); fetchCoupons() })
</script>

<template>
  <MainLayout>
    <section class="mx-auto max-w-6xl px-6 py-16 md:py-24">
      <div class="flex flex-col gap-8">
        <!-- 标题 -->
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">VIP & Coupons</p>
            <h1 class="font-serif text-3xl tracking-tight md:text-5xl">权益中心</h1>
            <p class="mt-1 text-sm text-[#1C1C1C]/40">成长等级 + 付费 VIP，自动发券与购买体验。</p>
          </div>
          <div class="border border-[#1C1C1C]/10 px-6 py-4">
            <div class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Effective Level</div>
            <div class="mt-1 font-serif text-3xl tracking-tight">VIP L{{ effectiveLevel }}</div>
            <p class="mt-1 text-xs text-[#1C1C1C]/40">累计消费 {{ formatPrice(spentYuan) }}</p>
          </div>
        </div>

        <!-- 四宫格 -->
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Growth</p>
            <div class="mt-2 text-xl">VIP L{{ growthLevel }} <span class="text-xs text-[#1C1C1C]/40">永久</span></div>
            <p class="mt-1 text-xs text-[#1C1C1C]/30">按累计实付计算</p>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Paid VIP</p>
            <div class="mt-2 text-xl">{{ paidLevel > 0 ? `VIP L${paidLevel}` : "未开通" }}</div>
            <p class="mt-1 text-xs text-[#1C1C1C]/40">到期 {{ paidExpireText }}</p>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Effective</p>
            <div class="mt-2 text-xl">VIP L{{ effectiveLevel }}</div>
            <p class="mt-1 text-xs text-[#1C1C1C]/40">权益按更高等级生效</p>
          </div>
          <div class="border border-[#1C1C1C]/10 p-4">
            <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">Total Spent</p>
            <div class="mt-2 text-xl">{{ formatPrice(spentYuan) }}</div>
            <p class="mt-1 text-xs text-[#1C1C1C]/40">含已支付订单</p>
          </div>
        </div>

        <!-- 成长进度 + 付费套餐 -->
        <div class="grid gap-8 lg:grid-cols-[1.4fr_0.8fr]">
          <Card>
            <CardHeader>
              <CardTitle class="font-serif text-xl tracking-tight">成长进度</CardTitle>
              <CardDescription class="text-[#1C1C1C]/40">按累计实付成长，L4 以上无需再冲刺。</CardDescription>
            </CardHeader>
            <CardContent class="space-y-5">
              <div class="flex flex-wrap items-center justify-between gap-3 text-sm text-[#1C1C1C]/60">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="border border-[#1C1C1C]/20 bg-[#1C1C1C]/5 px-3 py-1 text-xs text-[#1C1C1C]/70">累计 {{ formatPrice(spentYuan) }}</span>
                  <span class="border border-[#1C1C1C]/10 px-3 py-1 text-xs text-[#1C1C1C]/40">当前 L{{ growthLevel }}</span>
                </div>
                <div class="text-xs text-[#1C1C1C]/40">
                  <template v-if="nextThreshold">距 L{{ growthLevel + 1 }} 还差 <span class="text-[#1C1C1C]/70">{{ formatPrice(gapToNext / 100) }}</span></template>
                  <span v-else>已达最高等级</span>
                </div>
              </div>

              <div class="border border-[#1C1C1C]/10 p-5">
                <div class="relative h-1 w-full bg-[#1C1C1C]/10">
                  <div class="h-full bg-[#1C1C1C] transition-[width] duration-500 ease-out" :style="`width: ${overallGrowthProgress}%`"></div>
                </div>
                <div class="mt-3 flex items-center justify-between text-xs text-[#1C1C1C]/40">
                  <span>总进度 {{ overallGrowthProgress.toFixed(1).replace(/\.0$/, "") }}%</span>
                  <span>最高 L{{ highestGrowthLevel }}</span>
                </div>
                <div class="mt-4 grid gap-2 text-xs sm:grid-cols-2 lg:grid-cols-4">
                  <div
                    v-for="milestone in growthMilestones"
                    :key="`card-${milestone.level}`"
                    class="border border-[#1C1C1C]/10 px-3 py-2 transition"
                    :class="milestone.isCurrent ? 'border-[#1C1C1C]/30' : ''"
                  >
                    <div class="flex items-center justify-between">
                      <span class="font-medium">L{{ milestone.level }}</span>
                      <span class="text-[11px]" :class="milestone.isCurrent ? 'text-[#1C1C1C]/70' : milestone.achieved ? 'text-[#1C1C1C]/60' : 'text-[#1C1C1C]/30'">
                        {{ milestone.isCurrent ? "当前" : milestone.achieved ? "已达成" : "未达成" }}
                      </span>
                    </div>
                    <p class="mt-1 text-[11px] text-[#1C1C1C]/40">≥ {{ formatPrice(milestone.min / 100) }}</p>
                    <p v-if="milestone.isCurrent && nextThreshold" class="mt-1 text-[11px] text-[#1C1C1C]/70">距 L{{ growthLevel + 1 }} {{ formatPrice(gapToNext / 100) }}</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle class="font-serif text-xl tracking-tight">付费 VIP 套餐</CardTitle>
              <CardDescription class="text-[#1C1C1C]/40">即刻直升，生效期间取成长等级与付费等级最大值。</CardDescription>
            </CardHeader>
            <CardContent class="space-y-3">
              <div v-for="plan in paidPlans" :key="plan.id" class="flex items-center justify-between border border-[#1C1C1C]/10 px-4 py-3">
                <div class="space-y-1">
                  <p class="text-sm text-[#1C1C1C]/60">{{ plan.title }}</p>
                  <p class="text-lg font-medium">VIP L{{ plan.level }} · {{ plan.duration }} 天</p>
                  <p class="text-xs text-[#1C1C1C]/40">{{ plan.desc }}</p>
                </div>
                <div class="flex flex-col items-end gap-2 text-right">
                  <div class="text-lg">{{ formatPrice(plan.priceCents / 100) }}</div>
                  <MagmaButton class="px-4 py-2" :loading="buyingPlanId === plan.id" :disabled="loadingProfile" @click="purchaseVip(plan.id)">购买</MagmaButton>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <!-- 优惠券 + 购买 -->
        <div class="grid gap-8 lg:grid-cols-[1.6fr_0.8fr]">
          <Card>
            <CardHeader class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
              <div>
                <CardTitle class="font-serif text-xl tracking-tight">我的优惠券</CardTitle>
                <CardDescription class="text-[#1C1C1C]/40">登录时自动发放当月 VIP 配额券，可按状态筛选。</CardDescription>
              </div>
              <div class="flex flex-wrap gap-3 text-sm">
                <button
                  v-for="tab in couponTabs"
                  :key="tab.label"
                  class="hover-underline pb-0.5 transition-colors"
                  :class="couponStatus === tab.value ? 'text-[#1C1C1C] font-medium' : 'text-[#1C1C1C]/40'"
                  @click="couponStatus = tab.value"
                >
                  {{ tab.label }}
                </button>
              </div>
            </CardHeader>
            <CardContent>
              <div v-if="loadingCoupons" class="space-y-3">
                <div v-for="i in 3" :key="i" class="h-20 animate-pulse border border-[#1C1C1C]/10 bg-[#1C1C1C]/5"></div>
              </div>
              <div v-else-if="coupons.length === 0" class="border border-[#1C1C1C]/10 p-6 text-sm text-[#1C1C1C]/40">暂无符合条件的优惠券。</div>
              <div v-else class="space-y-3">
                <div v-for="coupon in coupons" :key="coupon.id" class="border border-[#1C1C1C]/10 p-4 transition-colors hover:border-[#1C1C1C]/30">
                  <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <div class="flex items-center gap-2">
                        <h3 class="font-serif text-lg tracking-tight">{{ coupon.title }}</h3>
                        <span class="border border-[#1C1C1C]/10 px-2 py-0.5 text-xs text-[#1C1C1C]/40">{{ coupon.type === "discount" ? "折扣券" : "满减券" }}</span>
                      </div>
                      <p class="text-xs text-[#1C1C1C]/40">{{ coupon.description || "系统发放" }}</p>
                    </div>
                    <span class="border px-3 py-1 text-xs" :class="statusTone(coupon.status)">{{ statusLabel(coupon.status) }}</span>
                  </div>
                  <div class="mt-3 flex flex-wrap items-center gap-4 text-sm">
                    <div class="text-2xl">{{ couponValueText(coupon) }}</div>
                    <span class="text-[#1C1C1C]/60">{{ couponRuleText(coupon) }}</span>
                    <span class="border border-[#1C1C1C]/10 px-3 py-1 text-xs text-[#1C1C1C]/40">{{ formatDate(coupon.valid_from) }} - {{ formatDate(coupon.valid_to) }}</span>
                    <span class="border border-[#1C1C1C]/10 px-3 py-1 text-xs text-[#1C1C1C]/40">{{ couponSourceLabel(coupon.obtained_from) }}</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle class="font-serif text-xl tracking-tight">购买优惠券</CardTitle>
              <CardDescription class="text-[#1C1C1C]/40">支持后台标记可购买的券，模拟支付后直接入账。</CardDescription>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="space-y-2 text-sm">
                <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">券 ID</label>
                <input
                  v-model="couponIdInput"
                  type="number"
                  min="1"
                  placeholder="输入后台配置的券模板 ID"
                  class="w-full border border-[#1C1C1C]/10 bg-transparent px-3 py-2 text-sm outline-none transition-colors focus:border-[#1C1C1C]"
                />
                <p class="text-xs text-[#1C1C1C]/30">VIP 每月自动发券无需购买；仅对可购买券有效。</p>
              </div>
              <MagmaButton class="w-full justify-center" :loading="purchasingCoupon" @click="purchaseCoupon">购买优惠券</MagmaButton>
              <div class="border border-[#1C1C1C]/10 p-3 text-xs text-[#1C1C1C]/40">
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
