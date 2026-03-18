# 运维与运行手册

## 目标
- 快速拉起本地依赖
- 提供单机生产部署方案
- 能排查常见故障
- 能执行压测和查看监控
- 能处理 Docker、Kafka、Redis、订单 pending 等常见异常

## 编排文件
| 文件 | 场景 | 说明 |
| --- | --- | --- |
| `docker-compose.dev.yaml` | 本地开发 / 联调 / 压测预演 | 配合 `.env.dev.local` 使用，暴露宿主机端口，包含 Kafka UI |
| `docker-compose.prod.yaml` | 单机生产 | 配合 `.env.prod.local` 使用，绑定 `127.0.0.1`，关闭 Kafka 自动建 topic |

> `docker-compose.prod.yaml` 是单机部署基线，不提供 Kafka 多 broker / 多 controller 高可用；正式生产建议迁移到托管 Kafka 或至少 3 broker 集群。

## 本地依赖编排
### 服务列表
| 服务 | 默认端口 | 用途 |
| --- | --- | --- |
| MySQL | `13306` | 订单、商品、支付、用户数据 |
| Redis | `16379` | 秒杀库存、限流、pending 状态 |
| Kafka | `19092` | 秒杀消息队列 |
| Kafka UI | `8080` | Kafka 可视化 |
| Prometheus | `9090` | 指标采集 |
| Grafana | `3000` | 指标面板 |

### 命令总览
```bash
make help
```

### 启动
```bash
make dev-init
make dev-up
```

- 首次启动后，建议先确认 `kafka-init` 已成功执行、`seckill_orders` 与 `seckill-order-dlq` 已创建，再启动 Worker。

应用侧建议统一使用同一份配置文件：
```bash
make dev-api
make dev-worker
```

- `.env.dev.local` 只给 `docker compose` 做变量替换
- `config.dev.local.yml` 才是 API / Worker 读取的应用配置
- API / Worker 未设置 `SNEAKERFLASH_CONFIG` 时不会启动

### 停止
```bash
make dev-down
```

## 单机生产基线
### 准备环境变量
```bash
make prod-init
```

- `.env.prod.local` 只给 `docker compose` 做变量替换
- `config.prod.local.yml` 才是 API / Worker 读取的应用配置
- API / Worker 未设置 `SNEAKERFLASH_CONFIG` 时不会启动

### 启动
```bash
make prod-init
make prod-up
```

- 若 Prometheus 需要从容器侧抓取宿主机 API 指标，请确保 API 监听地址对 `host.docker.internal` 可达，不要仅绑定 `127.0.0.1`。

应用侧建议显式绑定单机生产配置：
```bash
make prod-api
make prod-worker
```

### 停止
```bash
make prod-down
```

## 运维检查清单
### API 启动前
- MySQL 能连接
- Redis 能连接
- Kafka broker 可访问
- 当前环境对应的 `config.<env>.local.yml` 中 `server.machineid` 已配置

### Worker 启动前
- Kafka topic 可写
- `seckill_orders` 与 `seckill-order-dlq` 已创建
- Redis 可用
- DB 可写
- 若启用自动取消任务，确认 Worker 所在时区与业务期望一致；当前实现按应用本地时间扫描 15 分钟未支付订单

### 前端联调前
- API 服务可访问
- CORS 未被额外限制
- `VITE_API_BASE_URL` 配置正确

## 压测
### 标准入口
- `perf/k6-seckill.js`
- `perf/export_tokens.go`

### 推荐流程
1. 先导出 token
2. 创建压测商品
3. 低 RPS 试压
4. 逐步提升流量
5. 同时观察 Redis、Kafka、DB、API 指标

### 压测重点指标
- 秒杀成功率
- 业务失败率
- HTTP 错误率
- Kafka lag
- Redis latency
- DB 慢查询

## 监控建议
### Prometheus
- 采集 API `/metrics`
- 建议加入 Redis / MySQL / Kafka Exporter

### Grafana
- 面板建议至少覆盖：
  - API QPS / P95 / 错误率
  - Kafka lag / 消费速率
  - Redis 命中率 / 时延
  - MySQL QPS / 慢查询 / 连接数
  - 熔断状态切换次数 / 拒绝次数

## 日志
- 后端日志输出由 `internal/pkg/logger` 统一管理
- 日志文件默认位于 `log/`
- 推荐在生产环境按模块拆分 API / Worker 日志
- 后台关键运营动作会写入 `audit_logs` 表，可通过 `/admin/audit` 查询

## 优雅停机
- API 已接入 `SIGINT` / `SIGTERM` 信号处理：先停止接收新请求，再等待在途请求完成，最后依次关闭 Kafka、Redis、MySQL
- Worker 已接入同样的信号处理：先停止 cron，再退出 Kafka consumer，最后关闭 Kafka、Redis、MySQL
- 验证建议：
  - API：发送长请求后执行停止信号，确认请求能正常返回
  - Worker：消费中发送停止信号，确认不会丢失当前 batch

