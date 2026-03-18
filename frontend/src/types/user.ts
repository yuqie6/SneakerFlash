export interface User {
  id: number
  username: string
  balance: number
  avatar?: string
  total_spent_cents?: number
  growth_level?: number
  role?: string
  created_at?: string
  updated_at?: string
}
