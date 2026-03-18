<script setup lang="ts">
import { onMounted, reactive, ref } from "vue"
import api from "@/lib/api"
import { getAdminErrorMessage } from "@/lib/admin"
import { formatPrice } from "@/lib/utils"
import type { CouponTemplate } from "@/types/admin"
import MagmaButton from "@/components/motion/MagmaButton.vue"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { toast } from "vue-sonner"

const state = reactive({ items: [] as CouponTemplate[], total: 0, page: 1, pageSize: 20, loading: false, error: "" })
const showForm = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)

const emptyForm = () => ({
  type: "full_cut" as "full_cut" | "discount",
  title: "", description: "",
  amount_cents: 0, discount_rate: 90, min_spend_cents: 0,
  valid_from: "", valid_to: "",
  purchasable: false, price_cents: 0, status: "active",
})
const form = reactive(emptyForm())

const fetchCoupons = async () => {
  state.loading = true
  state.error = ""
  try {
    const res = await api.get<{ list: CouponTemplate[]; total: number }, { list: CouponTemplate[]; total: number }>("/admin/coupons", { params: { page: state.page, page_size: state.pageSize } })
    state.items = res.list || []
    state.total = res.total
  } catch (error) { state.error = getAdminErrorMessage(error) } finally { state.loading = false }
}

const resetForm = () => { Object.assign(form, emptyForm()); editingId.value = null; showForm.value = false }
const startEdit = (c: CouponTemplate) => {
  editingId.value = c.id
  Object.assign(form, { type: c.type, title: c.title, description: c.description, amount_cents: c.amount_cents, discount_rate: c.discount_rate, min_spend_cents: c.min_spend_cents, valid_from: c.valid_from?.slice(0, 16) || "", valid_to: c.valid_to?.slice(0, 16) || "", purchasable: c.purchasable, price_cents: c.price_cents, status: c.status })
  showForm.value = true
}

const submitForm = async () => {
  saving.value = true
  try {
    if (editingId.value) {
      await api.put(`/admin/coupons/${editingId.value}`, form)
      toast.success("更新成功")
    } else {
      await api.post("/admin/coupons", form)
      toast.success("创建成功")
    }
    resetForm()
    fetchCoupons()
  } catch {}
  finally { saving.value = false }
}

const deleteCoupon = async (id: number) => {
  try { await api.delete(`/admin/coupons/${id}`); toast.success("已删除"); fetchCoupons() }
  catch {}
}

const typeText = (t: string) => t === "full_cut" ? "满减" : "折扣"
const onPage = (d: number) => { state.page += d; fetchCoupons() }
onMounted(fetchCoupons)
</script>

