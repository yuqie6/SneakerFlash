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
      { path: "", name: "admin-dashboard", component: () => import("@/views/Admin/Dashboard.vue"), meta: { requiresPermission: "stats" } },
      { path: "users", name: "admin-users", component: () => import("@/views/Admin/Users.vue"), meta: { requiresPermission: "users" } },
      { path: "orders", name: "admin-orders", component: () => import("@/views/Admin/Orders.vue"), meta: { requiresPermission: "orders" } },
      { path: "coupons", name: "admin-coupons", component: () => import("@/views/Admin/Coupons.vue"), meta: { requiresPermission: "coupons" } },
      { path: "products", name: "admin-products", component: () => import("@/views/Admin/Products.vue"), meta: { requiresPermission: "products" } },
      { path: "risk", name: "admin-risk", component: () => import("@/views/Admin/Risk.vue"), meta: { requiresPermission: "risk" } },
      { path: "audit", name: "admin-audit", component: () => import("@/views/Admin/Audit.vue"), meta: { requiresPermission: "audit" } },
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
    const ensureProfile = async () => {
      if (!userStore.profile && userStore.accessToken) {
        await userStore.fetchProfile()
      }
    }

    if (to.meta.requiresAuth && !userStore.accessToken) {
      return { name: "login", query: { redirect: to.fullPath } }
    }

    if (to.meta.requiresAdmin) {
      await ensureProfile()
      if (!userStore.profile) {
        return { name: "login", query: { redirect: to.fullPath } }
      }
      if (!userStore.isAdmin) {
        return { name: "home" }
      }
    }

    if (typeof to.meta.requiresPermission === "string") {
      await ensureProfile()
      if (!userStore.hasPermission(to.meta.requiresPermission)) {
        const permissionRoutePairs: Array<[string, string]> = [
          ["stats", "admin-dashboard"],
          ["users", "admin-users"],
          ["orders", "admin-orders"],
          ["products", "admin-products"],
          ["coupons", "admin-coupons"],
          ["risk", "admin-risk"],
          ["audit", "admin-audit"],
        ]
        const fallback = permissionRoutePairs.find(([permission]) => userStore.hasPermission(permission))
        return fallback ? { name: fallback[1] as string } : { name: "home" }
      }
    }

    return true
  })

  return router
}

const router = createAppRouter()

export default router
