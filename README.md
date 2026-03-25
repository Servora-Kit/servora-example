# Servora IAM

简体中文

> 本项目是 [Servora](https://github.com/Servora-Kit/servora) 框架的**示例项目**，演示如何基于 servora 框架构建 IAM（身份与访问管理）微服务。

`servora-iam` 包含两个微服务（IAM、SayHello）与 IAM 前端应用，展示了 servora 框架在认证、授权、组织管理、项目管理等场景的完整实践。

## 包含内容

### 微服务

- **IAM 服务**（`app/iam/service/`）：身份与访问管理
  - 认证（Authn）：Keycloak OIDC 集成、JWT 签发与验证
  - 授权（Authz）：OpenFGA 细粒度权限控制
  - 组织管理（Organization）
  - 项目管理（Project）
  - 认证/授权中间件位于 `internal/server/middleware/`

- **SayHello 服务**（`app/sayhello/service/`）：独立示例服务，演示基本的 Kratos 微服务结构

### 前端

- **IAM 前端**（`web/iam/`）：IAM 管理界面
- **共享工具库**（`web/pkg/`）：请求处理、Token 管理、Kratos 错误解析
- **共享 UI 组件库**（`web/ui/`）：复用 UI 组件

### 部署

- K8s 清单：`manifests/k8s/iam/`、`manifests/k8s/sayhello/`
- OpenFGA model：`manifests/openfga/`

## 技术栈

- 框架：[servora](https://github.com/Servora-Kit/servora)（Kratos v2）
- API：Protobuf + Buf v2（业务 proto 依赖 [buf.build/servora/servora](https://buf.build/servora/servora)）
- DI：Google Wire
- ORM：Ent
- 认证：Keycloak（OIDC）/ JWT / JWKS
- 授权：OpenFGA
- 存储：PostgreSQL + Redis + Consul
- 前端：React + TypeScript + pnpm workspace

## 项目结构

```text
.
├── api/
│   └── gen/                         # 生成代码（Go + TS，勿手改）
├── app/
│   ├── iam/service/                 # IAM 微服务
│   │   ├── api/protos/              # IAM 业务 proto
│   │   ├── cmd/                     # 服务入口
│   │   ├── configs/                 # 配置文件
│   │   └── internal/                # 业务实现（service/biz/data/server）
│   └── sayhello/service/            # SayHello 示例服务
├── web/
│   ├── iam/                         # IAM 前端应用
│   ├── pkg/                         # 共享前端工具库（@servora/web-pkg）
│   └── ui/                          # 共享 UI 组件库（@servora/ui）
├── manifests/
│   ├── k8s/                         # K8s 部署清单
│   └── openfga/                     # OpenFGA model 与测试
├── buf.yaml                         # Buf v2 workspace（依赖 buf.build/servora/servora）
├── buf.go.gen.yaml                  # Go 代码生成模板
├── buf.typescript.gen.yaml          # TS 代码生成模板
├── docker-compose.yaml              # 基础设施
├── docker-compose.dev.yaml          # 开发环境（iam + sayhello）
├── pnpm-workspace.yaml              # pnpm monorepo
└── Makefile                         # 构建入口
```

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

# 启动前端开发服务器
make dev.web
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

# OpenFGA
make openfga.init           # 初始化 OpenFGA store
make openfga.model.validate # 验证 model
make openfga.model.test     # 测试 model
make openfga.model.apply    # 应用 model 更新
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
- 修改 OpenFGA model 后执行 `make openfga.model.apply`
- 提交前通过 `make lint` 与 `make test`

## License

MIT，详见 `LICENSE`。
