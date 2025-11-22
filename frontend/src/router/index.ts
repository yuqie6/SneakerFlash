import { createRouter, createWebHistory } from "vue-router"
import Home from "@/views/Home/Index.vue"
import Login from "@/views/Auth/Login.vue"
import Register from "@/views/Auth/Register.vue"
import ProductDetail from "@/views/Product/Detail.vue"
import { useUserStore } from "@/stores/userStore"

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: Home },
    { path: "/login", name: "login", component: Login },
    { path: "/register", name: "register", component: Register },
    { path: "/product/:id", name: "product-detail", component: ProductDetail, props: true },
  ],
})

router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth && !userStore.token) {
    next({ name: "login", query: { redirect: to.fullPath } })
    return
  }
  next()
})

export default router
