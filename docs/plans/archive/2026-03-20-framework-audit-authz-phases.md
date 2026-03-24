# 归档：Servora 框架演进 Phase 1–3.5（认证、授权、审计）

**原文档：** `docs/plans/2026-03-20-keycloak-openfga-audit-design.md`
**归档日期：** 2026-03-24
**覆盖阶段：** Phase 1（框架骨架）→ Phase 2a（审计 emit）→ Phase 2b（Audit Service）→ Phase 3（代码生成 + 去特化）→ Phase 3.5（authn/authz 接口化）

---

## 进度总览

| 阶段 | 名称 | OpenSpec |
|------|------|---------|
| Phase 1 | 框架骨架 | `openspec/changes/archive/2026-03-20-framework-audit-skeleton/` |
| Phase 2a | 审计 emit 接入（pkg 层） | `openspec/changes/archive/2026-03-22-audit-emit-integration/` |
| Phase 2b | Audit Service + ClickHouse | `openspec/changes/archive/2026-03-20-audit-service-clickhouse/` |
| Phase 3 | all-in-proto 代码生成 + pkg 去特化 | `openspec/changes/archive/2026-03-23-proto-codegen-audit-authz/` |
| Phase 3.5 | authn/authz 接口化 + 可插拔引擎 | `openspec/changes/archive/2026-03-24-pkg-auth-pluggable/` |

**已沉淀 specs（21 个）：** `openspec/specs/` 下的 actor-v2、audit-clickhouse-storage、audit-codegen-integration、audit-kafka-consumer、audit-proto、audit-query-api、audit-runtime、audit-service-scaffold、authn-interface、authz-audit-emit、authz-interface、broker-abstraction、config-proto-extension、identity-header-enhancement、infra-kafka-clickhouse、logger-refactor、openfga-audit-emit、openfga-framework-api、pkg-despecialization、proto-package-governance、protoc-gen-servora-audit。

---

## Phase 1：框架骨架

> 详细设计、spec 与实现索引见 `openspec/changes/archive/2026-03-20-framework-audit-skeleton/`。

| 交付物 | 关键决策 |
|--------|---------|
| pkg/logger v2 | 暴力重构：`New(app)` / `For(l,"mod")` / `With(l,"mod")` / `Zap()` / `Sync()` |
| Actor v2 | 扩展为完整身份模型（Subject/ClientID/Realm/Roles/Scopes/Attrs），新增 ServiceActor |
| IdentityFromHeader v2 | 8 种 gateway header → Actor v2，支持 WithHeaderMapping |
| audit.proto + annotations.proto | AuditEvent、4 typed detail、AuditRule method option |
| conf.proto 扩展 | Kafka（含 SASL）、ClickHouse、Audit 配置 |
| pkg/broker + kafka | franz-go（非 sarama）；Broker/Event/Subscriber/MiddlewareFunc 接口 |
| pkg/audit 骨架 | Emitter → Recorder → Middleware；3 种 emitter（Noop/Log/Broker） |
| 基础设施 | Kafka KRaft + ClickHouse；IAM/sayhello 从工具链移除 |

---

## Phase 2a：审计 emit 接入（pkg 层）

> 详细设计、spec 与实现索引见 `openspec/changes/archive/2026-03-22-audit-emit-integration/`。

| 交付物 | 关键决策 |
|--------|---------|
| pkg/openfga API 框架化 | `Check`/`ListObjects`/`CachedCheck` 参数从 `userID` 改为 `user`（完整 principal），移除 `"user:"` 硬编码 |
| pkg/openfga ClientOption 模式 | `NewClient(cfg, opts...)` + `WithAuditRecorder` + `WithComputedRelations` |
| pkg/openfga core/public 分层 | tuple 操作拆为 core + public wrapper，成功后自动 emit `tuple.changed` |
| pkg/openfga CachedCheck 扩展 | 返回值新增 `cacheHit` |
| pkg/authz audit 集成 | `WithAuditRecorder(r)` option + Check 后自动 emit `authz.decision` |
| app/iam/service 适配 | 全局适配 openfga 去特化 API |
| e2e 验证 | LogEmitter JSON 输出 + BrokerEmitter Kafka round-trip |

