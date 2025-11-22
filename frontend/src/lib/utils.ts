import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatPrice(price: number) {
  if (Number.isNaN(price)) return "¥0.00"
  return price.toLocaleString("zh-CN", { style: "currency", currency: "CNY" })
}
