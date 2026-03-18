import { createRouter, createWebHistory, type RouteRecordRaw, type RouterHistory } from "vue-router"
import Home from "@/views/Home/Index.vue"
import { useUserStore } from "@/stores/userStore"

const routes: RouteRecordRaw[] = [
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
  {
    path: "/admin",
    component: () => import("@/views/Admin/Layout.vue"),
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      { path: "", name: "admin-dashboard", component: () => import("@/views/Admin/Dashboard.vue") },
      { path: "users", name: "admin-users", component: () => import("@/views/Admin/Users.vue") },
      { path: "orders", name: "admin-orders", component: () => import("@/views/Admin/Orders.vue") },
      { path: "coupons", name: "admin-coupons", component: () => import("@/views/Admin/Coupons.vue") },
      { path: "products", name: "admin-products", component: () => import("@/views/Admin/Products.vue") },
      { path: "risk", name: "admin-risk", component: () => import("@/views/Admin/Risk.vue") },
    ],
  },
]

export function createAppRouter(history: RouterHistory = createWebHistory()) {
  const router = createRouter({
    history,
    routes,
  })

  router.beforeEach(async (to) => {
    const userStore = useUserStore()

    if (to.meta.requiresAuth && !userStore.accessToken) {
      return { name: "login", query: { redirect: to.fullPath } }
    }

    if (to.meta.requiresAdmin) {
      await userStore.fetchProfile()
      if (!userStore.profile) {
        return { name: "login", query: { redirect: to.fullPath } }
      }
      if (userStore.profile.role !== "admin") {
        return { name: "home" }
      }
    }

    return true
  })

  return router
}

const router = createAppRouter()

export default router
