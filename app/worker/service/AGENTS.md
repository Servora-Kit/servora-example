# Worker 示例服务

独立 gRPC 示例，受仓库生成流程管理；不装配具体 AuthN Adapter。

```bash
make gen
make wire
make build
make run
```

Proto 由根 `make gen` 生成到 `api/gen/go/`。Audit middleware 使用 `audit.WithRulesFuncs(workerpb.AuditRules)`；服务 Proto 声明 `servora.audit.v1.service_default` 后生成规则表。