## 未支付订单自动取消
- Worker 默认每 30 秒扫描一次 15 分钟前创建且仍为 `unpaid` 的订单
- 自动取消会执行：
  - 订单状态推进到 `cancelled`
  - 支付单从 `pending` 推进到 `failed`
  - 释放已占用优惠券
  - 回补 MySQL / Redis 库存
  - 删除 Redis 中的重复下单标记
- 建议把 `cancelled` 订单占比、自动取消数量纳入日常观测

## SSE 实时推送
- 当前提供：
  - `/api/v1/stream/orders/:id?access_token=<token>`
  - `/api/v1/stream/products/:id?access_token=<token>`
- SSE 仅用于提升前端刷新体验，不承担事实源角色；异常时前端仍回退轮询
- 若接入反向代理，需关闭响应缓冲并允许长连接

## Docker 代理修复
### 现象
- `docker compose -f docker-compose.dev.yaml up -d` 拉镜像时报：
```text
proxyconnect tcp: dial tcp 127.0.0.1:10808: connect: connection refused
```

### 根因
- Docker daemon 仍从 systemd drop-in 读取旧代理：
  - `/etc/systemd/system/docker.service.d/http-proxy.conf`
  - `/etc/systemd/system/docker.service.d/https-proxy.conf`

### 修复命令
```bash
sudo mkdir -p "/etc/systemd/system/docker.service.d/backup-sneakerflash"
sudo cp "/etc/systemd/system/docker.service.d/http-proxy.conf" "/etc/systemd/system/docker.service.d/backup-sneakerflash/http-proxy.conf.bak"
sudo cp "/etc/systemd/system/docker.service.d/https-proxy.conf" "/etc/systemd/system/docker.service.d/backup-sneakerflash/https-proxy.conf.bak"
sudo rm -f "/etc/systemd/system/docker.service.d/http-proxy.conf" "/etc/systemd/system/docker.service.d/https-proxy.conf"
sudo systemctl daemon-reload
sudo systemctl restart docker
docker info | grep -i proxy
docker compose -f docker-compose.dev.yaml up -d
```

### 说明
- 若还有其他 systemd drop-in，请一并检查：
```bash
systemctl cat docker
```

## 消息重试与死信
### Outbox
- `internal/cron/outbox_cron.go` 定时扫描发送失败的消息
- Kafka 故障时，消息通过这个定时任务重试

### 死信
- 超过重试阈值的消息进入 DLQ
- 消费侧重试计数优先落在 Redis，key 形如 `kafka:consume:retry:<topic>:<partition>:<offset>`，用于跨 handler / rebalance 延续计数
- 正常 ack 或消息进入 DLQ 后，会删除对应重试计数 key，避免脏状态残留
- 建议对 DLQ 建立单独告警与处理手册

### 验证与回归
- 常规集成测试：`make test-integration`
- 定向验证 Kafka 死信链路：
```bash
GOCACHE="/tmp/go-build" go test -tags=integration ./internal/infra/kafka -run TestBatchConsumerHandler_ConsumeToDLQWithRealKafka -count=1
```
- 默认使用开发环境 Kafka / Redis：`127.0.0.1:19092`、`127.0.0.1:16379`
- 如需覆盖目标环境，可设置：`SNEAKERFLASH_KAFKA_IT_BROKERS`、`SNEAKERFLASH_REDIS_IT_ADDR`、`SNEAKERFLASH_REDIS_IT_PASSWORD`

### 排障要点
- 消费失败日志会带 `retry_key`、`retry_count`、`max_retries`、`dlq_topic`、`topic`、`partition`、`offset`
- 若一直重试但未进入 DLQ，先检查 Redis 中对应 `kafka:consume:retry:*` key 是否持续增长，再确认 `dlq_topic` 是否存在
- 若日志出现 `投递 DLQ 失败`，优先检查 Kafka topic 是否创建成功、worker 的 Kafka 连接是否正常
- 若需要确认 topic 是否存在，可执行：
```bash
docker compose -f "docker-compose.dev.yaml" exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
```
- 若需要确认 DLQ 是否已有消息，可执行：
```bash
docker compose -f "docker-compose.dev.yaml" exec -T kafka /opt/kafka/bin/kafka-get-offsets.sh --bootstrap-server localhost:9092 --topic seckill-order-dlq
```

### Kafka 基线
- 开发与生产基线都关闭自动建 topic，由 `kafka-init` 显式创建 `seckill_orders` / `seckill-order-dlq`
- 开发环境暴露 `19092` 给宿主机；单机生产基线默认暴露 `127.0.0.1:9092`
- 单机生产只有 1 个 broker，副本因子只能为 1，不是高可用方案
