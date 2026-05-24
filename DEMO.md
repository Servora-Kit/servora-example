# demo/audit — servora 审计管线演示分支

本分支在 servora-example 上把 servora v0.4.x 的审计管线**端到端**接到 master / worker
两个示例服务，演示 servora `obs/audit` 提供的两条 emit 通道在真实 kratos 服务中的形态。

> ⚠️ 演示性质分支。不打算合入 main，本地用 `go.work` 把 `servora` 依赖指向工作区
> HEAD（v0.4.5+）；用远端依赖（`go.mod` require `v0.3.1`）的 CI 路径会编译失败，
> 这是预期行为。

## 演示了什么

servora 的审计管线提供两条互补的 emit 通道：

| 通道 | 触发位置 | 责任 |
|------|----------|------|
| **Tier 1 — push-ctx + Collector** | 中间件链上写 ctx，链尾 `audit.Collector` 后置阶段读 ctx → emit | 跨切面（authn / authz）由框架统一记录，业务层无需感知 |
| **Tier 2 — Recorder.Emit 直发** | handler / biz 层主动调 `recorder.Emit(...)` 或 `recorder.RecordResourceMutation(...)` | 业务级事件（CRUD、领域动作），由业务代码显式记录 |

本分支 master 演示 Tier 1（无业务事件），worker 演示两条通道叠加。

## 一次 RPC 调用的事件流（实测）

`curl 'http://127.0.0.1:8001/v1/hello?greeting=demo'` 返回
`{"reply":"master relay -> worker says hello, demo"}`，两个服务的 stdout 各自出
audit 事件（servora zap logger，dev 模式 console + 颜色码；下面节选 prod 模式
JSON 输出便于阅读）：

```
master.service.log:
2026-05-08T17:23:35.067+0900  info  log/helper.go:122  {"service":"master.service",
  "module":"audit/emitter/log",
  "msg":"audit_event event_id=70b8668e... type=authn.result
   service=master.service operation=/servora.master.service.v1.MasterService/Hello
   payload={...\"authnDetail\":{\"method\":\"demo\",\"success\":true}...}"}

worker.service.log:
2026-05-08T17:23:35.067+0900  info  log/helper.go:122  {"service":"worker.service",
  "module":"audit/emitter/log",
  "msg":"audit_event event_id=2a8e4abd... type=resource.mutation
   service=worker.service operation=/servora.worker.service.v1.WorkerService/Hello
   payload={...\"resourceMutationDetail\":{\"mutationType\":\"RESOURCE_MUTATION_TYPE_CREATE\",
   \"resourceType\":\"hello.reply\",\"resourceId\":\"demo\"}...}"}

2026-05-08T17:23:35.067+0900  info  log/helper.go:122  {"service":"worker.service",
  "module":"audit/emitter/log",
  "msg":"audit_event event_id=5380dc67... type=authn.result
   service=worker.service operation=/servora.worker.service.v1.WorkerService/Hello
   payload={...\"authnDetail\":{\"method\":\"demo\",\"success\":true}...}"}
```

3 条事件共享同一 `traceId`，说明 OTel trace 与 audit 正确关联。worker 端
`resource.mutation`（Tier 2 直发）先于 `authn.result`（Tier 1 LIFO 后置）出，
印证 v0.4.5 装配契约：handler 内直发立即生效，Collector 在 handler 返回后才读
ctx → emit。

## 关键改动

### Worker (`app/worker/service/`)

- `internal/server/audit.go` — `ProvideAuditEmitter` / `ProvideAuditRecorder` wire provider
- `internal/server/demo_identity.go` — `demoIdentityMiddleware()`：每次请求往 ctx
  写一条固定 `AuthnDetail{Method:"demo", Success:true}`，让 `audit.Collector` 后置阶段
  能采集到 AUTHN 事件
- `internal/server/server.go` — `ProviderSet` 加上两个 audit provider
- `internal/server/grpc.go` — 链上 `.WithAudit(rec)` + append `demoIdentityMiddleware()`
- `internal/service/worker.go` — `NewWorkerService` 注入 `*audit.Recorder`；`Hello`
  内调用 `s.rec.RecordResourceMutation(...)` 演示 Tier 2

