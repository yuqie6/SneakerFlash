# 运维与运行手册

## 目标
- 快速拉起本地依赖
- 能定位常见链路问题
- 能执行压测和观测
- 能处理 Docker、Kafka、Redis、订单 pending 等常见异常

## 本地依赖编排
### 服务列表
| 服务 | 默认端口 | 用途 |
| --- | --- | --- |
| MySQL | `13306` | 订单、商品、支付、用户数据 |
| Redis | `16379` | 秒杀库存、限流、pending 状态 |
| Kafka | `9092` / `19092` | 秒杀消息与异步建单 |
| Kafka UI | `8080` | Kafka 可视化 |
| Prometheus | `9090` | 指标采集 |
| Grafana | `3000` | 指标面板 |

### 启动
```bash
docker compose up -d
```

### 停止
```bash
docker compose down
```

## 运维检查清单
### API 启动前
- MySQL 能连接
- Redis 能连接
- Kafka broker 可访问
- `config.yml` 中 `server.machineid` 已配置

### Worker 启动前
- Kafka topic 可写
- Redis 可用
- DB 可写

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

## 日志
- 后端日志输出由 `internal/pkg/logger` 统一管理
- 日志文件默认位于 `log/`
- 推荐在生产环境按模块拆分 API / Worker 日志

## Docker 代理修复
### 现象
- `docker compose up -d` 拉镜像时报：
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
docker compose up -d
```

### 说明
- 若还有其他 systemd drop-in，请一并检查：
```bash
systemctl cat docker
```

## 补偿与治理
### Outbox
- `internal/cron/outbox_cron.go` 定时扫描未成功发送的消息
- 若 Kafka 故障，消息最终通过补偿任务重试

### 死信
- 超过重试阈值的消息进入 DLQ
- 建议对 DLQ 建立单独告警与处理手册

## 推荐后续补强
- 增加生产环境部署文档
- 增加告警阈值基线
- 增加常见运维 SOP 与回滚手册

