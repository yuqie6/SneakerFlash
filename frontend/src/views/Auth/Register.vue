<script setup lang="ts">
import { reactive } from "vue"
import { useRouter, RouterLink } from "vue-router"
import { useUserStore } from "@/stores/userStore"
import AuthLayout from "@/layout/AuthLayout.vue"
import MagmaButton from "@/components/motion/MagmaButton.vue"

const form = reactive({
  user_name: "",
  user_password: "",
})

const userStore = useUserStore()
const router = useRouter()

const onSubmit = async () => {
  await userStore.register(form)
  await userStore.login(form)
  router.push("/")
}
</script>

<template>
  <AuthLayout>
    <h1 class="mb-2 font-serif text-2xl tracking-tight">创建账号</h1>
    <p class="mb-8 text-sm text-[#1C1C1C]/60">加入 SneakerFlash，锁定下一场抢购席位。</p>
    <form class="space-y-5" @submit.prevent="onSubmit">
      <div class="space-y-2">
        <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">用户名</label>
        <input
          v-model="form.user_name"
          type="text"
          required
          class="w-full border-b border-[#1C1C1C]/10 bg-transparent px-0 py-3 text-sm outline-none transition-colors focus:border-[#1C1C1C] placeholder:text-[#1C1C1C]/30"
          placeholder="设置用户名"
        />
      </div>
      <div class="space-y-2">
        <label class="text-xs uppercase tracking-[0.2em] text-[#1C1C1C]/40">密码</label>
        <input
          v-model="form.user_password"
          type="password"
          required
          class="w-full border-b border-[#1C1C1C]/10 bg-transparent px-0 py-3 text-sm outline-none transition-colors focus:border-[#1C1C1C] placeholder:text-[#1C1C1C]/30"
          placeholder="设置密码"
        />
      </div>
      <MagmaButton class="w-full justify-center" :loading="userStore.loading">注册</MagmaButton>
      <p class="text-center text-sm text-[#1C1C1C]/40">
        已有账号？
        <RouterLink to="/login" class="hover-underline text-[#1C1C1C]">直接登录</RouterLink>
      </p>
    </form>
  </AuthLayout>
</template>
