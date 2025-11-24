# 压力测试方案（k6）

本方案使用 k6 复现高并发秒杀链路，自动完成注册、登录、商品创建和高频 `/seckill` 调用，便于评估 Redis/Lua、Kafka 投递和 HTTP 层表现。

## 前置条件
- 本地依赖已按 `config.yml` 启动（MySQL/Redis/Kafka），建议关闭或放宽 `risk.enable` 相关限流，避免压测被中间件拦截。
- API 服务已启动：`go run ./cmd/api`；异步处理可同时启动 worker：`go run ./cmd/worker`。
- 已安装 k6：`https://k6.io/docs/get-started/installation/`。

## 运行示例
```bash
USER_COUNT=3000 USER_BATCH=500 SETUP_TIMEOUT=5m RATE=300 DURATION=60s \
k6 run perf/k6-seckill.js
```

可用的环境变量：
- `BASE_URL`：目标接口地址，默认 `http://localhost:8000/api/v1`。
- `RATE`：每秒请求数（constant-arrival-rate），默认 `200`。
- `DURATION`：压测时长，默认 `30s`。
- `PRE_ALLOCATED_VUS` / `MAX_VUS`：预分配和上限虚拟用户数。
- `USER_PREFIX` / `USER_COUNT` / `USER_PASSWORD`：压测账户前缀、数量与密码，默认 `perf_u`、`3000`、`PerfTest#123`；脚本会批量登录并仅对缺失用户注册。
- `USER_BATCH`：批量登录/注册并发度，默认 `200`，首次跑大用户量时可适当调高减少 `setup` 时间（也可通过 `SETUP_TIMEOUT=5m` 增大超时）。
- `PRODUCT_STOCK` / `PRODUCT_PRICE`：压测商品库存与单价。
- `START_DELAY_SEC`：商品开始时间距当前的秒数（需大于 0）。

## 指标关注点
- `seckill_success_rate`：业务成功率（code=200）。
- `seckill_business_fail_rate`：业务失败占比（如售罄、重复秒杀，单用户压测会接近 100%，多用户才有意义）。
- `http_error_rate` 与 `http_req_duration`：HTTP 层错误与延迟分布（默认阈值：错误率 <2%，p95 <800ms）。
- 结合 Redis/Kafka/DB 监控：确认库存扣减一致性、Kafka 投递失败率和消费堆积。

## 使用建议
- 先以低 `RATE` 探测，再逐级提升，结合 Redis/Kafka/DB 监控确认瓶颈位置。
- 如果需要只压 Redis/Lua 路径，可暂时停用 Kafka 或改写 `Send` 为空实现，以分离链路影响（改动需谨慎并限定在压测环境）。
