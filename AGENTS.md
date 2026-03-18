# Repository Guidelines

## 核心上下文
- 后端：Go（Gin + GORM + Redis + Kafka + Viper），入口为 `cmd/api` 与 `cmd/worker`，业务代码位于 `internal/`。
- 前端：Vue 3 + TypeScript + Vite + Tailwind，入口为 `frontend/src/main.ts`，主要页面与状态在 `frontend/src/views`、`frontend/src/stores`。
- 文档：`README.md` 与 `docs/README.md` 是恢复上下文入口；`docs/*.md` 是事实来源，`AGENTS.md` 只保留仓库级约束。
- Agent 上下文：优先使用 `skills/` 下的 SneakerFlash skills 按任务加载文档，不要把任务细节长期堆在 `AGENTS.md`。

## 技能索引
- `sneakerflash-backend`：后端功能开发、接口联调、业务规则修改、Go 测试。
- `sneakerflash-frontend`：前端页面开发、交互调整、状态管理、前端测试。
- `sneakerflash-ops`：本地运行、依赖编排、压测、运行环境核查。
- `sneakerflash-troubleshooting`：故障定位、日志核查、标准化排障。
- `sneakerflash-doc-maintainer`：代码变更后的文档同步与入口维护。

## 通用不变式
- 配置统一通过 `SNEAKERFLASH_CONFIG` 指向本地 `config.<env>.local.yml`；优先复用 `make dev-init`、`make prod-init`。
- 不提交真实密钥、连接串、上传调试产物；`config.*.local.yml`、`.env*.local`、`uploads/` 保持忽略。
- 未经用户明确要求，不要执行 `git commit`、`git push`、创建/切换分支、重置历史。
- 新增或修改能力时保持 KISS、YAGNI、DRY；先复用现有实现，再做最小必要改动。

## 后端约束
- 严格保持 `handler -> service -> repository` 分层，仓储/服务继续使用 `WithContext`。
- 统一响应格式保持 `{ code, msg, data }`，错误码定义集中在 `internal/pkg/e`。
- 错误处理使用 `fmt.Errorf(... %w ...)` 与 `errors.Is`，不要新引入 `errors.Wrap`。
- 日志继续走 `internal/pkg/logger`；不要跨层直接访问底层资源。
- 涉及秒杀、订单、支付时，优先保护 Redis Lua 原子扣减、失败回滚、订单幂等、支付回调状态推进。
- `server.machineid` 是启动前置条件；涉及启动逻辑时不要破坏 Snowflake 初始化要求。

## 前端约束
- 保持 Editorial 视觉基调：浅底纸张层次、硬边细边框、克制动效；避免引入新的 UI 库破坏现有体系。
- 优先复用 `frontend/src/components/ui`、`frontend/src/components/motion`、`frontend/src/layout/MainLayout.vue`。
- 请求与鉴权逻辑继续经由 `frontend/src/lib/api.ts`；token 键保持 `access_token`、`refresh_token`。
- 新增状态管理优先遵循 `frontend/src/stores/userStore.ts`、`frontend/src/stores/productStore.ts` 的模式。
- 路由鉴权继续使用 `meta.requiresAuth` 与现有用户态守卫。

## 验证基线
- 后端改动至少运行相关 `go test`；涉及主链路、集成或跨模块行为时运行 `make test`、`make test-integration`。
- 后端代码检查使用根目录 `.golangci.yml`；后端改动完成后优先执行 `golangci-lint run ./...`。
- 前端改动至少保证 `pnpm build` 或 `make test-frontend` 相关检查通过；交互主链路改动时补跑 `make test-e2e`。
- 当前测试现状与覆盖范围以 `docs/testing.md` 为准，不要在 `AGENTS.md` 重复维护细表。

## 文档同步
- 接口或契约变化：更新 `docs/backend-api.md` 与 `docs/swagger.*`。
- 配置变化：更新 `docs/configuration.md`。
- 启动、编排、运维、压测、排障变化：更新 `docs/operations.md`、`docs/troubleshooting.md`、`docs/perf.md`。
- 前端结构或交互变化：更新 `docs/frontend-plan.md`。
- 测试基线变化：更新 `docs/testing.md`。
- 阶段完成度、维护规则、文档入口变化：更新 `docs/plan.md`、`docs/governance.md`、`docs/README.md`。
