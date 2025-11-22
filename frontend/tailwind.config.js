import { fontFamily } from "tailwindcss/defaultTheme"
import animate from "tailwindcss-animate"

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter", ...fontFamily.sans],
      },
      colors: {
        magma: {
          DEFAULT: "#f97316",
          glow: "#f9731680",
          dark: "#ea580c",
        },
        obsidian: {
          bg: "#050505",
          card: "#0a0a0a",
          border: "#27272a",
        },
      },
      backgroundImage: {
        "magma-gradient": "linear-gradient(135deg, #f97316 0%, #ea580c 100%)",
      },
      animation: {
        "pulse-fast": "pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite",
      },
    },
  },
  plugins: [animate],
}
