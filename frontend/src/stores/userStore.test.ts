import { createPinia, setActivePinia } from "pinia"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { useUserStore } from "@/stores/userStore"

const { apiMock } = vi.hoisted(() => ({
  apiMock: {
    post: vi.fn(),
    get: vi.fn(),
    put: vi.fn(),
  },
}))

vi.mock("@/lib/api", () => ({
  default: apiMock,
}))

vi.mock("vue-sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}))

describe("userStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it("persists tokens on login", async () => {
    apiMock.post.mockResolvedValueOnce({
      access_token: "access-token",
      refresh_token: "refresh-token",
      expires_in: 3600,
    })
    apiMock.get.mockResolvedValueOnce({
      id: 1,
      username: "alice",
    })

    const store = useUserStore()
    await store.login({ user_name: "alice", user_password: "password-123" })

    expect(store.accessToken).toBe("access-token")
    expect(store.refreshToken).toBe("refresh-token")
    expect(store.profile?.username).toBe("alice")
    expect(localStorage.getItem("access_token")).toBe("access-token")
  })

  it("clears auth state when fetchProfile fails", async () => {
    const store = useUserStore()
    store.setTokens("access-token", "refresh-token")
    apiMock.get.mockRejectedValueOnce(new Error("unauthorized"))

    await store.fetchProfile()

    expect(store.accessToken).toBe("")
    expect(store.refreshToken).toBe("")
    expect(store.profile).toBeNull()
  })

  it("updates profile data", async () => {
    const store = useUserStore()
    apiMock.put.mockResolvedValueOnce({
      id: 1,
      username: "alice-updated",
      avatar: "/uploads/a.png",
    })

    await store.updateProfile({ user_name: "alice-updated" })

    expect(store.profile?.username).toBe("alice-updated")
  })
})
