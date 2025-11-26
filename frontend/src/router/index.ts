import { createRouter, createWebHistory } from "vue-router"
import Home from "@/views/Home/Index.vue"
import Login from "@/views/Auth/Login.vue"
import Register from "@/views/Auth/Register.vue"
import ProductDetail from "@/views/Product/Detail.vue"
import ProductPublish from "@/views/Product/Publish.vue"
import Profile from "@/views/User/Profile.vue"
import VipCenter from "@/views/User/VipCenter.vue"
import Orders from "@/views/Orders/Index.vue"
import OrderDetail from "@/views/Orders/Detail.vue"
import { useUserStore } from "@/stores/userStore"

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", name: "home", component: Home },
    { path: "/login", name: "login", component: Login },
    { path: "/register", name: "register", component: Register },
    { path: "/product/:id", name: "product-detail", component: ProductDetail, props: true },
    { path: "/products/publish", name: "product-publish", component: ProductPublish, meta: { requiresAuth: true } },
    { path: "/profile", name: "profile", component: Profile, meta: { requiresAuth: true } },
    { path: "/vip", name: "vip", component: VipCenter, meta: { requiresAuth: true } },
    { path: "/orders", name: "orders", component: Orders, meta: { requiresAuth: true } },
    { path: "/orders/:id", name: "order-detail", component: OrderDetail, meta: { requiresAuth: true }, props: true },
  ],
})

router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth && !userStore.accessToken) {
    next({ name: "login", query: { redirect: to.fullPath } })
    return
  }
  next()
})

export default router
