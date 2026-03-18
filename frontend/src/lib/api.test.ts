import axios from "axios"
import MockAdapter from "axios-mock-adapter"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import api, { resolveAssetUrl } from "@/lib/api"

vi.mock("vue-sonner", () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}))

describe("api client", () => {
  let apiMock: MockAdapter
  let axiosMock: MockAdapter

  beforeEach(() => {
    apiMock = new MockAdapter(api)
    axiosMock = new MockAdapter(axios)
    localStorage.clear()
  })

  afterEach(() => {
    apiMock.restore()
    axiosMock.restore()
  })

  it("unwraps business success payload", async () => {
    apiMock.onGet("/products").reply(200, {
      code: 200,
      msg: "ok",
      data: { list: [{ id: 1 }] },
    })

    await expect(api.get("/products")).resolves.toEqual({ list: [{ id: 1 }] })
  })

  it("rejects business error payload", async () => {
    apiMock.onGet("/products").reply(200, {
      code: 30001,
      msg: "售罄",
    })

    await expect(api.get("/products")).rejects.toThrow("售罄")
  })

  it("refreshes token and retries on 401", async () => {
    localStorage.setItem("access_token", "expired-token")
    localStorage.setItem("refresh_token", "refresh-token")

    apiMock.onGet("/profile").replyOnce(401, { msg: "expired" })
    apiMock.onGet("/profile").reply(200, {
      code: 200,
      msg: "ok",
      data: { id: 1, username: "alice" },
    })
    axiosMock.onPost("/api/v1/refresh").reply(200, {
      code: 200,
      msg: "ok",
      data: {
        access_token: "new-access-token",
      },
    })

    await expect(api.get("/profile")).resolves.toEqual({ id: 1, username: "alice" })
    expect(localStorage.getItem("access_token")).toBe("new-access-token")
    expect(axiosMock.history.post).toHaveLength(1)
  })

  it("queues concurrent 401 requests behind one refresh", async () => {
    localStorage.setItem("access_token", "expired-token")
    localStorage.setItem("refresh_token", "refresh-token")

    apiMock.onGet("/orders").replyOnce(401, { msg: "expired" })
    apiMock.onGet("/orders").replyOnce(401, { msg: "expired" })
    apiMock.onGet("/orders").reply(200, {
      code: 200,
      msg: "ok",
      data: { list: [] },
    })
    axiosMock.onPost("/api/v1/refresh").reply(200, {
      code: 200,
      msg: "ok",
      data: {
        access_token: "refreshed-token",
      },
    })

    const first = api.get("/orders")
    const second = api.get("/orders")

    await expect(Promise.all([first, second])).resolves.toEqual([{ list: [] }, { list: [] }])
    expect(axiosMock.history.post).toHaveLength(1)
    expect(localStorage.getItem("access_token")).toBe("refreshed-token")
  })

  it("clears tokens when refresh fails", async () => {
    localStorage.setItem("access_token", "expired-token")
    localStorage.setItem("refresh_token", "refresh-token")

    apiMock.onGet("/orders").replyOnce(401, { msg: "expired" })
    axiosMock.onPost("/api/v1/refresh").reply(401, { msg: "refresh expired" })

    await expect(api.get("/orders")).rejects.toBeTruthy()
    expect(localStorage.getItem("access_token")).toBeNull()
    expect(localStorage.getItem("refresh_token")).toBeNull()
  })

  it("stops retrying when refreshed request is still 401", async () => {
    localStorage.setItem("access_token", "expired-token")
    localStorage.setItem("refresh_token", "refresh-token")

    apiMock.onGet("/profile").replyOnce(401, { msg: "expired" })
    apiMock.onGet("/profile").replyOnce(401, { msg: "still expired" })
    axiosMock.onPost("/api/v1/refresh").reply(200, {
      code: 200,
      msg: "ok",
      data: {
        access_token: "new-access-token",
      },
    })

    await expect(api.get("/profile")).rejects.toBeTruthy()
    expect(axiosMock.history.post).toHaveLength(1)
    expect(localStorage.getItem("access_token")).toBeNull()
    expect(localStorage.getItem("refresh_token")).toBeNull()
  })
})

describe("resolveAssetUrl", () => {
  it("returns empty string for empty src", () => {
    expect(resolveAssetUrl("")).toBe("")
    expect(resolveAssetUrl(null)).toBe("")
  })

  it("keeps absolute and root-relative urls", () => {
    expect(resolveAssetUrl("https://example.com/a.png")).toBe("https://example.com/a.png")
    expect(resolveAssetUrl("/uploads/a.png")).toBe("/uploads/a.png")
    expect(resolveAssetUrl("uploads/a.png")).toBe("/uploads/a.png")
  })
})
