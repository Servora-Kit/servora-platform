# Servora Platform

简体中文

> 本项目是 [Servora](https://github.com/Servora-Kit/servora) 框架的**示例项目**，提供平台级基础微服务实现。

`servora-platform` 当前包含审计（Audit）微服务，后续将持续扩展更多平台级基础服务。

## 包含内容

### 微服务

- **Audit 服务**（`app/audit/service/`）：全链路审计日志服务
  - 基于 Kafka 消费审计事件
  - ClickHouse 持久化存储
  - 审计日志查询 API

### 部署

- OpenFGA model：`manifests/openfga/`

## 技术栈

- 框架：[servora](https://github.com/Servora-Kit/servora)（Kratos v2）
- API：Protobuf + Buf v2（业务 proto 依赖 [buf.build/servora/servora](https://buf.build/servora/servora)）
- DI：Google Wire
- 消息：Kafka（franz-go）
- 存储：ClickHouse（审计日志）
- 授权：OpenFGA

## 项目结构

```text
.
├── api/
│   └── gen/go/                      # Go 生成代码（业务 proto，勿手改）
├── app/
│   └── audit/service/               # Audit 微服务
│       ├── api/protos/              # 审计业务 proto
│       ├── cmd/                     # 服务入口
│       ├── configs/
│       │   ├── local/               # 本地开发配置（air 热重载读取）
│       │   └── docker/              # 容器部署配置
│       └── internal/                # 业务实现（service/biz/data/server）
├── manifests/
│   └── openfga/                     # OpenFGA model 与测试
├── buf.yaml                         # Buf v2 workspace（依赖 buf.build/servora/servora）
├── buf.go.gen.yaml                  # Go 代码生成模板
├── docker-compose.yaml              # 基础设施（Kafka、ClickHouse、Consul 等）
├── docker-compose.apps.yaml         # 应用容器（audit 生产镜像）
├── docker-compose.dev.yaml          # 开发环境（audit 容器化开发）
└── Makefile                         # 构建入口
```

## 快速开始

### 前置要求

- Go 1.26+
- Make
- Docker / Docker Compose

### 安装工具

```bash
make init    # 安装 protoc 插件与 CLI 工具
```

### 生成代码

```bash
make gen     # 统一生成（api + wire）
```

### 启动开发环境

两种工作流，按需选择：

**方式一：本地热重载（推荐日常开发）**

```bash
# 启动基础设施
make compose.up

# 在服务目录用 air 热重载启动
cd app/audit/service && make dev
```

**方式二：全容器化**

```bash
# 构建应用镜像
make compose.build

# 启动基础设施 + 应用容器
make compose.up.all

# 或仅启动开发环境（带容器内服务）
make compose.dev
```

### 常用命令

```bash
# 代码生成
make gen                    # 统一生成
make api                    # 仅生成 proto Go 代码
make wire                   # 仅生成 Wire

# 质量检查
make test                   # 运行测试
make lint                   # Go lint
make lint.proto             # Proto lint

# 服务目录（app/audit/service/）
make dev                    # air 热重载启动（读 configs/local/）
make run                    # 直接运行（读 configs/local/）
make build                  # 编译二进制

# Compose - 基础设施
make compose.up             # 启动基础设施
make compose.stop           # 停止基础设施
make compose.down           # 移除容器/网络（保留数据卷）
make compose.reset          # 移除容器/网络/数据卷

# Compose - 应用镜像
make compose.build          # 构建应用镜像（同时打 :latest tag）
make compose.up.all         # 启动基础设施 + 应用容器

# Compose - 开发环境（容器化）
make compose.dev            # 启动开发环境并 tail 日志
make compose.dev.up         # 启动开发环境（后台）
make compose.dev.restart    # 重启微服务容器
make compose.dev.stop       # 停止微服务容器
make compose.dev.down       # 移除开发环境容器/网络
make compose.dev.reset      # 移除开发环境容器/网络/数据卷

# OpenFGA
make openfga.init           # 初始化 store
make openfga.model.validate # 验证 model
make openfga.model.test     # 测试 model
make openfga.model.apply    # 应用 model 更新
```

## 依赖关系

本项目依赖 servora 核心框架：

- **Go 依赖**：`github.com/Servora-Kit/servora`（基础库）、`github.com/Servora-Kit/servora/api/gen`（框架 proto 生成代码）
- **Proto 依赖**：`buf.build/servora/servora`（框架公共 proto）
- **CLI 工具**：`svr`、`protoc-gen-servora-authz`、`protoc-gen-servora-audit`、`protoc-gen-servora-mapper`

本地联合开发时通过顶层 `go.work` 实现跨仓库引用。

## 质量约束

- 不要手动编辑生成代码：`api/gen/go/`、`wire_gen.go`
- 修改 proto 后执行 `make gen`
- 修改 Wire 依赖图后执行 `make wire`
- 修改 OpenFGA model 后执行 `make openfga.model.apply`
- 提交前通过 `make lint` 与 `make test`

## License

MIT，详见 `LICENSE`。
