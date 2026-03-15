# Servora 压测结果模板

## 1. 测试概况

- 日期：2026-03-11
- 执行人：horonlee
- 分支：example
- 提交：18716000328e49b3a70d36889d2537986ee248ed
- 环境：`compose`
- 服务版本：servora example branch (local compose stack)
- k6 版本：k6 v1.6.1 (go1.26.0, darwin/arm64)
- 压测机规格：Mac (Apple Silicon, local)

## 2. 环境信息

### 2.1 服务配置

- `BASE_URL`: http://127.0.0.1:8000
- HTTP 端口：8000
- gRPC 端口：8001 (unused in this test)
- 数据库：docker compose postgres (servora_db)
- Redis：docker compose redis (servora_redis)
- 下游服务：sayhello (compose)

### 2.2 观测入口

- Prometheus：http://localhost:9090
- Grafana：http://localhost:3001 (servora Overview dashboard)
- Jaeger：http://localhost:16686
- `/metrics`：http://127.0.0.1:8000/metrics

## 3. 测试参数

| 脚本 | 接口 | 模式 | 目标速率 | 持续时间 | 阈值 |
| --- | --- | --- | --- | --- | --- |
| `manifests/scripts/k6/baseline-test.js` | `POST /v1/test/test` | ramp | 200 RPS | 4m | p95<200ms, p99<500ms, fail<0.1% |
| `manifests/scripts/k6/hello-chain-test.js` | `POST /v1/test/Hello` | ramp | 120 RPS | 4m | p95<200ms, p99<500ms, fail<0.1% |
| `manifests/scripts/k6/auth-scenarios.js` | 鉴权场景 (login/read) | ramp | Login 30 RPS, Read 50 RPS | 4m | p95<200ms (login/read), p99<500ms, fail<0.1% |

## 4. 结果汇总

| 接口 | 环境 | 可持续 QPS | p50 | p95 | p99 | 错误率 | 主要瓶颈 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /v1/test/test` | compose | 200 RPS | 0.95ms | 2.42ms | 6.51ms | 0% | 无 |
| `POST /v1/test/Hello` | compose | 120 RPS | 3.96ms | 8.99ms | 28.05ms | 0% | 无 |
| `POST /v1/auth/login/email-password` | compose | ~30 RPS | 279ms | 333ms | 386ms | 0% | Login 阶段耗时超 p95 阈值，可能受密码校验/数据库访问影响 |
| `GET /v1/user/info` | compose | ~50 RPS | 1.0ms | 9.1ms | 17.2ms | 0% | 正常 |

## 5. 各场景详细记录

### 5.1 基线场景

- 脚本：`manifests/scripts/k6/baseline-test.js`
- 实际请求速率：200 RPS ramp stages
- 阈值是否通过：通过（p95 2.42ms）
- 关键异常：无
- 备注：VUs 1-50，CPU/内存稳定

### 5.2 跨服务链路场景

- 脚本：`manifests/scripts/k6/hello-chain-test.js`
- 实际请求速率：120 RPS ramp stages
- 阈值是否通过：通过（p95 8.99ms）
- 关键异常：无
- 备注：调用 sayhello 链路正常

### 5.3 鉴权场景

- 脚本：`manifests/scripts/k6/auth-scenarios.js`
- 实际请求速率：Login 30 RPS，Authenticated Read 50 RPS（ramp 4 stages）
- 阈值是否通过：登录 p95 超过 200ms（未通过），其余通过
- 关键异常：无错误请求，但登录耗时偏高
- 备注：LOGIN_EMAIL=admin@example.com, LOGIN_PASSWORD=114514，经 env 调用；setup/login 正常

## 6. 资源与依赖观察

### 6.1 应用侧

- CPU：峰值 ~5 cores，稳定在 <1 core（dashboard 截图）
- 内存：RSS ~45 MiB，Heap 10→15 MiB，平稳
- goroutine：33→35 左右，稳定
- GC pause：<0.001ms，稳定

### 6.2 PostgreSQL

- 连接数：未特别监控（compose 默认）
- 慢查询：未观测
- 锁等待：未观测

### 6.3 Redis

- ops/sec：未专门记录（compose redis）
- latency：未观测
- 内存：未观测

### 6.4 下游 sayhello

- 调用成功率：100%
- 耗时：p95 8.99ms（从 hello-chain 脚本推测）
- 异常：无

## 7. 结论

- 基线接口最大可持续 QPS：≥200 RPS（阈值通过）
- 跨服务接口最大可持续 QPS：≥120 RPS（阈值通过）
- 鉴权链路最大可持续 QPS：Login 30 RPS, Auth-Read 50 RPS（login p95 未达标）
- 主要系统瓶颈：登录接口耗时，p95≈333ms
- 下一步优化建议：
  1. 分析 `/v1/auth/login/email-password` 代码路径（密码哈希、数据库读写、token 签发）以降低延迟
  2. 结合 PostgreSQL/Redis 指标确认无资源争用；若需要更高并发可调优连接池
  3. 后续可加入 refresh 场景测试，并根据需求调整阈值
