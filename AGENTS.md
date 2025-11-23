# Repository Guidelines

## 项目结构与模块
- 后端 Go 代码集中于 `cmd/`（`cmd/api` HTTP 服务，`cmd/worker` Kafka 消费者）与 `internal/`（配置、数据库、仓储、服务、路由等），公共配置示例位于根目录 `config.yml`。
- 前端位于 `frontend/`，使用 Vue 3 + TypeScript + Vite + Tailwind。静态资源在 `frontend/public/`，入口 `frontend/src/main.ts`，页面与组件在 `frontend/src`。
- 设计文档与接口说明在 `docs/`，可先读 `docs/backend-api.md`、`docs/frontend-plan.md`。

## 构建、运行与开发命令
- 后端：`go mod tidy` 拉取依赖；`go run ./cmd/api` 启动 HTTP 服务（需本地 MySQL/Redis/Kafka 按 `config.yml` 配置）；`go run ./cmd/worker` 启动异步订单消费者；`go build ./cmd/api` 可生成可执行文件。
- 前端：在 `frontend/` 下 `npm install`，`npm run dev` 本地开发，`npm run build` 产出静态包到 `frontend/dist`，`npm run preview` 本地预览构建产物。
- Docker：根目录有 `docker-compose.yaml.example` 作为本地依赖参考，可复制为 `docker-compose.yaml` 后按需调整。

## 编码风格与命名约定
- Go：使用 `gofmt`/`goimports` 保持风格；包名小写、无下划线；错误返回使用 `errors.Wrap` 风格可读信息；处理请求参数请遵循 handler/ service/repository 分层。
- TypeScript/Vue：组件/文件用 PascalCase（例如 `ProductCard.vue`），组合式 API `<script setup>`；样式用 Tailwind 原子类，复用样式可用 `class-variance-authority`；避免在组件内直接写业务请求，使用独立的 API 封装。
- 命名：HTTP 路由和前端路径统一小写短横线；Kafka/Redis 主题与键名保持配置中心化，勿硬编码。

## 测试指引
- 当前未提供现成测试文件；新增或修改核心逻辑时请补充 Go `testing` 表驱动用例（按包拆分），确保 `go test ./...` 可通过。
- 前端如引入交互逻辑，建议添加最小化单元/组件测试（可接入 Vitest），并保持构建通过。
- 数据相关功能请提供最小化种子数据或假设说明，避免测试依赖生产配置。

## 提交与 Pull Request
- Commit 信息沿用仓库历史的简洁中文动词短语（示例：“完善订单模型”、“前端跟进”），一次提交聚焦单一主题。
- PR 描述应包含：变更目的、主要改动点、影响范围（后端接口/前端页面/配置）、本地验证方式（命令输出或截图），如关联 Issue 请显式引用。
- 确认已运行相关构建/测试命令并同步结果；涉及 API 变更请更新 `docs/backend-api.md` 或在 PR 中附差异说明。

## 配置与安全
- 勿提交真实密钥或生产连接串；`config.yml` 仅为本地示例，可通过环境变量覆盖（读取逻辑位于 `internal/config`）。
- 变更 Kafka/Redis/MySQL 连接信息前请同步 worker 与 API 入口；确保雪花算法 `machineid` 在多实例场景唯一。
