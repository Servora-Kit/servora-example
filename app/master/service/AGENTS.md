# AGENTS.md - app/master/service/

<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-03-15 | Updated: 2026-07-09 -->

## Purpose

独立示例服务，用于演示框架最小结构。含自有 `api/protos/`、`cmd/`、`internal/`，受根 `go.work` 管理。

## 常用命令

```bash
make gen
make build
make run
make wire
```

## For AI Agents

- 新增服务可参考本目录结构
- Proto 由根 `make gen` 统一生成到 `api/gen/go/`
- authn wiring 使用 `authn.Server + authn.Multi + authn.Named`，覆盖 jwt + apikey 双后端；不要添加 `jwt.Server()` wrapper。

### Audit 装配

audit middleware 使用 `audit.WithRulesFuncs(masterpb.AuditRules)` 传入生成的规则表，不使用 `WithSubjectFunc` / `WithAuthTypeFunc`（这两个 Option 已删除）：

```go
mw = append(mw, audit.Middleware(auditor,
    audit.WithRulesFuncs(masterpb.AuditRules),
))
```

authn 中间件配置了 `authn.WithAuditor(auditor)` 后会自动 emit 认证成功/失败事件，无需在 audit middleware 重复配置。

proto 文件需声明 audit 注解才会生成 `AuditRules()`：

```proto
import "servora/audit/v1/annotations.proto";

service MasterService {
  option (servora.audit.v1.service_default) = { mode: AUDIT_MODE_ENABLED };
  // ...
}
```
