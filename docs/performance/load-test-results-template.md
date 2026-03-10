# Servora 压测结果模板

## 1. 测试概况

- 日期：
- 执行人：
- 分支：
- 提交：
- 环境：`local` / `compose` / `k8s`
- 服务版本：
- k6 版本：
- 压测机规格：

## 2. 环境信息

### 2.1 服务配置

- `BASE_URL`:
- HTTP 端口：
- gRPC 端口：
- 数据库：
- Redis：
- 下游服务：

### 2.2 观测入口

- Prometheus：
- Grafana：
- Jaeger：
- `/metrics`：

## 3. 测试参数

| 脚本 | 接口 | 模式 | 目标速率 | 持续时间 | 阈值 |
| --- | --- | --- | --- | --- | --- |
| `scripts/k6/baseline-test.js` | `POST /v1/test/test` | ramp / steady |  |  |  |
| `scripts/k6/hello-chain-test.js` | `POST /v1/test/Hello` | ramp / steady |  |  |  |
| `scripts/k6/auth-scenarios.js` | 鉴权场景 | ramp / steady |  |  |  |

## 4. 结果汇总

| 接口 | 环境 | 可持续 QPS | p50 | p95 | p99 | 错误率 | 主要瓶颈 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /v1/test/test` |  |  |  |  |  |  |  |
| `POST /v1/test/Hello` |  |  |  |  |  |  |  |
| `POST /v1/auth/login/email-password` |  |  |  |  |  |  |  |
| `GET /v1/user/info` |  |  |  |  |  |  |  |

## 5. 各场景详细记录

### 5.1 基线场景

- 脚本：`scripts/k6/baseline-test.js`
- 实际请求速率：
- 阈值是否通过：
- 关键异常：
- 备注：

### 5.2 跨服务链路场景

- 脚本：`scripts/k6/hello-chain-test.js`
- 实际请求速率：
- 阈值是否通过：
- 关键异常：
- 备注：

### 5.3 鉴权场景

- 脚本：`scripts/k6/auth-scenarios.js`
- 实际请求速率：
- 阈值是否通过：
- 关键异常：
- 备注：

## 6. 资源与依赖观察

### 6.1 应用侧

- CPU：
- 内存：
- goroutine：
- GC pause：

### 6.2 PostgreSQL

- 连接数：
- 慢查询：
- 锁等待：

### 6.3 Redis

- ops/sec：
- latency：
- 内存：

### 6.4 下游 sayhello

- 调用成功率：
- 耗时：
- 异常：

## 7. 结论

- 基线接口最大可持续 QPS：
- 跨服务接口最大可持续 QPS：
- 鉴权链路最大可持续 QPS：
- 主要系统瓶颈：
- 下一步优化建议：
