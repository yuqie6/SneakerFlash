import { createPinia, setActivePinia } from "pinia"
import { createMemoryHistory } from "vue-router"
import { beforeEach, describe, expect, it } from "vitest"
import { createAppRouter } from "@/router"
import { useUserStore } from "@/stores/userStore"

describe("router guards", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it("redirects unauthenticated users to login", async () => {
    const router = createAppRouter(createMemoryHistory())

    await router.push("/orders")
    await router.isReady()

    expect(router.currentRoute.value.name).toBe("login")
    expect(router.currentRoute.value.query.redirect).toBe("/orders")
  })

  it("allows authenticated users to protected routes", async () => {
    const router = createAppRouter(createMemoryHistory())
    const userStore = useUserStore()
    userStore.setTokens("access-token", "refresh-token")

    await router.push("/orders")
    await router.isReady()

    expect(router.currentRoute.value.name).toBe("orders")
  })
})
