# AGENTS.md - app/

<!-- Parent: ../AGENTS.md -->
<!-- Updated: 2026-07-25 -->

## 目录概览

`app/` 当前包含两个独立 Go module：

- `app/master/service/`：HTTP、gRPC 与 TCP 注册示例
- `app/worker/service/`：gRPC Worker 示例

两个模块都通过根 `go.work` 纳管。

## 服务共同结构

```text
app/{master,worker}/service/
├── api/
│   ├── protos/                 # 服务 Proto 源文件
│   └── buf.openapi.gen.yaml    # 服务自有 OpenAPI 模板
├── cmd/                        # 启动入口
├── configs/                    # local/docker 配置
├── internal/                   # 业务实现
├── Makefile                    # include ../../../make/core.mk
└── go.mod                      # 独立模块
```

## 关键约定

- 服务目录中的 `make gen` 执行 `api + openapi + wire + gen.ent`
- 服务目录中的 `make build` 先执行 `make gen`，再编译当前服务
- 服务目录中的 `make api` 回到仓库根目录执行统一 Go API 生成
- 服务级 `make openapi` 读取本服务 `api/buf.openapi.gen.yaml`
- 当前没有 TypeScript 生成模板；空的 `app/master/web/` 不构成 TypeScript 消费方

## 常用命令

```bash
cd app/master/service && make run
cd app/master/service && make build
cd app/master/service && make wire
cd app/worker/service && make run
```
