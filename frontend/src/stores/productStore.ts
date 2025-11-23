import { defineStore } from "pinia"
import api from "@/lib/api"
import type { Product } from "@/types/product"

type ProductListResponse = {
  items: Product[]
  total: number
  page: number
}

export const useProductStore = defineStore("product", {
  state: () => ({
    items: [] as Product[],
    total: 0,
    loading: false,
    detailMap: new Map<number, Product>(),
    myItems: [] as Product[],
    myTotal: 0,
  }),
  getters: {
    detail: (state) => (id: number) => state.detailMap.get(id),
  },
  actions: {
    async fetchProducts(page = 1, size = 10, append = false) {
      this.loading = true
      try {
        const res = await api.get<ProductListResponse, ProductListResponse>("/products", { params: { page, size } })
        const list = Array.isArray(res?.items) ? res.items : []
        const total = Number(res?.total) || list.length
        this.items = append ? [...this.items, ...list] : list
        this.total = total
      } finally {
        this.loading = false
      }
    },
    async fetchProductDetail(id: number, refresh = false): Promise<Product> {
      const cached = this.detailMap.get(id)
      if (cached && !refresh) return cached
      const res = await api.get<ProductDetailResponse, Product>(`/product/${id}`)
      this.detailMap.set(id, res)
      return res
    },
    updateProduct(product: Product) {
      this.detailMap.set(product.id, product)
      const idx = this.items.findIndex((p) => p.id === product.id)
      if (idx >= 0) this.items[idx] = product
    },
    async fetchMyProducts(page = 1, size = 10) {
      this.loading = true
      try {
        const res = await api.get<{ items: Product[]; total: number; page: number }, { items: Product[]; total: number; page: number }>(
          "/products/mine",
          { params: { page, size } }
        )
        this.myItems = res.items || []
        this.myTotal = Number(res.total) || this.myItems.length
      } finally {
        this.loading = false
      }
    },
    async updateProductRemote(id: number, payload: Partial<Product>) {
      await api.put(`/products/${id}`, payload)
      await this.fetchMyProducts()
      await this.fetchProducts(1, 12)
    },
    async deleteProduct(id: number) {
      await api.delete(`/products/${id}`)
      await this.fetchMyProducts()
      await this.fetchProducts(1, 12)
    },
  },
})
