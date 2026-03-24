# 设计文档：Servora Keycloak 接入

**日期：** 2026-03-24
**前身：** 2026-03-20 认证/授权/审计框架演进（Phase 1–3.5 已归档）
**状态：** 规划中

---

## 前置工作归档

Phase 1–3.5（框架骨架、审计全链路、代码生成、authn/authz 接口化）已全部完成，详见：
- **归档文档**：`docs/plans/archive/2026-03-20-framework-audit-authz-phases.md`
- **OpenSpec 归档**：`openspec/changes/archive/` 下的 5 个已归档 change
- **沉淀 specs（21 个）**：`openspec/specs/` 目录

---

## 背景

Servora 框架层的认证、授权、审计基础设施已就绪：
- `pkg/authn`：可插拔 `Authenticator` 接口，现有 `jwt/` 和 `noop/` 引擎
- `pkg/authz`：可插拔 `Authorizer` 接口，现有 `openfga/` 和 `noop/` 引擎
- `pkg/audit`：全链路审计（Kafka → Audit Service → ClickHouse）
- `pkg/actor`：通用 principal 模型（user/service/anonymous/system）

下一步是接入 **Keycloak** 作为认证中心，完成从自建 IAM 到外部 IdP 的切换。

---

## 核心决策

| 决策点 | 结论 |
|---|---|
| 认证中心 | **Keycloak** |
| 网关认证 | 网关统一验 token → 注入 principal header → 业务服务信任 header |
| 网关选型 | 先用 **Traefik**，但保持可插拔（配置化，不硬绑） |
| 业务服务验 JWT | 默认**不重复验**，信任网关 header；`pkg/authn/jwt/` 引擎保留用于需要直接验 JWT 的场景 |
| IAM 服务 | **保留**作为示例服务，不清理 |
| 前端 | **暂不对接**，当前无需运行前端 |
| actor 模型 | `Keycloak claims → 网关 header → HeaderAuthenticator → actor.Actor`，业务服务只依赖 actor |

---

## 职责分工

```
┌──────────┐    OIDC     ┌──────────┐   principal   ┌────────────────┐
│ Keycloak │◄───────────►│  网关     │──headers──────►│  业务服务       │
│          │  验 token    │(Traefik) │               │ pkg/authn/header│
└──────────┘             └──────────┘               │ → actor.Actor  │
                                                     │ pkg/authz      │
                                                     │ → OpenFGA      │
                                                     │ pkg/audit      │
                                                     │ → Kafka        │
                                                     └────────────────┘
```

- **Keycloak**：用户认证、OIDC/OAuth2、token 签发、JWKS、realm/client/role 管理
- **网关**：统一入口、对接 Keycloak 验 token、将 principal 注入上游 header
- **业务服务**：从 header 构建 actor → `pkg/authz` 本地授权 → OpenFGA → 审计 emit

---

## 分阶段计划

### Phase 1：Keycloak 基础设施

**目标**：开发环境一键启动 Keycloak，OIDC endpoints 可用。

**核心任务**：
1. docker-compose 新增 Keycloak 服务（`quay.io/keycloak/keycloak`，`start-dev` 模式）
2. 创建 realm 初始化文件 `manifests/keycloak/servora-realm.json`：
   - `servora` realm
   - OAuth2 client（如 `servora-gateway`、`servora-web`）
   - 测试用户（admin、普通用户）
   - 基本 realm roles 配置
3. 使用 `--import-realm` 挂载到 `/opt/keycloak/data/import/` 实现自动初始化
4. 验证 OIDC discovery endpoint (`/.well-known/openid-configuration`) 和 JWKS endpoint 可用

**不做**：
- 不对接网关
- 不改动 pkg 代码
- 不修改现有服务

### Phase 2：pkg/authn/header/ 引擎

**目标**：实现 `HeaderAuthenticator`，从网关注入的 header 构造 `actor.Actor`。

**核心任务**：
1. 创建 `pkg/authn/header/` 子目录
2. 实现 `HeaderAuthenticator` 实现 `authn.Authenticator` 接口
3. 从 header 映射 actor 字段：
   - `X-User-ID` → `Actor.ID()`
   - `X-Subject` → `Actor.Subject()`
   - `X-Client-ID` → `Actor.ClientID()`
   - `X-Principal-Type` → `Actor.Type()`
   - `X-Realm` → `Actor.Realm()`
   - `X-Email` → `Actor.Email()`
   - `X-Roles` → `Actor.Roles()`
   - `X-Scopes` → `Actor.Scopes()`
4. 支持 `WithHeaderMapping` 自定义 header 名称
5. `X-Principal-Type` 决定 actor 类型（user/service/anonymous）

**已有 spec**：`openspec/specs/identity-header-enhancement/spec.md`（已更新为 HeaderAuthenticator 方向）

**不做**：
- 不迁移或删除现有代码
- 不修改网关配置

### Phase 3：网关认证集成

**目标**：网关对接 Keycloak，完成 token 验证 → principal header 注入 → 业务服务 authn 的完整链路。

**核心任务**：
1. Traefik 配置对接 Keycloak OIDC（ForwardAuth 或 OIDC plugin）
2. 验证链路：用户登录 → Keycloak 签发 token → 请求带 token → 网关验证 → 注入 principal header → 业务服务 `authn.Server(headerAuth)` → actor in context
3. 保持网关可插拔：认证配置不硬编码在业务服务或 pkg 中

**依赖**：Phase 1（Keycloak 可用）+ Phase 2（HeaderAuthenticator 可用）

**不做**：
- 不清理 IAM 中的 issuer 能力
- 不对接前端

---

## 未来方向

- **Servora 生态扩展**：`pkg/broker` 补更多实现（NATS/RabbitMQ）、`pkg/task`/`pkg/queue` 任务队列、统一 observability
- **前端对接**：Keycloak 登录流程（当需要前端时再规划）
- **IAM 演进**：保留作为示例服务，可能逐步演化为管理控制台

---

## 约束继承

本阶段继承 Phase 1–3.5 确立的所有实现约束（详见归档文档 `docs/plans/archive/2026-03-20-framework-audit-authz-phases.md` 中的"实现约束"章节）。Keycloak 接入相关的新约束将在各 Phase 实现时补充。
