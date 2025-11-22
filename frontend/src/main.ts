import { createApp } from "vue"
import { createPinia } from "pinia"
import Lenis from "lenis"
import App from "./App.vue"
import router from "./router"
import "./assets/css/index.css"
import "vue-sonner/style.css"

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount("#app")

const lenis = new Lenis({ duration: 1.1, smoothWheel: true })
const raf = (time: number) => {
  lenis.raf(time)
  requestAnimationFrame(raf)
}
requestAnimationFrame(raf)