### Master (`app/master/service/`)

- 与 worker 同样的 `audit.go` / `demo_identity.go` / `server.go` ProviderSet 改动
- `internal/server/grpc.go` 与 `internal/server/http.go` 链上 `.WithAudit(rec)` + append
- 不动 service / biz 层（master 仅演示 Tier 1）

### Wire 代码

`make wire`（或 `wire ./cmd/server/...`）已重新生成两个服务的 `wire_gen.go`，注入链
为 `Runtime/Logger → Emitter → Recorder → Server`，服务身份由 `Runtime.NewApp` 注入。

## 如何运行

### 一次性准备

```bash
# 在 servora-kit 根目录起基础设施
cd servora-kit && make compose.up.infra   # consul / jaeger / otel-collector
```

### 两个终端跑服务

```bash
# 终端 A — worker
cd servora-kit/servora-example/app/worker/service
make dev   # air 热重载

# 终端 B — master
cd servora-kit/servora-example/app/master/service
make dev
```

### 触发请求

```bash
curl --location --request GET 'http://127.0.0.1:8001/v1/hello?greeting=demo'
```

### 在两个服务输出中查找

audit 事件喂给 servora 服务主 logger（`logger.Logger` → `audit.NewLogEmitter`）。
具体输出位置取决于 `app.env`：

| `app.env` | logger sink | 在哪看 audit 事件 |
|-----------|-------------|------------------|
| `"dev"`（默认 local 配置） | console-only stdout | air dev / `make run` 终端 stdout 直接看 |
| `"prod"` | tee stdout + `./logs/<svc>.service.log` | 终端或 log 文件均可 |

> ℹ️ servora `obs/logging/log.go` 在 `env == "dev"` 分支里只构造 console core
> （`zapcore.NewConsoleEncoder + os.Stdout`），bootstrap.yaml 里的 `filename` /
> `max_size` / `max_backups` 字段在 dev 时**全部被忽略**——logs/ 目录恒空。
> 不是 bug，是设计选择。想让本地也写文件，把 bootstrap.yaml 的 `app.env` 改成
> `"prod"`。

```bash
# dev 模式直接看终端 stdout（或重定向后的文件）
grep audit_event /path/to/master.stdout /path/to/worker.stdout

# prod 模式还可以看 log 文件
grep audit_event app/master/service/logs/master.service.log \
                 app/worker/service/logs/worker.service.log
```

应看到 3 行（master 1 条 authn.result，worker 1 条 authn.result + 1 条
resource.mutation），共享同一 `traceId`。生产场景下若要把 audit 与服务主日志解耦
（避免 audit 受主 logger 配置波及），应换成 `BrokerEmitter`（写 Kafka / NATS）
或自定义 Emitter。

## 等 P0-4（proto-driven authn）落地后怎么演进

`demoIdentityMiddleware` 是临时 fixture。等 `protoc-gen-servora-authn` 插件 + 真
authn middleware 落地后，把这一行：

```go
mw = append(mw, demoIdentityMiddleware())
```

替换为：

```go
mw = append(mw, authn.Server(
    authn.Multi(
        authn.Named(authjwt.Scheme, authjwt.NewAuthenticator(authjwt.WithVerifier(verifier))),
        authn.Named(apikey.Scheme, apikey.NewAuthenticator(apikey.WithStore(keyStore))),
    ),
    authn.WithRulesFuncs(examplev1.AuthnRules),
))
```

链路其它部分（`.WithAudit(rec)` 装配位置、Tier 2 业务级 emit）保持不变 ——
这是 v0.4.4 / v0.4.5 设计的接入摩擦保证。

## 已知限制

- LogEmitter 写入 servora logger（即 kratos `log.Logger` 包装），观测时建议 `tail -f`
  各服务 `logs/` 下的 log 文件，或 air dev 模式下直接看终端 stdout。
- 远端 CI 用 `GOWORK=off` 走 servora `v0.3.1`，**不含** `LogEmitter` / `WithAudit`，
  所以本分支远端编译失败。这是设计内的 — 不要给本分支单独升 `go.mod` require 来"修"
  CI，直到 servora `v0.4.x` tag 推送到 GitHub。
