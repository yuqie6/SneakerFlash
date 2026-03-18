import { createPinia, setActivePinia } from "pinia"
import { beforeEach, describe, expect, it, vi } from "vitest"
import { useProductStore } from "@/stores/productStore"

const { apiMock } = vi.hoisted(() => ({
  apiMock: {
    get: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

vi.mock("@/lib/api", () => ({
  default: apiMock,
}))

describe("productStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it("fetches product list", async () => {
    apiMock.get.mockResolvedValueOnce({
      list: [{ id: 1, name: "AJ 1" }],
      total: 1,
      page: 1,
      page_size: 10,
    })

    const store = useProductStore()
    await store.fetchProducts()

    expect(store.items).toEqual([{ id: 1, name: "AJ 1" }])
    expect(store.total).toBe(1)
  })

  it("returns cached product detail when refresh is false", async () => {
    const store = useProductStore()
    store.detailMap.set(1, { id: 1, name: "cached" } as never)

    const detail = await store.fetchProductDetail(1)

    expect(detail).toEqual({ id: 1, name: "cached" })
    expect(apiMock.get).not.toHaveBeenCalled()
  })

  it("refreshes product lists after delete", async () => {
    apiMock.delete.mockResolvedValueOnce({})
    apiMock.get
      .mockResolvedValueOnce({ list: [], total: 0, page: 1, page_size: 10 })
      .mockResolvedValueOnce({ list: [], total: 0, page: 1, page_size: 12 })

    const store = useProductStore()
    await store.deleteProduct(1)

    expect(apiMock.delete).toHaveBeenCalledWith("/products/1")
    expect(apiMock.get).toHaveBeenCalledTimes(2)
  })
})
