# AGENTS.md - web/ui/

<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-03-15 | Updated: 2026-03-22 -->

## 目录定位

`web/ui/` 是共享 UI 组件库，包名 `@servora/ui`。
当前以 React 组件、`styles.css` 与 `utils/*` 形式向各前端应用导出通用界面能力。

## 当前导出

- 组件：`Button`、`Card`、`Input`、`Label`、`Badge`、`Separator`
- 工具：`cn`
- 样式：`styles.css`

这些导出由 `src/index.ts` 和 `package.json` 的 `exports` 字段统一暴露。

## 修改约定

- 这里放 **可复用的基础 UI 组件与样式能力**，不要放服务专属页面组件
- 新增组件前先判断是否真的具备跨应用复用价值；如果只是 `web/iam/` 局部使用，优先留在应用内
- 修改导出面时，同步检查 `src/index.ts` 与 `package.json` 的 `exports` 是否一致
- 保持组件 API 稳定、通用，避免把具体业务语义写进 props 命名
- 组件风格当前围绕共享主题与 shadcn/ui 习惯组织，扩展时优先保持现有命名和文件粒度

## 目录边界

- `src/components/`：基础组件实现
- `src/utils/`：样式/类名等通用工具
- `src/styles.css`：共享样式入口
- 不要把页面容器、数据请求、路由、业务 hooks 放进 `web/ui/`

## 依赖与兼容性

- 这是 workspace 内共享包，消费方通过 `@servora/ui` 引用
- 当前 `peerDependencies` 包含 `react`、`tailwind-merge`、`clsx`
- 调整 peer 依赖或导出路径时，要考虑已有前端应用是否会被一并影响

## 验证建议

- 修改组件导出后，检查依赖它的前端应用导入路径是否仍然成立
- 如变更公共样式入口，确认不会破坏现有组件的视觉基线与类名组合方式
