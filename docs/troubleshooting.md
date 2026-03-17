# 故障排查手册

## 使用方式
- 先看“现象”
- 再看“根因判断”
- 最后执行“处理动作”

## Docker 拉镜像失败
### 现象
```text
proxyconnect tcp: dial tcp 127.0.0.1:10808: connect: connection refused
```

### 根因判断
- Docker daemon 仍使用旧代理
- 常见位置：
  - `/etc/systemd/system/docker.service.d/http-proxy.conf`
  - `/etc/systemd/system/docker.service.d/https-proxy.conf`

### 处理动作
- 按 `operations.md` 中“Docker 代理修复”章节执行

## API 启动失败：Snowflake 初始化失败
### 现象
- 启动日志提示初始化雪花算法失败

### 根因判断
- `server.machineid` 缺失或非法

### 处理动作
- 在 `config.yml` 中补充：
```yaml
server:
  machineid: 1
```

## 秒杀接口直接报错或高频失败
### 可能原因
- Redis 未连通
- `risk.enable` 打开且阈值过低
- 商品未开始或已结束

### 排查顺序
1. 检查 Redis 连接
2. 检查商品 `start_time` / `end_time`
3. 检查限流配置
4. 检查 API 日志与 `/metrics`

## 订单长期 pending
### 可能原因
- Worker 未启动
- Kafka broker 不可用
- Outbox 消息未成功投递
- Worker 消费失败并未完成补偿

### 排查动作
1. 确认 `go run ./cmd/worker` 正常运行
2. 打开 Kafka UI 检查 topic 是否有堆积
3. 查看 Worker 日志是否存在事务失败
4. 查看 `outbox_messages` 状态与重试次数

## 支付后订单状态未变化
### 可能原因
- `POST /payment/callback` 未成功调用
- `payment_id` 不存在
- 支付状态不符合状态机推进条件

### 排查动作
1. 检查支付回调请求与响应
2. 查询 `payments` 表状态
3. 查询 `orders` 表状态
4. 检查订单是否已被重复回调处理

## 优惠券无法使用
### 可能原因
- 用户券状态不是 `available`
- 已过期
- 未达到使用门槛
- 订单状态不是待支付

### 排查动作
1. 查用户券状态
2. 查券模板门槛
3. 查订单状态与支付状态

## Kafka 发送或消费异常
### 排查动作
1. 确认 broker 地址与 `advertised.listeners`
2. 检查容器内外访问地址是否混用
3. 确认 topic 存在
4. 检查 producer / consumer 日志

## Redis 库存与 DB 库存不一致
### 根因判断
- 入口抢占成功，但 Worker 事务失败后未及时恢复
- 缓存刷新任务丢失或延迟

### 排查动作
1. 查看 Worker 事务失败日志
2. 确认回滚逻辑是否执行
3. 手动对账 Redis 与 DB
4. 必要时人工重建指定商品库存缓存

## 文档不同步
### 现象
- Swagger、接口文档、代码行为不一致

### 处理动作
1. 以代码为准核对 `handler` 与 `service`
2. 更新 `docs/backend-api.md`
3. 重新生成 Swagger
4. 在 `governance.md` 记录本次同步