---

## Phase 2b：Audit Service + ClickHouse

> 详细设计、spec 与实现索引见 `openspec/changes/archive/2026-03-20-audit-service-clickhouse/`。

| 交付物 | 关键决策 |
|--------|---------|
| `app/audit/service` 微服务 | 严格分层（service→biz→data）；Kafka consumer → ClickHouse → 查询 API |
| ClickHouse 存储 | 官方 native driver `clickhouse-go/v2`；DDL 内嵌 `CREATE TABLE IF NOT EXISTS` |
| `pkg/db/clickhouse` | 框架级连接 helper `NewConnOptional`，Optional-init 模式 |
| 查询 API | `ListAuditEvents`（cursor 分页 + 多维筛选）、`CountAuditEvents`；gRPC + HTTP 转码 |
| E2E 验证 | sayhello → Kafka → audit service → ClickHouse → 查询 API 全链路 |

---

## Phase 3：all-in-proto 代码生成 + pkg 去特化

> 详细设计、spec 与实现索引见 `openspec/changes/archive/2026-03-23-proto-codegen-audit-authz/`。

| 交付物 | 关键决策 |
|--------|---------|
| `cmd/protoc-gen-servora-audit` | 审计注解代码生成器 |
| `protoc-gen-servora-authz` 改造 | `func AuthzRules()` 返回 copy（不可变） |
| `pkg/authz` 去特化 | 动态 principal 构造；拒绝 anonymous 而非仅接受 user |
| `pkg/actor` 去特化 | 删除业务 scope 常量和便捷方法 |
| `pkg/transport/middleware/scope` 可配置化 | `ScopeFromHeaders(bindings ...ScopeBinding)` |

---

## Phase 3.5：authn/authz 接口化 + 可插拔引擎

> 详细设计、spec 与实现索引见 `openspec/changes/archive/2026-03-24-pkg-auth-pluggable/`。

| 交付物 | 关键决策 |
|--------|---------|
| `pkg/authn` 接口化 | `Authenticator` 接口 + `Server()` 中间件 |
| `pkg/authn/jwt/` 引擎 | ClaimsMapper 体系：DefaultClaimsMapper + KeycloakClaimsMapper |
| `pkg/authn/noop/` 引擎 | 返回 anonymous actor |
| `pkg/authz` 接口化 | `Authorizer` 接口 + `Server()` 中间件 |
| `pkg/authz/openfga/` 引擎 | 封装 openfga.Client + 可选 Redis 缓存 |
| `pkg/authz/noop/` 引擎 | 总是放行 |
| 审计解耦 | `WithAuditRecorder` → `WithDecisionLogger` 回调模式 |
| `AuthzMode` proto 迁移 | 从 IAM 服务 proto 移至共享 proto |

---

## 实现约束（Phase 1–3.5 确立）

