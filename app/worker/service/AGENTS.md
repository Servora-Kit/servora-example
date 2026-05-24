# AGENTS.md - app/sayhello/service/

<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-03-15 | Updated: 2026-05-24 -->

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
- Proto 由根 `make api` 统一生成到 `api/gen/go/`
- authn wiring 使用 `authn.Server + authn.Multi + authn.Named`，覆盖 jwt + apikey 双后端；不要添加 `jwt.Server()` wrapper。
