# SneakerFlash

球鞋秒杀电商系统。Go 后端处理秒杀、订单、支付；Vue 3 前端提供购买界面。

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.25, Gin, GORM, Redis, Kafka, Viper |
| 前端 | Vue 3, TypeScript, Vite, Pinia, Tailwind CSS |
| 基础设施 | MySQL 8.4, Redis 8, Kafka 4.x (KRaft), Prometheus, Grafana |
| 工具链 | golangci-lint, ESLint, Prettier, k6 |

## 快速启动

前置条件：Go >= 1.25、Node.js >= 18、Docker

```bash
# 1. 生成本地配置文件，启动 MySQL/Redis/Kafka 等依赖
make dev-up

# 2. 启动后端 API（端口 8000）
make dev-api

# 3. 启动 Kafka 消费者（处理订单写入）
make dev-worker

# 4. 启动前端（端口 5173）
make dev-frontend
```

首次启动前编辑 `config.dev.local.yml`，填入 `server.machineid`（Snowflake ID 机器号），否则启动会报错。配置详情见 [docs/configuration.md](docs/configuration.md)。

## 仓库结构

```
cmd/                     API 和 Worker 启动入口
internal/                后端核心代码（handler/service/repository/middleware）
frontend/                Vue 3 前端
docs/                    项目文档
perf/                    k6 压测脚本
docker/                  Prometheus、Grafana 配置
docker-compose.dev.yaml  开发环境容器编排
docker-compose.prod.yaml 单机生产容器编排
Makefile                 所有常用命令的入口
```

## 秒杀流程

1. 用户发起 `POST /api/v1/seckill` 请求
2. Redis 通过 Lua 脚本一次性完成：查库存 → 扣库存 → 标记用户已购买
3. 扣减成功后，消息发到 Kafka
4. Worker 从 Kafka 取出消息，在数据库中创建订单和支付单
5. 前端轮询 `GET /api/v1/orders/poll/:order_num` 等待结果
6. 支付回调到达后，更新订单状态、用户等级、优惠券

## 常用命令

| 命令 | 作用 |
|---|---|
| `make dev-up` | 启动开发环境依赖 |
| `make dev-down` | 停止开发环境依赖 |
| `make dev-api` | 启动 API 服务 |
| `make dev-worker` | 启动 Kafka Worker |
| `make dev-admin USERNAME=alice` | 将开发环境中的指定用户提权为管理员 |
| `make dev-frontend` | 启动前端开发服务器 |
| `make lint` | Go + 前端代码检查 |
| `make test` | Go 单元测试 |
| `make test-integration` | Go 集成测试 |
| `make test-all` | 全部测试（单元 + 集成 + 前端） |
| `make build-api` | 编译 API |
| `make build-worker` | 编译 Worker |
| `make frontend-build` | 构建前端生产包 |
| `make help` | 查看所有命令 |

生产环境对应 `make prod-*` 系列命令，用法相同。

如果需要启用管理后台，不再需要手写 SQL；先注册普通用户，再执行：

```bash
make dev-admin USERNAME=alice
```

随后重新登录该账号即可获得管理员权限。若要使用其他配置文件，也可以执行：

```bash
make admin CONFIG=./config.dev.local.yml USERNAME=alice
```

## 文档

| 文档 | 内容 |
|---|---|
| [docs/development.md](docs/development.md) | 开发指南 |
| [docs/configuration.md](docs/configuration.md) | 配置说明 |
| [docs/architecture.md](docs/architecture.md) | 架构设计 |
| [docs/backend-api.md](docs/backend-api.md) | 接口文档 |
| [docs/operations.md](docs/operations.md) | 运维手册 |
| [docs/troubleshooting.md](docs/troubleshooting.md) | 故障排查 |
| [docs/testing.md](docs/testing.md) | 测试方案 |
| [docs/perf.md](docs/perf.md) | 压测说明 |

## 已知问题

- Docker 绑了旧代理时拉镜像会失败（`proxyconnect tcp: dial tcp 127.0.0.1:10808`），解决方法见 [docs/troubleshooting.md](docs/troubleshooting.md)
- `docker-compose.prod.yaml` 是单机配置，不是高可用方案
