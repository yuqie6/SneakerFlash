export interface Product {
  id: number
  user_id: number
  name: string
  price: number
  stock: number
  start_time: string
  end_time?: string  // 可选，结束时间
  image?: string
  created_at?: string
  updated_at?: string
}
