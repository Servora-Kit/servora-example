# servora-example

Master/Worker 基础示例，用于演示服务注册、发现、追踪和 Audit；具体 AuthN/AuthZ 能力在 `plateau` 实践。

## 目录

- `app/master/service/`：HTTP、gRPC、TCP 示例
- `app/worker/service/`：gRPC 示例
- `api/gen/go/`：生成代码
- `make/`：共享 Make 逻辑

本示例不装配 JWT/API Key Authenticator，不保留 sign-token、credential stub 或 AuthN scheme 注解。

## 命令

```bash
make init
make gen
make lint.go
make compose.up
make compose.build
```

服务目录使用 `make dev`、`make wire`、`make build`。生成代码由 `make gen` 维护；容器独立构建使用 `GOWORK=off`。

提交格式：`type(scope): description`。
