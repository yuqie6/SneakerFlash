# SneakerFlash 测试现状与策略

## 当前现状（2026-03-18）
### 后端已落地
- 单元测试已覆盖：JWT、密码工具、限流器、用户服务、订单服务、秒杀服务、优惠券服务、健康检查、Admin Handler。
- 集成测试已覆盖：认证流程、支付流程、Worker 流程、管理后台流程。
- 测试目录已存在于 `internal/service`、`internal/middlerware`、`internal/pkg`、`internal/handler`、`internal/integration`。

### 前端已落地
- 已接入 `Vitest`、`@vue/test-utils`、`Playwright`。
- 当前已有单测：`frontend/src/lib/api.test.ts`、`frontend/src/router/index.test.ts`、`frontend/src/stores/userStore.test.ts`、`frontend/src/stores/productStore.test.ts`。
- 当前已有 E2E：`frontend/tests/e2e/auth.spec.ts`、`frontend/tests/e2e/seckill.spec.ts`、`frontend/tests/e2e/order-payment.spec.ts`。

## 运行方式
- 后端单元测试：`make test` 或 `go test ./...`
- 后端集成测试：`make test-integration`
- 前端单测：`make test-frontend` 或 `cd frontend && pnpm test:unit`
- 前端 E2E：`make test-e2e` 或 `cd frontend && pnpm test:e2e`
- 全量：`make test-all`

## 当前覆盖重点
### 后端
- 用户注册、登录、刷新 token、管理员提权
- JWT 鉴权与管理员权限校验
- 秒杀主链路的业务分支
- 订单查询、轮询、支付回调、优惠券应用
- 优惠券模板与用户券核心逻辑
- 健康检查与管理后台关键接口

### 前端
- API 封装的 token 刷新与异常处理
- Router 登录/管理员守卫
- 用户与商品 Store 的状态流转
- 认证、秒杀、订单支付主链路 E2E

## 仍然不足的部分
- 缺少针对风控黑灰名单、热点参数限流的集成测试
- 缺少 Outbox 定时补偿、DLQ 回放、Kafka 异常注入测试
- 缺少未支付自动取消、实时推送、优雅停机等 P2 能力测试（对应能力本身也未完成）
- 前端复杂视图组件覆盖仍偏少，管理后台页面尚未形成系统化组件测试

## 建议的补强顺序
1. 补风控限流与 Outbox 补偿的集成测试，先覆盖当前最容易回归的后端链路。
2. 为管理后台关键页面补充前端组件测试，减少仅靠 E2E 兜底。
3. 等 P2 能力落地后，再补未支付自动取消、实时推送、优雅停机的专项测试。
