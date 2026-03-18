import { createRouter, createWebHistory, type RouterHistory } from "vue-router"
import Home from "@/views/Home/Index.vue"
import { useUserStore } from "@/stores/userStore"

const routes = [
  { path: "/", name: "home", component: Home },
  { path: "/login", name: "login", component: () => import("@/views/Auth/Login.vue") },
  { path: "/register", name: "register", component: () => import("@/views/Auth/Register.vue") },
  {
    path: "/product/:id",
    name: "product-detail",
    component: () => import("@/views/Product/Detail.vue"),
    props: true,
  },
  {
    path: "/products/publish",
    name: "product-publish",
    component: () => import("@/views/Product/Publish.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/profile",
    name: "profile",
    component: () => import("@/views/User/Profile.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/vip",
    name: "vip",
    component: () => import("@/views/User/VipCenter.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/orders",
    name: "orders",
    component: () => import("@/views/Orders/Index.vue"),
    meta: { requiresAuth: true },
  },
  {
    path: "/orders/:id",
    name: "order-detail",
    component: () => import("@/views/Orders/Detail.vue"),
    meta: { requiresAuth: true },
    props: true,
  },
] as const

export function createAppRouter(history: RouterHistory = createWebHistory()) {
  const router = createRouter({
    history,
    routes: [...routes],
  })

  router.beforeEach((to, _from, next) => {
    const userStore = useUserStore()
    if (to.meta.requiresAuth && !userStore.accessToken) {
      next({ name: "login", query: { redirect: to.fullPath } })
      return
    }
    next()
  })

  return router
}

const router = createAppRouter()

export default router
