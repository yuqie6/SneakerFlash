import { fontFamily } from "tailwindcss/defaultTheme"
import animate from "tailwindcss-animate"

/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        serif: ['"Playfair Display"', ...fontFamily.serif],
        sans: ["Inter", ...fontFamily.sans],
      },
      colors: {
        editorial: {
          bg: "#F9F8F6",
          card: "#FFFFFF",
          border: "rgba(28,28,28,0.10)",
          text: "#1C1C1C",
        },
      },
      animation: {
        shake: "shake 0.35s ease-in-out",
      },
      keyframes: {
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
