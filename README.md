# Servora Example

简体中文

> 本项目是 [Servora](https://github.com/Servora-Kit/servora) 框架的**示例项目**，演示如何基于 servora 框架构建微服务。

`servora-example` 包含两个微服务（master、worker），展示了 servora 框架的基本用法和功能。

## 快速开始

### 场景一：快速体验（全容器化）

使用 Docker 一键启动基础设施 + 应用容器，无需本地 Go 环境。

**前置条件：** Docker Desktop / Docker Engine + Docker Compose

#### 1) 构建应用镜像

```bash
cd servora-example
make compose.build
```

#### 2) 启动全部服务

```bash
make compose.up.all
```

启动后可访问：

- Consul UI：<http://localhost:8500>（可查看 master/worker 注册状态）
- Jaeger UI：<http://localhost:16686>（可查看调用链路）

#### 3) 验证服务

```bash
curl --location --request GET 'http://127.0.0.1:8001/v1/hello?greeting=hello'
```

#### 4) 停止并清理

```bash
make compose.down    # 停止并移除容器/网络
# 或
make compose.reset   # 同上，并清除 volumes
```

---

### 场景二：快速开发（本地热重载）

使用 [air](https://github.com/air-verse/air) 热重载在本地开发，基础设施仍通过 Docker 运行。

**前置条件：** Go 1.23+、Docker

#### 1) 初始化开发环境（首次）

```bash
cd servora-example
make init
```

> `make init` 会安装所有 CLI 工具，包括 air、wire、buf、golangci-lint 等。

#### 2) 启动基础设施容器

```bash
cd servora-example
make compose.up.infra
```

启动后可访问 Consul UI：<http://localhost:8500>

#### 3) 启动 worker（终端 A）

```bash
cd servora-example/app/worker/service
make dev
```

worker 读取 `./configs/local/`，gRPC 监听 `0.0.0.0:8010`，注册到 `localhost:8500`。

#### 4) 启动 master（终端 B）

```bash
cd servora-example/app/master/service
make dev
```

master 读取 `./configs/local/`，通过服务发现调用 worker，HTTP 监听 `0.0.0.0:8001`。

#### 5) 验证服务

```bash
curl --location --request GET 'http://127.0.0.1:8001/v1/hello?greeting=hello'
```

在 Jaeger UI 中查询 `master.service` 或 `worker.service` 可查看完整调用链路。

#### 6) 停止环境

停止本地进程后，回到 `servora-example` 根目录：

```bash
make compose.down
```

## 端口一览

### 本地热重载（场景二）

| 服务    | HTTP  | gRPC  | TCP（注册）|
|---------|-------|-------|-----------|
| master  | 8001  | 8000  | 8002      |
| worker  | —     | 8010  | —         |
| Consul  | 8500  | —     | —         |
| Jaeger  | 16686 | —     | —         |

### 全容器化（场景一）

| 服务    | HTTP（宿主机:容器）| gRPC（宿主机:容器）|
|---------|-------------------|-------------------|
| master  | 8001:8001         | 8000:8000         |
| Consul  | 8500              | —                 |
| Jaeger  | 16686             | —                 |

## 目录结构

```
servora-example/
├── app/
│   ├── master/service/     # master 微服务
│   │   ├── configs/local/  # 本地开发配置
│   │   ├── configs/docker/ # 容器环境配置
│   │   └── .air.toml       # air 热重载配置
│   └── worker/service/     # worker 微服务
│       ├── configs/local/
│       ├── configs/docker/
│       └── .air.toml
├── api/                    # Proto 定义及生成代码
├── docker-compose.yaml     # 基础设施（Consul/Jaeger/OTel）
├── docker-compose.apps.yaml # 应用容器
└── Makefile
```