<template>
  <div class="space-y-8">
    <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
      <div>
        <p class="text-xs uppercase tracking-[0.3em] text-[#1C1C1C]/40">Coupons</p>
        <h1 class="font-serif text-2xl tracking-tight md:text-3xl">优惠券模板</h1>
      </div>
      <MagmaButton @click="showForm = !showForm; if (!showForm) resetForm()">
        {{ showForm ? "取消" : "创建券模板" }}
      </MagmaButton>
    </div>

    <div v-if="state.error" class="border border-dashed border-[#1C1C1C]/20 p-4 text-sm text-[#1C1C1C]/50">{{ state.error }}</div>

    <div v-if="showForm" class="space-y-4 border border-[#1C1C1C]/10 bg-white p-6">
      <p class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">{{ editingId ? "编辑券模板" : "新建券模板" }}</p>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">标题</label>
          <Input v-model="form.title" placeholder="满100减20" />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">类型</label>
          <select v-model="form.type" class="w-full border border-[#1C1C1C]/10 bg-white px-3 py-2 text-sm">
            <option value="full_cut">满减</option>
            <option value="discount">折扣</option>
          </select>
        </div>
        <div class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">描述</label>
          <Input v-model="form.description" placeholder="可选" />
        </div>
        <div v-if="form.type === 'full_cut'" class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">减免金额（分）</label>
          <Input v-model.number="form.amount_cents" type="number" />
        </div>
        <div v-if="form.type === 'discount'" class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">折扣率（90=九折）</label>
          <Input v-model.number="form.discount_rate" type="number" />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">使用门槛（分）</label>
          <Input v-model.number="form.min_spend_cents" type="number" />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">开始时间</label>
          <Input v-model="form.valid_from" type="datetime-local" />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">结束时间</label>
          <Input v-model="form.valid_to" type="datetime-local" />
        </div>
        <div class="space-y-1">
          <label class="text-xs text-[#1C1C1C]/40">售价（分，0=免费）</label>
          <Input v-model.number="form.price_cents" type="number" />
        </div>
        <div class="flex items-end gap-4">
          <label class="flex items-center gap-2 text-sm">
            <input v-model="form.purchasable" type="checkbox" class="accent-[#1C1C1C]" />
            可购买
          </label>
          <select v-model="form.status" class="border border-[#1C1C1C]/10 bg-white px-3 py-2 text-sm">
            <option value="active">启用</option>
            <option value="inactive">停用</option>
          </select>
        </div>
      </div>
      <div class="flex gap-3 border-t border-[#1C1C1C]/10 pt-4">
        <MagmaButton :disabled="saving" @click="submitForm">{{ saving ? "保存中..." : "保存" }}</MagmaButton>
        <Button variant="outline" @click="resetForm">取消</Button>
      </div>
    </div>

    <div v-if="state.loading" class="space-y-2">
      <div v-for="i in 5" :key="i" class="h-12 animate-pulse bg-[#1C1C1C]/5"></div>
    </div>

    <div v-else-if="state.items.length > 0" class="overflow-x-auto border border-[#1C1C1C]/10 bg-white">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[#1C1C1C]/10">
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">标题</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">类型</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">优惠</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">门槛</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">可购买</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">状态</th>
            <th class="px-4 py-3 text-left text-xs font-normal uppercase tracking-[0.2em] text-[#1C1C1C]/40">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#1C1C1C]/5">
          <tr v-for="c in state.items" :key="c.id" class="hover:bg-[#1C1C1C]/[0.02]">
            <td class="px-4 py-3">{{ c.title }}</td>
            <td class="px-4 py-3 text-[#1C1C1C]/60">{{ typeText(c.type) }}</td>
            <td class="px-4 py-3">{{ c.type === "full_cut" ? `减${formatPrice(c.amount_cents / 100)}` : `${c.discount_rate}折` }}</td>
            <td class="px-4 py-3 text-[#1C1C1C]/40">{{ c.min_spend_cents > 0 ? `满${formatPrice(c.min_spend_cents / 100)}` : "无" }}</td>
            <td class="px-4 py-3">{{ c.purchasable ? "是" : "否" }}</td>
            <td class="px-4 py-3"><span class="border border-[#1C1C1C]/10 px-2 py-0.5 text-xs" :class="c.status === 'active' ? 'text-[#1C1C1C]' : 'text-[#1C1C1C]/30'">{{ c.status === "active" ? "启用" : "停用" }}</span></td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <button class="hover-underline text-xs text-[#1C1C1C]/60 hover:text-[#1C1C1C]" @click="startEdit(c)">编辑</button>
                <button class="hover-underline text-xs text-[#1C1C1C]/30 hover:text-[#1C1C1C]" @click="deleteCoupon(c.id)">删除</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else-if="!state.loading" class="border border-[#1C1C1C]/10 p-8 text-center text-sm text-[#1C1C1C]/40">暂无券模板</div>

    <div v-if="state.total > state.pageSize" class="flex items-center justify-between text-sm text-[#1C1C1C]/40">
      <span>第 {{ state.page }} 页 / 共 {{ Math.ceil(state.total / state.pageSize) }} 页</span>
      <div class="flex gap-3">
        <MagmaButton :disabled="state.page <= 1" @click="onPage(-1)">上一页</MagmaButton>
        <MagmaButton :disabled="state.page * state.pageSize >= state.total" @click="onPage(1)">下一页</MagmaButton>
      </div>
    </div>
  </div>
</template>
