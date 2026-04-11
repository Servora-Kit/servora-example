# AGENTS.md - servora-example

<!-- Updated: 2026-04-11 -->

## 项目概览

`servora-example` 是 [Servora](https://github.com/Servora-Kit/servora) 框架的**示例项目**，包含 master 和 worker 两个最小化实现的微服务，演示 Servora 框架的服务注册、服务发现、链路追踪等核心能力。

## 仓库结构

```
servora-example/
├── app/
│   ├── master/service/         # master 微服务（HTTP + gRPC + TCP 注册）
│   │   ├── cmd/server/         # 入口
│   │   ├── configs/local/      # 本地开发配置（air 读取）
│   │   ├── configs/docker/     # 容器部署配置
│   │   ├── internal/           # 业务逻辑（biz/data/service）
│   │   ├── .air.toml           # air 热重载配置，入口 ./configs/local/
│   │   └── Makefile            # 使用 app.mk，包含 dev/build/wire 等目标
│   └── worker/service/         # worker 微服务（gRPC）
│       ├── configs/local/
│       ├── configs/docker/
│       └── .air.toml
├── api/
│   ├── protos/                 # Proto 定义
│   └── gen/go/                 # buf generate 产物（勿手动修改）
├── docker-compose.yaml         # 基础设施：consul / jaeger / otel-collector
├── docker-compose.apps.yaml    # 应用容器：master / worker
├── Dockerfile                  # 多阶段构建，GOWORK=off，静态编译
└── Makefile                    # 根 Makefile
```

## 端口约定

### 本地热重载（`configs/local/`）

| 服务           | HTTP  | gRPC  | TCP（服务注册）|
|----------------|-------|-------|--------------|
| master         | 8001  | 8000  | 8002         |
| worker         | —     | 8010  | —            |
| Consul         | 8500  | —     | —            |
| Jaeger UI      | 16686 | —     | —            |
| OTel Collector | —     | 4317  | —            |

### 全容器化（`configs/docker/`，宿主机端口）

| 服务           | HTTP  | gRPC  |
|----------------|-------|-------|
| master         | 8001  | 8000  |
| worker         | —     | 8010  |
| Consul         | 8500  | —     |
| Jaeger UI      | 16686 | —     |
| OTel Collector | —     | 4317  |

## 配置约定

### 本地开发（`configs/local/`）

- `bootstrap.yaml`：服务名称、logger、注册中心地址（`localhost:8500`）
- `biz.yaml`：业务参数
- `tcp.yaml`（master）：TCP server 配置，`registry.endpoint` 填 `tcp://host.docker.internal:<port>`，供容器化 Consul 回连时使用

> ⚠️ `tcp.yaml` 中不要设置 `registry.host`，用 `registry.endpoint` 完整 URL 代替。

### 容器部署（`configs/docker/`）

- 与 `local/` 结构相同，`bootstrap.yaml` 中注册中心地址改为 `consul:8500`
- 由 `docker-compose.apps.yaml` 通过 volume mount 注入

### 环境变量

本地开发不依赖 `.env`，容器部署通过 `docker-compose.apps.yaml` 的 `environment` 注入（如需覆盖配置）。

## 框架版本

| 模块                                      | 版本   |
|-------------------------------------------|--------|
| `github.com/Servora-Kit/servora`          | v0.3.1 |
| `github.com/Servora-Kit/servora/api/gen`  | v0.3.1 |

## 常用 Makefile 目标

### 根目录（`servora-example/`）

| 目标                  | 说明                                       |
|-----------------------|--------------------------------------------|
| `make compose.up.infra` | 启动基础设施容器（Consul/Jaeger/OTel）       |
| `make compose.up.all`  | 启动基础设施 + 应用容器                     |
| `make compose.build`   | 构建所有微服务 Docker 镜像                  |
| `make compose.rebuild` | 重新构建镜像并启动基础设施                  |
| `make compose.down`    | 停止并移除所有容器/网络                     |
| `make compose.reset`   | 停止并移除所有容器/网络/volumes             |
| `make compose.ps`      | 查看基础设施容器状态                        |
| `make compose.logs`    | 跟踪基础设施容器日志                        |
| `make gen`             | 生成全部代码（proto / openapi / wire / ent）|
| `make api-go`          | 仅生成 Go proto 代码（`buf.go.gen.yaml`）  |
| `make tidy`            | 对所有模块执行 `go mod tidy`               |
| `make test`            | 运行所有模块测试                            |
| `make lint.go`         | Go lint                                     |

### 服务目录（`app/{master,worker}/service/`）

| 目标         | 说明                          |
|--------------|-------------------------------|
| `make dev`   | air 热重载启动（读 `configs/local/`）|
| `make wire`  | 生成 wire DI 代码             |
| `make build` | 编译二进制                    |
| `make openapi` | 生成 OpenAPI 文档            |

## 开发工作流

### 本地热重载开发

> 首次开发前需在 `servora-example/` 根目录执行 `make init`，安装 air、wire、buf、golangci-lint 等所有 CLI 工具。

```bash
# 终端 0：启动基础设施
cd servora-example && make compose.up.infra

# 终端 A：启动 worker
cd servora-example/app/worker/service && make dev

# 终端 B：启动 master
cd servora-example/app/master/service && make dev

# 验证
curl --location --request GET 'http://127.0.0.1:8013/v1/hello?greeting=hello'
```

### 修改 Proto

```bash
cd servora-example
make api-go      # 重新生成 api/gen/go/
make wire        # 重新生成 wire 代码（如接口有变化）
```

### 全容器化体验

```bash
cd servora-example
make compose.build && make compose.up.all
curl --location --request GET 'http://127.0.0.1:8001/v1/hello?greeting=hello'
```

## 提交规范

格式：`type(scope): description`

- type：`feat` / `fix` / `refactor` / `docs` / `test` / `chore`
- scope 示例：`master` / `worker` / `api` / `docker` / `config`

## 注意事项

- `api/gen/go/` 为 buf 生成产物，**不要手动修改**
- `Dockerfile` 使用 `GOWORK=off` + CGO 禁用静态编译，确保容器内可独立运行
- `go.work` 仅存在于顶层 `servora-kit/` 工作区，本仓库 CI 独立构建时不使用 `go.work`
- 修改框架依赖版本后，在 `servora-kit/` 根目录执行 `make sync` 同步 `go.work` replace 指令
