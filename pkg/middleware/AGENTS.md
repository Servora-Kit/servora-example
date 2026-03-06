<!-- Parent: ../AGENTS.md -->
# 中间件 (pkg/middleware)

**最后更新时间**: 2026-03-06

## 模块目的

提供可复用中间件工具。当前重点是：
- `cors/`：HTTP CORS 中间件
- `whitelist.go`：基于 operation 的白名单选择器

## 当前实现事实

- `whitelist.go` 不是 IP 白名单，而是给 Kratos `selector.MatchFunc` 使用的 operation 白名单
- 支持两种匹配模式：`Exact` 与 `Prefix`
- `WhiteList` 提供 `Add`、`Set`、`Clear`、`Snapshot`、`Merge`、`MatchFunc`

## 使用示例

```go
wl := middleware.NewWhiteList(middleware.Exact, "auth.service.v1.Auth/Login")
selector.Server(authMiddleware).Match(wl.MatchFunc())
```

## 测试

```bash
go test ./pkg/middleware/...
go test ./pkg/middleware/cors/...
```
