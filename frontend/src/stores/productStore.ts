import { defineStore } from "pinia"
import api from "@/lib/api"
import type { Product } from "@/types/product"

type ProductListResponse = {
  data: Product[]
  total: number
  page: number
}

type ProductDetailResponse = {
  data: Product
}

export const useProductStore = defineStore("product", {
  state: () => ({
    items: [] as Product[],
    total: 0,
    loading: false,
    detailMap: new Map<number, Product>(),
  }),
  getters: {
    detail: (state) => (id: number) => state.detailMap.get(id),
  },
  actions: {
    async fetchProducts(page = 1, size = 10, append = false) {
      this.loading = true
      try {
        const res = await api.get<ProductListResponse, ProductListResponse>("/products", { params: { page, size } })
        this.items = append ? [...this.items, ...res.data] : res.data
        this.total = res.total
      } finally {
        this.loading = false
      }
    },
    async fetchProductDetail(id: number, refresh = false): Promise<Product> {
      const cached = this.detailMap.get(id)
      if (cached && !refresh) return cached
      const res = await api.get<ProductDetailResponse, ProductDetailResponse>(`/product/${id}`)
      this.detailMap.set(id, res.data)
      return res.data
    },
    updateProduct(product: Product) {
      this.detailMap.set(product.id, product)
      const idx = this.items.findIndex((p) => p.id === product.id)
      if (idx >= 0) this.items[idx] = product
    },
  },
})
