# k6 压测脚本说明

本目录提供面向 `servora` 的最小可用 k6 模板，覆盖三类常见场景：

- 基线场景：`baseline-test.js`
- 跨服务链路场景：`hello-chain-test.js`
- 鉴权与已登录接口场景：`auth-scenarios.js`

## 1. 前置准备

### 本地模式

```bash
make compose.up
cd app/servora/service
make run
```

确认以下地址正常：

- `http://127.0.0.1:8000/healthz`
- `http://127.0.0.1:8000/metrics`

### Compose 全链路模式

```bash
make compose.dev.up
```

## 2. 安装 k6

macOS 可直接使用：

```bash
brew install k6
```

## 3. 通用环境变量

- `BASE_URL`：默认 `http://127.0.0.1:8000`
- `MODE`：`ramp` 或 `steady`，默认 `ramp`
- `RATE`：目标请求速率，默认随脚本而定
- `DURATION`：稳态持续时间，默认 `3m`
- `PRE_ALLOCATED_VUS`：默认 `50`
- `MAX_VUS`：默认 `500`
- `P95_MS`：默认 `200`
- `P99_MS`：默认 `500`
- `FAIL_RATE`：默认 `0.001`
- `SCENARIOS`：仅对 `auth-scenarios.js` 生效，可选 `login,read,refresh`

## 4. 运行示例

### 4.1 基线接口

```bash
k6 run scripts/k6/baseline-test.js
```

稳态验证 300 QPS：

```bash
MODE=steady RATE=300 DURATION=5m k6 run scripts/k6/baseline-test.js
```

### 4.2 跨服务链路接口

```bash
k6 run scripts/k6/hello-chain-test.js
```

### 4.3 鉴权场景

登录压力 + 已登录接口验证依赖以下变量：

- `LOGIN_EMAIL`
- `LOGIN_PASSWORD`

示例：

```bash
LOGIN_EMAIL=admin@example.com \
LOGIN_PASSWORD=123456 \
k6 run scripts/k6/auth-scenarios.js
```

如果你已经手动拿到了 Token，也可以直接传：

- `ACCESS_TOKEN`
- `REFRESH_TOKEN`

只压已登录读取接口时，推荐显式指定：

```bash
SCENARIOS=read \
ACCESS_TOKEN=your-token \
k6 run scripts/k6/auth-scenarios.js
```

如果同时具备 access token 和 refresh token，可指定：

```bash
SCENARIOS=read,refresh \
ACCESS_TOKEN=your-token \
REFRESH_TOKEN=your-refresh-token \
k6 run scripts/k6/auth-scenarios.js
```

## 5. 推荐输出与记录方式

建议每次压测都加一个 summary 导出：

```bash
k6 run --summary-export tmp/k6-summary.json scripts/k6/baseline-test.js
```

再把结果填入：

- `docs/performance/load-test-results-template.md`

## 6. 建议流程

1. 先跑 `baseline-test.js`，拿基线 QPS
2. 再跑 `hello-chain-test.js`，确认跨服务链路上限
3. 最后跑 `auth-scenarios.js`，确认登录和已鉴权接口的成本
4. 将本地模式与 Compose 模式分别记录并对比
