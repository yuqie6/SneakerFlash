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
    <h1 class="mb-2 text-2xl font-semibold">创建账号</h1>
    <p class="mb-8 text-sm text-white/70">加入黑金赛道，提前锁定下一场抢购席位。</p>
    <form class="space-y-5" @submit.prevent="onSubmit">
      <div class="space-y-2">
        <label class="text-sm text-white/70">用户名</label>
        <input
          v-model="form.user_name"
          type="text"
          required
          class="w-full rounded-xl border border-obsidian-border/70 bg-obsidian-card px-4 py-3 text-sm text-white outline-none transition focus:border-magma focus:shadow-[0_0_0_4px_rgba(249,115,22,0.25)]"
          placeholder="设置用户名"
        />
      </div>
      <div class="space-y-2">
        <label class="text-sm text-white/70">密码</label>
        <input
          v-model="form.user_password"
          type="password"
          required
          class="w-full rounded-xl border border-obsidian-border/70 bg-obsidian-card px-4 py-3 text-sm text-white outline-none transition focus:border-magma focus:shadow-[0_0_0_4px_rgba(249,115,22,0.25)]"
          placeholder="设置密码"
        />
      </div>
      <MagmaButton class="w-full justify-center" :loading="userStore.loading">立即注册</MagmaButton>
      <p class="text-center text-sm text-white/60">
        已有账号？
        <RouterLink to="/login" class="text-magma hover:text-magma-dark">直接登录</RouterLink>
      </p>
    </form>
  </AuthLayout>
</template>
