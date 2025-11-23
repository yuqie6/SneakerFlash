# SneakerFlash 前端执行方案（Midnight Magma Ver.）

## 目标与约束
- 目标：打造暗夜黑金风格的高并发秒杀/下单前端，交互流畅、状态透明（库存、下单、支付状态）。
- 约束：对接当前后端接口（`docs/backend-api.md`），覆盖注册/登录/刷新、商品列表与详情、发布/管理商品、秒杀、订单与支付状态查看。遵循 KISS/YAGNI：仅落地已有接口，不预留未实现功能。

## 技术栈与初始化
- Vue 3 + TypeScript + Vite 5（Script Setup）
- Tailwind CSS v3 + Shadcn-vue（New York / Zinc），Motion-v，Lenis
- 状态：Pinia；请求：Axios；工具：@vueuse/core、lucide-vue-next、date-fns、clsx/tailwind-merge/cva、vue-sonner
- 初始化命令（在项目根执行）：
  1) `npm create vite@latest frontend -- --template vue-ts`
  2) `cd frontend`
  3) `npm install -D tailwindcss postcss autoprefixer && npx tailwindcss init -p`
  4) `npm install axios pinia vue-router motion-v @vueuse/core lucide-vue-next clsx tailwind-merge class-variance-authority date-fns vue-sonner`
  5) `npx shadcn-vue@latest init`（New York, Zinc）

## 目录结构
```
src/
├─ assets/css/        # index.css (Tailwind directives + 纹理 + 变量)
├─ components/
│  ├─ ui/             # Shadcn 基础组件
│  └─ motion/         # MagmaButton, ParallaxCard 等
├─ composables/       # useSeckill, useCountDown, useAuthGuard
├─ layout/            # MainLayout, AuthLayout
├─ lib/               # api.ts (Axios 拦截), utils.ts (cn/formatPrice)
├─ stores/            # userStore, productStore
├─ types/             # user.ts, product.ts, order.ts
├─ views/             # Home/, Product/, Auth/
└─ router/index.ts
```

## 设计系统（暗夜黑金）
- Tailwind：`obsidian` 背景、`magma` 主色、`magma-gradient` 背景、`pulse-fast` 动画，字体建议 Inter。
- 全局样式：噪点纹理、`.glass` 工具类、`:root/.dark` 变量兼容 Shadcn；`body` 使用 `bg-obsidian-bg text-white`.
- 组件：Button/Input/Card/Progress/Dialog/Toast Provider 通过 Shadcn 生成；自定义 `MagmaButton` 使用 shimmer + pulse。

## API 对接要点
- BaseURL：`http://localhost:8000/api/v1`
- 鉴权：`Authorization: Bearer <access_token>`；`refresh_token` 备用。存储键建议 `sf_access_token` / `sf_refresh_token`
- 接口（核心）：
  - 认证：`POST /register`，`POST /login`（返回 access/refresh/expires_in），`POST /refresh`
  - 用户：`GET /profile`，`PUT /profile`
  - 上传：`POST /upload`（multipart）
  - 商品：`GET /products`，`GET /product/:id`，`POST /products`，`PUT /products/:id`，`DELETE /products/:id`，`GET /products/mine`
  - 秒杀：`POST /seckill {product_id}`
  - 订单：`POST /orders {product_id}`，`GET /orders`，`GET /orders/:id`（含支付单）
- 响应包装：后端统一 `{code,msg,data}`；Axios 拦截器处理 `code!=200`、401 自动跳登录，必要时用 refresh_token 重试一次，再失败清理态并跳转。toast 显示业务错误与限流/黑名单提示。

## 核心业务流
- `useSeckill`：状态机 `idle/loading/success/failed`；调用 `/seckill`；成功展示 `order_num`；按业务码提示售罄/重复/限流。
- 下单/支付：详情页支持 `/orders` 下单，拿到 `payment_id/status/amount_cents`；显示“待支付/已支付/失败”态，支付成功后刷新库存与订单列表。
- 倒计时：`useCountDown(start_time)` 返回剩余秒、`isStarted`；未开始按钮禁用显示 `MM:SS`。
- 库存/订单轮询：详情页可选 3-5s 轮询 `GET /product/:id`；订单详情可 5-10s 轮询支付状态（pending→paid/failed）。
- 路由守卫：访问需要登录的操作（详情页抢购、下单）缺 token 时重定向 `/login`。

## 开发里程碑
1) 工程初始化 + Tailwind/Shadcn 配置 + 基础别名（`@/*`）。
2) 公共层：`lib/api.ts`、`lib/utils.ts`、`types/*`、`stores/userStore`（登录、注销、鉴权状态）、`stores/productStore`（列表/详情缓存）。
3) UI 基座：Shadcn 组件生成、`MagmaButton`、`ParallaxCard`。
4) 路由 & 布局：AuthLayout、MainLayout、路由守卫。
5) 页面：
   - Auth：玻璃态表单，登录成功存 token 跳首页。
   - Home：Hero + 瀑布流列表，Hover 3D 倾斜，库存条、倒计时。
   - Product Detail：左右分栏，按钮状态（Pending/Active/Loading/Result），结果动画（confetti/shake），可选库存轮询，支持秒杀/下单入口。
   - Orders：列表（分页、状态筛选）、详情展示订单+支付状态。
6) 体验：引入 Motion-v + Lenis，全局 toast provider，焦点态/加载遮罩。

## 测试与验证
- 运行 `npm run dev` 验证路由与接口调用（需后端/Redis/Kafka 就绪）。
- `npm run build` 或 `vue-tsc --noEmit` 做类型/构建检查。
- 手工覆盖：登录失败提示、列表空态、秒杀失败提示、token 失效跳转。
