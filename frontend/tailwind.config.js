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
        "shimmer": "shimmer 1.5s infinite",
        "shake": "shake 0.35s ease-in-out",
      },
      keyframes: {
        shimmer: {
          "0%": { transform: "translateX(-100%)" },
          "100%": { transform: "translateX(100%)" },
        },
        shake: {
          "10%, 90%": { transform: "translateX(-2px)" },
          "20%, 80%": { transform: "translateX(3px)" },
          "30%, 50%, 70%": { transform: "translateX(-5px)" },
          "40%, 60%": { transform: "translateX(5px)" },
        },
      },
    },
  },
  plugins: [animate],
}
