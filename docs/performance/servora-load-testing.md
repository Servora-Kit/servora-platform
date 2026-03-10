# Servora 压测指南（k6）

本文档给 `servora` 示例项目提供一套可直接落地的压测方案，目标是测清三类能力：

- HTTP 框架与中间件的基线吞吐
- 鉴权链路的吞吐与尾延迟
- 跨服务调用链路的真实可持续 QPS

## 1. 压测目标

本项目不适合直接问“理论上能跑多少 QPS”，因为最终上限受以下因素共同影响：

- `servora` 自身 HTTP/gRPC 处理能力
- JWT 鉴权与中间件开销
- PostgreSQL / Redis 依赖负载
- 下游 `sayhello` 服务的 gRPC 链路能力
- 本地与 Compose 环境的资源差异

因此建议用“最大可持续 QPS”作为统一口径：

- 错误率 `< 0.1%`
- `p95 < 200ms`
- `p99 < 500ms`
- 在目标速率下持续稳定 `3-5` 分钟

如果某个速率下 QPS 还能增长，但尾延迟或错误率已经明显恶化，则该速率不计入“可持续 QPS”。

## 2. 仓库中的关键事实

### 2.1 运行入口

- 服务入口：`app/servora/service/cmd/server/main.go`
- 本地运行：`cd app/servora/service && make run`
- 基础设施：仓库根目录 `make compose.up`
- 完整开发栈：仓库根目录 `make compose.dev.up`

### 2.2 端口与依赖

以 `app/servora/service/configs/local/bootstrap.yaml` 为准：

- HTTP: `0.0.0.0:8000`
- gRPC: `0.0.0.0:8001`
- PostgreSQL: `127.0.0.1:5432`
- Redis: `127.0.0.1:6379`
- OTel Collector: `localhost:4317`

注意：README 中提到的 gRPC `9000` 更像示例说明；本地压测请优先以 `app/servora/service/configs/local/bootstrap.yaml` 为准。

### 2.3 观测入口

HTTP server 会自动注册以下端点：

- `GET /metrics`
- `GET /healthz`
- `GET /readyz`

Prometheus 已在 `manifests/prometheus/prometheus.yml` 中配置抓取 `servora:8000/metrics`。

## 3. 推荐压测对象

### 3.1 基线接口

- 路径：`POST /v1/test/test`
- 请求体：`{}`
- 特点：公开接口、轻逻辑、无下游 IO
- 目标：测 HTTP 框架、公共中间件、序列化的基线开销

### 3.2 跨服务链路接口

- 路径：`POST /v1/test/Hello`
- 请求体：`{"req":"hello"}`
- 特点：会走到 `sayhello` 的 gRPC 调用链路
- 目标：测跨服务调用下的真实吞吐与尾延迟

### 3.3 鉴权场景

- 登录：`POST /v1/auth/login/email-password`
- 刷新 Token：`POST /v1/auth/refresh-token`
- 用户信息：`GET /v1/user/info`
- 私有测试接口：`POST /v1/test/private`

这组接口适合区分：

- 登录本身的吞吐上限
- 已鉴权接口的中间件开销
- Token 刷新路径的稳定性

对于刷新 Token 场景，建议不要让多个 VU 复用同一个 refresh token。更稳妥的做法有两种：

- 预先准备一组 refresh token，按 VU 分配
- 或者在每次刷新前先登录一次，获取新的 refresh token，再执行刷新

第一种更适合测 refresh 接口本身，第二种更适合做链路可用性验证。

### 3.4 写路径场景

- 注册：`POST /v1/auth/signup/using-email`
- 创建用户：`POST /v1/user/save`
- 更新用户：`POST /v1/user/update`

这类接口通常更容易受数据库写入、唯一约束、事务和数据污染影响，建议在完成只读链路压测后再测。

## 4. 建议的测试顺序

建议始终按照“从简单到复杂”的顺序推进：

