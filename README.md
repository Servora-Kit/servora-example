# Servora Example

简体中文

> 本项目是 [Servora](https://github.com/Servora-Kit/servora) 框架的**示例项目**，演示如何基于 servora 框架构建 IAM（身份与访问管理）微服务。

`servora-example` 包含两个微服务（master、worker），展示了 servora 框架在认证、授权、组织管理、项目管理等场景的完整实践。


## 快速开始

### 前置要求

- Go 1.26+
- Node.js 20+ / pnpm
- Make
- Docker / Docker Compose

### 安装工具

```bash
make init    # 安装 protoc 插件、CLI 工具、前端依赖
```

### 生成代码

```bash
make gen     # 统一生成（api + wire + ent + openapi + ts）
```

### 启动开发环境

```bash
# 仅启动基础设施
make compose.up

# 启动基础设施 + 微服务（Air 热重载）
make compose.dev
```

### 常用命令

```bash
# 代码生成
make gen                    # 统一生成
make api                    # 仅生成 proto 代码（Go + TS + AuthZ + Audit）
make wire                   # 仅生成 Wire
make ent                    # 仅生成 Ent

# 质量检查
make test                   # 运行测试
make lint                   # lint.go + lint.ts
make lint.go                # Go lint
make lint.ts                # TS lint
make lint.proto             # Proto lint

# 前端
make pnpm.install           # 安装前端依赖
make dev.web                # 启动前端开发服务器

# Compose
make compose.up             # 启动基础设施
make compose.dev            # 启动开发环境
make compose.stop           # 停止基础设施
make compose.down           # 移除容器/网络（保留数据卷）
make compose.reset          # 移除容器/网络/数据卷
```

## 依赖关系

本项目依赖 servora 核心框架：

- **Go 依赖**：`github.com/Servora-Kit/servora`（基础库）、`github.com/Servora-Kit/servora/api/gen`（框架 proto 生成代码）
- **Proto 依赖**：`buf.build/servora/servora`（框架公共 proto）
- **CLI 工具**：`svr`、`protoc-gen-servora-authz`、`protoc-gen-servora-audit`、`protoc-gen-servora-mapper`

本地联合开发时通过顶层 `go.work` 实现跨仓库引用。

## 质量约束

- 不要手动编辑生成代码：`api/gen/`、`wire_gen.go`、`openapi.yaml`、`*_rules.gen.go`
- 修改 proto 后执行 `make gen`
- 修改 Wire 依赖图后执行 `make wire`
- 提交前通过 `make lint` 与 `make test`

## License

MIT，详见 `LICENSE`。
