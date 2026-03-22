## 1. 规范与映射基线

- [x] 1.1 产出全仓 proto package / 目录 / go_package 到目标 `servora.*.v1` 的完整映射表
- [x] 1.2 确认业务 proto 是否只补 `servora.` 前缀还是同步收敛到 `servora.iam.*` 等领域层级
- [x] 1.3 补充仓库级 proto 命名规范文档，明确顶层命名空间、版本后缀、`service` 层与目录对齐规则

## 2. 迁移公共 proto

- [x] 2.1 迁移 `api/protos/audit/v1` 到 `api/protos/servora/audit/v1`，更新 `package`、imports、`go_package`（需 `make api`）
- [x] 2.2 迁移 `api/protos/mapper/v1`、`pagination/v1`、`conf/v1` 到 `api/protos/servora/**/v1`，更新 `package`、imports、`go_package`（需 `make api`）
- [x] 2.3 更新依赖公共 proto 的其他 `.proto` 文件 import 路径与 option 引用，确保 Buf lint 通过

## 3. 迁移业务与模板 proto

- [x] 3.1 迁移 IAM 相关 proto 目录与 package（`iam`、`authn`、`authz`、`user`、`application`），同步更新 `go_package` 与相互 import（需 `make api`）
- [x] 3.2 迁移 `sayhello` 与 `template` proto 目录与 package，保持示例与模板和正式规范一致（需 `make api`）
- [x] 3.3 更新 service proto 中对 audit annotation、shared proto 的引用到新的 `servora/*` 路径

## 4. 更新生成产物与代码引用

- [x] 4.1 运行代码生成，刷新 Go / TypeScript 生成产物并确认输出目录切换到 `api/gen/go/servora/**`（需 `make api`，如适用执行 `make api-ts`）
- [x] 4.2 批量更新 Go 代码中的 generated import 路径与 package alias，修复编译错误
- [x] 4.3 批量更新前端或工具链中引用生成 proto 路径的代码与配置，修复构建错误

## 5. 验证与收尾

- [x] 5.1 执行 `buf lint`、`make api`、`make build`、`make lint`，确认 package/目录/go_package 一致性生效
- [x] 5.2 如仓库当前启用 TS 生成校验，执行 `make api-ts` 与相关前端检查，确认路径迁移无回归
- [x] 5.3 更新受影响设计文档与 OpenSpec 引用中的 proto 路径，并整理迁移说明供后续实现使用