1. `POST /v1/test/test`，先测基线
2. `POST /v1/test/Hello`，再测跨服务链路
3. 登录与已鉴权只读接口，测真实业务链路
4. 写接口，测数据库写放大和冲突成本

这样更容易定位瓶颈。如果一开始就混合所有接口，最终拿到的只是一个笼统的系统压力数字，很难知道上限被谁卡住了。

## 5. k6 场景模型

建议优先使用 arrival-rate 模型，而不是只用固定并发模型：

- `ramping-arrival-rate`：逐步升压，用于找性能拐点
- `constant-arrival-rate`：在某个速率下稳态运行，用于验证最大可持续 QPS

推荐流程：

1. 先预热 `30-60` 秒
2. 用 `ramping-arrival-rate` 找到延迟和错误率开始恶化的拐点
3. 选择拐点前一档速率，再用 `constant-arrival-rate` 稳态跑 `3-5` 分钟

对于鉴权脚本，建议把登录、已鉴权读取、刷新三个子场景拆开看，不要混用一个总阈值。否则你只能得到一个混合延迟分位，无法判断到底是哪条子链路先成为瓶颈。

## 6. 结果判定方法

每个接口都建议输出一行独立结论，例如：

| 接口 | 环境 | 可持续 QPS | p95 | p99 | 错误率 | 主要瓶颈 |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /v1/test/test` | local | 420 | 42ms | 110ms | 0.00% | CPU |
| `POST /v1/test/Hello` | compose | 180 | 96ms | 240ms | 0.02% | 下游 sayhello |

不要只记录“峰值 QPS”。真正有意义的是：

- 该 QPS 是否连续稳定
- 对应尾延迟是否仍满足目标
- 是否出现依赖侧抖动或排队

## 7. 配套观测建议

压测期间建议同时观察：

- k6 侧：`http_reqs`、`http_req_duration`、`http_req_failed`
- 服务侧：`/metrics` 中的请求量和耗时直方图
- 资源侧：CPU、内存、goroutine、GC pause
- 依赖侧：PostgreSQL 连接数、慢查询、Redis ops/sec

可直接使用本仓库已有组件：

- Grafana: `http://localhost:3001`
- Prometheus: `http://localhost:9090`
- Jaeger: `http://localhost:16686`

## 8. 建议的落地步骤

### 8.1 本地模式

1. 启动基础设施：`make compose.up`
2. 启动服务：`cd app/servora/service && make run`
3. 确认 `http://127.0.0.1:8000/healthz` 与 `/metrics` 正常
4. 运行 `scripts/k6/` 下的 k6 脚本

### 8.2 Compose 全链路模式

1. 在仓库根目录执行 `make compose.dev.up`
2. 确认 `servora`、`sayhello`、PostgreSQL、Redis、Prometheus、Grafana 都已启动
3. 用相同脚本重复压测
4. 对比本地模式与全链路模式的差异

## 9. 结果解读经验

### 9.1 基线 QPS 高、链路 QPS 低

这通常说明框架开销不是问题，瓶颈更可能在：

- 下游 `sayhello`
- 数据库
- Redis
- 鉴权逻辑

### 9.2 QPS 还能涨，但 p99 爆掉

这不应算进“最大可持续 QPS”。系统虽然还在吞请求，但用户体感和系统稳定性已经恶化。

### 9.3 写接口数据越来越乱

写路径压测容易受到数据污染影响。建议：

- 使用独立压测数据库
- 用时间戳或随机后缀避免唯一键冲突
- 每轮压测前后清理测试数据

## 10. 对当前仓库的推荐结论格式

对 `servora`，建议最终至少给出以下三条结论：

- `POST /v1/test/test` 的最大可持续 QPS
- `POST /v1/test/Hello` 的最大可持续 QPS
- 一个鉴权接口（推荐 `GET /v1/user/info` 或 `POST /v1/auth/login/email-password`）的最大可持续 QPS

这样你最终得到的不是一个模糊的“服务上限”，而是三种不同链路模型下的清晰上限。
