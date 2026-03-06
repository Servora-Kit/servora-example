# AGENTS.md - cmd/svr CLI

<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-03-03 | Updated: 2026-03-06 -->

## 目录定位

`cmd/svr/` 是仓库内统一开发 CLI，当前重点命令为：
- `svr gen gorm`
- `svr new api`

该工具默认假设 **从项目根目录运行**。

## 当前结构

```text
cmd/svr/
├── main.go
└── internal/
    ├── cmd/
    │   ├── gen/
    │   └── new/
    ├── discovery/
    ├── generator/
    ├── root/
    ├── scaffold/
    └── ux/
```

## 命令说明

### `svr gen gorm`
- 支持多服务参数
- 无参数时进入 `huh` 交互选择
- `--dry-run` 只输出路径，不连数据库
- 批量失败不立即中断，最终统一汇总
- 发现与配置校验逻辑在 `internal/discovery/`

### `svr new api`
- 输入必须是小写 snake_case，可带点分层级
- 默认模板目录：`api/protos/template/service/v1`
- 默认输出目录：`api/protos/`
- 若目标目录已存在，直接报错退出

## 当前实现事实

- `main.go` 只调用 `root.Execute()`，失败时 `os.Exit(1)`
- `new/api.go` 通过字符串替换处理 `template` / `Template` / `TEMPLATE`
- `gen/gorm.go` 定义 4 类失败：`service-not-found`、`config-invalid`、`db-connect-failed`、`generation-failed`
- `discovery.ListAvailableServices()` 依据 `app/*/service` 扫描可用服务

## 常用命令

```bash
go run ./cmd/svr gen gorm servora
go run ./cmd/svr gen gorm servora --dry-run
go run ./cmd/svr new api billing.invoice
go run ./cmd/svr new api user --template /custom/templates
```

## 维护提示

- 文档示例必须以项目根目录为基准，不要写成在服务目录执行 `go run ./cmd/svr ...`
- `svr new api` 现在只负责生成共享模块骨架；服务专属 proto 仍可按项目约定放到服务自己的 `api/protos/`
