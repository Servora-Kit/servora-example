# Servora Example

简体中文

> 本项目是 [Servora](https://github.com/Servora-Kit/servora) 框架的**示例项目**，演示如何基于 servora 框架构建微服务。

`servora-example` 包含两个微服务（master、worker），展示了 servora 框架的基本用法和功能。

## 快速开始

### 前置要求

- Go 1.26+
- Make
- Docker / Docker Compose

### 安装工具

```bash
make init    # 安装 protoc 插件与 CLI 工具
```

### 生成代码

```bash
make gen     # 统一生成（api + wire）
```

### 启动开发环境

Compose 负责基础设施，应用通过 `make run` 在本机启动：

```bash
# 启动基础设施
make compose.up

# 终端 A：启动 worker
cd app/worker/service && make run

# 终端 B：启动 master
cd app/master/service && make run
```

worker 读取 `./configs/local/`，gRPC 监听 `0.0.0.0:8010`；master 读取 `./configs/local/`，HTTP 监听 `0.0.0.0:8001`。

### 验证服务

```bash
curl --location --request GET 'http://127.0.0.1:8001/v1/hello?greeting=hello'
```

### 常用命令

```bash
# 代码生成
make gen                    # 统一生成
make api                    # 仅生成 proto 代码
make wire                   # 仅生成 Wire

# 质量检查
make lint                   # Go lint
make lint.proto             # Proto lint

# 服务目录（app/master/service/、app/worker/service/）
make run                    # 直接运行（读 configs/local/）
make build                  # 编译二进制

# Compose - 基础设施
make compose.up             # 启动基础设施
make compose.stop           # 停止容器，不删除容器
make compose.down           # 移除容器/网络（保留数据卷）
make compose.reset          # 移除容器/网络/数据卷
make compose.ps             # 查看 Compose 服务状态
make compose.logs           # 跟踪 Compose 服务日志

# OpenFGA
make openfga.init           # 初始化 store
make openfga.model.validate # 验证 model
make openfga.model.test     # 测试 model
make openfga.model.apply    # 应用 model 更新
```

## 端口一览

| 服务    | HTTP | gRPC |
|---------|------|------|
| master  | 8001 | 8000 |
| worker  | —    | 8010 |
