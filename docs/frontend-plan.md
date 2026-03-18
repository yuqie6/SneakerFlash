# SneakerFlash 前端实现方案

## 当前结论
- 前端已经不是初始化阶段方案，而是**当前实现说明**。
- 现有页面覆盖：认证、首页、商品详情与发布、订单列表与详情、个人资料、VIP 中心、管理后台。
- 设计基调已经收敛为 Editorial 风格，后续迭代应以现有视觉体系和路由结构为准。

## 技术栈
- Vue 3 + TypeScript + Vite
- Vue Router + Pinia
- Tailwind CSS v3 + `class-variance-authority`
- Axios
- Motion-v + Lenis
- `vue-sonner` 用于提示反馈

## 目录结构
```text
src/
├─ assets/css/        # 全局样式、主题变量、toast 覆盖
├─ components/
│  ├─ ui/             # 通用基础组件
│  └─ motion/         # 动效组件
├─ layout/            # MainLayout
├─ lib/               # api.ts、admin.ts、utils.ts
├─ router/            # 路由与守卫
├─ stores/            # userStore、productStore
├─ types/             # 业务类型定义
└─ views/
   ├─ Admin/
   ├─ Auth/
   ├─ Home/
   ├─ Orders/
   ├─ Product/
   └─ User/
```

## 页面与路由
- `/`：商品首页
- `/login`、`/register`：认证页面
- `/product/:id`：商品详情与秒杀入口
- `/products/publish`：发布商品
- `/orders`、`/orders/:id`：订单列表与订单详情
- `/profile`：个人资料
- `/vip`：VIP 中心与优惠券入口
- `/admin` 及其子路由：统计、用户、订单、优惠券、商品、风控、审计日志

## 视觉与交互约束
- 主题色以 `#F9F8F6`、`#FFFFFF`、`#1C1C1C` 为主
- 标题字体使用 `Playfair Display`，正文使用 `Inter`
- 组件保持硬边、细边框、低装饰、克制动效
- 页面级风格以现有 Editorial 基调为准，不再沿用旧版“初始化阶段 UI 方案”

## 数据流与状态管理
### API 封装
- 统一通过 `frontend/src/lib/api.ts` 发起请求
- 默认 Base URL：`VITE_API_BASE_URL || /api/v1`
- 后端响应 `code != 200` 时统一抛错
- `401` 时自动用 `refresh_token` 刷新 access token，按后端 `{ code, msg, data }` 契约解包；失败后清理登录态并跳转登录页

### Token 存储
- `localStorage.access_token`
- `localStorage.refresh_token`

### Store 约定
- `userStore`：登录态、用户资料、管理员判定、资源级权限判定
- `productStore`：商品列表/详情缓存与刷新
- 新增 store 时保持同样模式，不直接在视图层散落请求与缓存逻辑

## 后端接口对接范围
- 认证：`/register`、`/login`、`/refresh`
- 用户：`/profile`、`/upload`
- 商品：`/products`、`/product/:id`、`/products/mine`
- 秒杀：`/seckill`
- 订单：`/orders`、`/orders/:id`、`/orders/poll/:order_num`、`/orders/:id/apply-coupon`
- 实时推送：`/stream/orders/:id`、`/stream/products/:id`
- 支付：`/payment/callback`
- VIP：`/vip/profile`、`/vip/purchase`
- 优惠券：`/coupons/mine`、`/coupons/purchase`
- 管理后台：`/admin/*`

## 已实现的关键体验
- 未登录访问受保护页面会跳转登录页
- 秒杀成功后支持轮询订单状态
- 订单详情支持支付态查看、优惠券应用，以及基于 SSE 的订单状态/库存同步；SSE 失败时回退到轮询
- 用户中心可查看 VIP 信息
- 管理后台已支持资源级权限裁剪；菜单和路由会按 `/profile.permissions` 收敛
- 管理后台已支持基础运营数据、风控名单维护与审计日志查询

## 测试与构建
- 开发：`pnpm dev`
- 构建：`pnpm build`
- 单测：`pnpm test:unit`
- E2E：`pnpm test:e2e`

## 后续前端工作建议
1. 若继续推进 P2，可补更多管理后台页面级组件测试，覆盖权限裁剪后的不同角色视图。
2. 若未来把 SSE 扩展到更多场景，仍建议保留现有轮询兜底，不把推送当成唯一事实源。
