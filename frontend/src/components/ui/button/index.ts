import type { VariantProps } from "class-variance-authority"
import { cva } from "class-variance-authority"

export { default as Button } from "./Button.vue"

export const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap text-sm tracking-wide transition-colors focus-visible:outline-none active:scale-95 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "border border-[#1C1C1C] bg-[#1C1C1C] text-white hover:bg-[#1C1C1C]/90",
        accent: "border border-[#1C1C1C] bg-[#1C1C1C] text-white hover:bg-[#1C1C1C]/90",
        destructive: "border border-[#1C1C1C] bg-transparent text-[#1C1C1C] hover:border-[#1C1C1C] hover:bg-[#1C1C1C]/5",
        outline: "border border-[#1C1C1C]/20 bg-transparent text-[#1C1C1C] hover:border-[#1C1C1C] hover:bg-[#1C1C1C]/5",
        secondary: "bg-[#1C1C1C]/5 text-[#1C1C1C] hover:bg-[#1C1C1C]/10",
        ghost: "hover:bg-[#1C1C1C]/5 hover:text-[#1C1C1C]",
        link: "hover-underline pb-0.5 text-[#1C1C1C]",
      },
      size: {
        "default": "min-h-9 px-6 py-3",
        "xs": "min-h-7 px-3 py-1.5 text-xs",
        "sm": "min-h-8 px-4 py-2 text-xs",
        "lg": "min-h-10 px-8 py-3",
        "icon": "h-9 w-9",
        "icon-sm": "size-8",
        "icon-lg": "size-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)

export type ButtonVariants = VariantProps<typeof buttonVariants>
