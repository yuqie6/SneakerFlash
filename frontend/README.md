# SneakerFlash Frontend

## 技术栈
- `Vue 3`
- `TypeScript`
- `Vite`
- `Pinia`
- `Vue Router`
- `Tailwind CSS`

## 主要职责
- 商品大厅展示
- 商品详情与秒杀交互
- 订单详情与支付模拟
- 个人资料、VIP 与优惠券页面

## 启动
```bash
npm install
npm run dev
```

## 常用命令
```bash
npm run lint
npm run build
npm run preview
```

## 关键目录
```text
src/
├── views/          # 页面
├── stores/         # Pinia 状态
├── lib/api.ts      # 统一 API 封装
├── router/         # 路由
├── layout/         # 布局
└── components/     # UI 与动效组件
```

## 开发约定
- 统一通过 `src/lib/api.ts` 调后端
- 登录态与用户信息统一走 `userStore`
- 商品列表与详情缓存统一走 `productStore`
- 视觉风格保持暗夜黑金与玻璃态统一

