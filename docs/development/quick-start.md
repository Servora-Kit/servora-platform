# 开发环境快速上手

## 前置要求

- Go 1.26+
- Docker & Docker Compose
- [Buf CLI](https://buf.build/docs/installation)（proto 编译）
- [fga CLI](https://openfga.dev/docs/getting-started/cli)（可选，用于模型验证和测试）

## 一键初始化

```bash
# 安装所有开发工具（protoc 插件 + CLI 工具 + svr + protoc-gen-servora-authz）
make init

# 启动基础设施（PostgreSQL + Redis + OpenFGA）
make compose.up

# 初始化 OpenFGA store 和授权模型
svr openfga init

# 统一代码生成（api + openapi + authz rules + wire + ent）
make gen

# 构建所有服务
make build
```

## 日常开发

### 启动开发环境

```bash
make compose.dev          # 启动基础设施 + 服务（Air 热重载）
```

### 修改 Proto 后

```bash
make gen                  # 统一生成（api + openapi + wire + ent）
```

`make gen` 内部按顺序执行：
1. `make api` → Go 代码 + AuthZ rules
2. `make openapi` → OpenAPI spec
3. `make wire` → 依赖注入
4. `make ent` → ORM schema

### 修改 Wire 依赖后

```bash
cd app/<service>/service && make wire
```

### 修改 OpenFGA 模型后

```bash
make openfga.model.validate    # 可选：语法检查
make openfga.model.test        # 可选：运行测试
make openfga.model.apply       # 上传新版本到 OpenFGA
# 然后重启服务
```

## 工具链总览

### `make init` 安装的工具

| 工具 | 来源 | 用途 |
|------|------|------|
| `protoc-gen-go` | google.golang.org/protobuf | proto → Go struct |
| `protoc-gen-go-grpc` | google.golang.org/grpc | proto → gRPC stub |
| `protoc-gen-go-http` | go-kratos | proto → Kratos HTTP handler |
| `protoc-gen-go-errors` | go-kratos | proto → Kratos error code |
| `protoc-gen-validate` | envoyproxy | proto → 验证规则 |
| `protoc-gen-openapi` | gnostic | proto → OpenAPI spec |
| `protoc-gen-typescript-http` | go-kratos | proto → TypeScript HTTP client |
| `protoc-gen-servora-authz` | **本仓库** `cmd/protoc-gen-servora-authz` | proto → AuthZ rules map |
| `buf` | bufbuild | Proto 编译管理 |
| `wire` | google | 编译期依赖注入 |
| `ent` | entgo.io | ORM schema 管理 |
| `svr` | **本仓库** `cmd/svr` | 统一开发 CLI |
| `kratos` | go-kratos | Kratos CLI |
| `gnostic` | google | OpenAPI 工具 |
| `golangci-lint` | golangci | Go lint 工具 |

### `svr` CLI 命令

```bash
svr new api <name> <server>      # 生成 proto 脚手架
svr gen gorm <service...>        # GORM GEN 代码生成
svr openfga init                 # 初始化 OpenFGA store + model
svr openfga model apply          # 更新 OpenFGA model 版本
```

## 环境变量

项目根目录的 `.env` 文件由 `svr openfga init` 自动维护：

| 变量 | 说明 | 来源 |
|------|------|------|
| `FGA_API_URL` | OpenFGA API 地址（通用） | `svr openfga init` 写入 |
| `FGA_STORE_ID` | OpenFGA store ID（通用） | `svr openfga init` 写入 |
| `FGA_MODEL_ID` | 当前 authorization model 版本 ID（通用） | `svr openfga init/model apply` 写入 |
| `IAM_FGA_*` | IAM 服务的 Kratos 前缀变量 | `svr openfga init --env-prefix IAM_` 写入 |

### Kratos 环境变量前缀机制

Kratos 配置加载器（`pkg/bootstrap/config/loader.go`）会根据服务名自动推导环境变量前缀：

```
serviceName = "iam.service" → prefix = "IAM_"
```

`bootstrap.yaml` 中的 `${FGA_STORE_ID}` 占位符需要名为 `IAM_FGA_STORE_ID` 的 OS 环境变量才能正确解析。因此 `.env` 中同时保留通用变量（供 `svr` CLI 使用）和带前缀的变量（供 Kratos 配置加载使用）。

`svr openfga init` 和 `model apply` 的 `--env-prefix` 参数会自动写入两套变量。`Makefile` 中 `OPENFGA_ENV_PREFIX` 默认为 `IAM_`。

## 清理

```bash
make compose.stop         # 停止容器（保留数据）
make compose.down         # 移除容器和网络（保留数据卷）
make compose.reset        # 彻底清理（含数据卷）
make clean                # 清理生成代码
```
