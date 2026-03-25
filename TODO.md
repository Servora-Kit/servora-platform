# TODO - servora-platform

## Grafana Dashboards

- [ ] 创建 Audit 服务专属 overview dashboard（参考 `servora-iam/manifests/grafana/dashboards/iam-overview.json`），包含：
  - 请求速率（HTTP/gRPC）
  - 错误率
  - Goroutine / 内存 / GC 指标
  - Kafka consumer lag
  - ClickHouse 写入速率
  - 审计事件处理延迟
- [ ] 调整通用 dashboard（`servora-traces`、`servora-logs`）的查询条件，限定为 platform 相关服务（`audit.service`）
