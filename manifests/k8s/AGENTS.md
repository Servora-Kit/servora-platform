# AGENTS.md - manifests/k8s/

<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-03-15 | Updated: 2026-03-22 -->

## 目录定位

`manifests/k8s/` 保存基于 Kustomize 的 Kubernetes 清单。
父级 `manifests/AGENTS.md` 负责部署资产全局边界；本文件只补充 K8s 子树内的本地规则。

## 当前结构

```text
manifests/k8s/
├── base/         # namespace、rbac 与跨目录资源聚合入口
├── iam/          # IAM 服务部署与依赖资源
└── sayhello/     # SayHello 服务部署
```

## 本地约定

- `base/` 下维护集群级基础资源与聚合入口，重点查看 `kustomization.yaml` 中引用了哪些下游目录
- `iam/`、`sayhello/` 这类服务目录各自维护 `deployment.yaml`、`service.yaml`、`configmap.yaml` 等服务资源
- 若某个服务目录额外包含 `postgres.yaml`、`redis.yaml`、`init.sql` 等依赖清单，这些依赖被视为该服务部署的一部分，应与服务目录一起维护

## 修改 Kustomization 时必查

1. `resources:` 中引用的路径今天是否真实存在
2. `labels:`、`namespace:` 是否仍符合当前部署约定
3. 基础目录与服务目录职责是否混淆
4. 是否把已经迁移/废弃的旧服务路径继续保留在聚合入口里

## 维护提醒

- 当前 `base/kustomization.yaml` 若引用 `app/.../deployment/kubernetes` 这类路径，修改前先核对是否仍与仓库现状一致
- Kustomize 目录名以实际服务名为准，文档与清单都不要继续使用过时别名
- 若新增服务目录，优先复用现有 `iam/`、`sayhello/` 的清单组织方式，再决定是否需要更细的子级 AGENTS

## 禁止事项

- 不要把 Helm、Compose 或非 K8s 规则写进这个子级文件
- 不要在本文件重复父级已经说明的 `openfga/`、`scripts/` 规则