1. **Optional-init 模式统一**：可选基础设施组件使用 `NewXxxOptional`，nil 配置返回 nil
2. **Proto 集中配置**：框架级配置通过 `servora/conf/v1/conf.proto` 统一管理
3. **Logger 桥接模式**：第三方库通过 `logger.Zap()` 获取底层 `*zap.Logger`
4. **Module 命名规范**：`component/layer/service` 格式
5. **broker 接口扩展点**：新增实现只需实现 `broker.Broker` interface
6. **OpenSpec 主 spec 格式**：必须包含 Purpose/Requirements/Scenario
7. **Proto 包治理规范**：`servora.*` package，目录对齐，go_package 落到 `api/gen/go/servora/**`
8. **pkg 框架包去特化原则**：不含业务特化逻辑；authn/authz 仅定义接口，引擎在子目录
9. **ClientOption 模式**：`NewClient(cfg, opts...)` 模式
10. **core/public 分层模式**：cross-cutting concern 拆为 core + public wrapper
11. **Kafka 双 listener**：PLAINTEXT (9092) + EXTERNAL (29092)
12. **Kafka topic 预创建**：开发环境 auto-create，生产环境需 admin API/init job
13. **nil Timestamp 空值处理**：接收 proto timestamp 参数必须先判 nil
14. **Data 层结构统一**：`Data` struct + `NewData` 函数
15. **分层依赖与接口返回**：service→biz→data，data 构造函数返回 biz 接口
16. **Codegen 不可变规则**：unexported var + exported func 返回 copy
17. **pkg/actor 去特化原则**：无业务 scope 常量，通用 API only
18. **ScopeFromHeaders 可配置化**：`ScopeBinding` 切片参数
19. **SystemActor ID 调用方提供**：无自动前缀
20. **authz principal 动态构造**：`string(Type())+":"+ID()`，引擎无需关心
21. **authn/authz 接口驱动架构**：接口 + Server() + 引擎子目录
22. **AuthzMode 共享 proto**：定义在 `servora/authz/v1/authz.proto`

---

## pkg 生态状态（Phase 3.5 完成后快照）

| 包 | 状态 | 说明 |
|----|------|------|
| `pkg/actor` | ✅ v2 | 通用 principal 模型，4 种 actor type |
| `pkg/authn` | ✅ 接口化完成 | `Authenticator` 接口 + `Server()` 中间件；引擎：`jwt/`、`noop/` |
| `pkg/authz` | ✅ 接口化完成 | `Authorizer` 接口 + `Server()` 中间件；引擎：`openfga/`、`noop/`；`WithDecisionLogger` 回调 |
| `pkg/audit` | ✅ 主链已接入 | Recorder + LogEmitter/BrokerEmitter |
| `pkg/broker` | ✅ 接口 + kafka | franz-go |
| `pkg/db/clickhouse` | ✅ 框架级封装 | `NewConnOptional` |
| `pkg/logger` | ✅ v2 | 暴力重构后的简洁 API |
| `pkg/openfga` | ✅ 框架化完成 | ClientOption、去特化、core/public 分层、tuple audit emit |
| `pkg/transport` | ✅ 可用 | 中间件体系 |

---

## 已执行变更索引

| 组件 | 操作 | 阶段 |
|------|------|------|
| IAM/sayhello 工具链入口 | 从 Makefile 移除 | Phase 1 |
| IAM/sayhello 源代码 | 保留作为参考模板 | 保留 |
| `pkg/actor` | v2 破坏性升级 | Phase 1 |
| `pkg/logger` | 暴力重构 | Phase 1 |
| `pkg/openfga` | API 框架化 + ClientOption + audit emit | Phase 2a |
| `pkg/authz` | 接口化 + DecisionLogger 解耦 | Phase 2a → 3.5 |
| `pkg/audit` | 主链接入 + e2e 验证 | Phase 2a |
| `app/audit/service` | 新建审计微服务 | Phase 2b |
| `pkg/db/clickhouse` | 新建框架级连接 helper | Phase 2b |
| `pkg/authn` 接口化 | Authenticator 接口 + jwt/noop 引擎 | Phase 3.5 |
| `pkg/authz` 接口化 | Authorizer 接口 + openfga/noop 引擎 | Phase 3.5 |
| `AuthzMode` proto | 从 IAM 移至共享 proto | Phase 3.5 |

---

## 参考项目

| 项目 | 参考内容 | 不参考内容 |
|------|---------|-----------|
| kratos-transport | broker 接口设计、option 组织、middleware 模式 | 整套外部抽象边界、直接作为依赖 |
| Kemate | docker-compose 配置（Kafka KRaft）、optional-init 模式 | sarama 选型 |
