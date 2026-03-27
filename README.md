# Servora Example

简体中文

> 本项目是 [Servora](https://github.com/Servora-Kit/servora) 框架的**示例项目**，演示如何基于 servora 框架构建 IAM（身份与访问管理）微服务。

`servora-example` 包含两个微服务（master、worker），展示了 servora 框架的基本用法和功能。

## 快速开始

以下流程使用 1 个容器化 Consul + 2 个本地进程（`master`、`worker`）启动整套示例。

### 1) 准备环境

```bash
cd servora-example
make init
cp .env.example .env
```

编辑 `.env`，至少填写本机局域网 IP（Consul 容器需要通过这个地址回连本地服务）：

```bash
# 示例：替换为你的本机局域网 IP
G_REGISTRY_HOST=192.168.31.71
```

> `G_REGISTRY_HOST` 会写入服务注册信息中的 `server.grpc.registry.host`。若不填写，容器化 Consul 可能无法访问到 `master` / `worker` 的健康检查端点。

### 2) 启动 Consul（容器）

```bash
cd servora-example
make compose.up
```

启动后可访问 Consul UI：<http://localhost:8500>

### 3) 本地启动 worker（终端 A）

```bash
cd servora-example/app/worker/service
make run
```

默认会读取 `./configs/local/`，并注册到 `localhost:8500` 的 Consul。

### 4) 本地启动 master（终端 B）

```bash
cd servora-example/app/master/service
make run
```

`master` 会通过服务发现调用 `worker.service`。

### 5) 验证服务

- `worker` gRPC 监听：`0.0.0.0:8011`
- `master` gRPC 监听：`0.0.0.0:8012`
- `master` HTTP 监听：`0.0.0.0:8013`
- Consul 中可看到 `master.service` 与 `worker.service` 已注册

### 6) 停止环境

停止本地服务进程后，在 `servora-example` 根目录执行：

```bash
make compose.down
```
