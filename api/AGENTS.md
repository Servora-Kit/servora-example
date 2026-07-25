# AGENTS.md - api/

<!-- Parent: ../AGENTS.md -->
<!-- Updated: 2026-07-25 -->

## 目录职责

`api/` 只承载仓库级 Go 生成模块：

- `gen/go.mod`、`gen/go.sum`：生成代码的 Go module 元数据
- `gen/go/`：Master 与 Worker Proto 的 Go 生成产物

服务 Proto 源文件位于各自的 `app/{service}/service/api/protos/`，不在 `api/` 下维护第二份定义。

## 生成规则

| 命令 | 模板 | 输出 | clean |
|------|------|------|-------|
| `make api` | `buf.go.gen.yaml` | `api/gen/go/` | true |
| `make api-go` | `buf.go.gen.yaml` | `api/gen/go/` | true |
| `make openapi` | 各服务 `api/buf.openapi.gen.yaml` | 各服务目录的 `openapi.yaml` | false |

仓库没有 TypeScript 消费方，因此不安装或执行 TypeScript 生成插件。以后服务真正增加 Web 时，由该服务维护自己的 TypeScript 模板和输出目录。

## 开发约定

- 修改任一服务 Proto 后，在仓库根目录运行 `make api`
- 服务目录中的 `make api` 会回到根目录执行统一 Go 生成
- **禁止手动编辑** `api/gen/go/`
- `clean: true` 只清理 `api/gen/go/`，不会删除 `api/gen/go.mod` 或 `api/gen/go.sum`

## 常用命令

```bash
make api          # 生成全部 Go API 代码
make api-go       # 仅执行 Go API 生成目标
make openapi      # 生成各服务 OpenAPI 文档
buf lint          # 检查整个 Buf workspace
```
