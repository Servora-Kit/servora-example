# IAM 统一身份平台重构 — 实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 IAM 服务从后台管理系统重构为统一身份平台，移除不属于 IAM 的模块，增强身份与认证能力。

**Architecture:** IAM 只负责身份管理、认证服务、授权基础设施。Tenant/Org/Dict/Position/RBAC 全部移除，User + Application 作为核心实体。authz.proto 泛化为通用注解。

**Tech Stack:** Go/Kratos, Ent ORM, zitadel/oidc, OpenFGA, protobuf/buf, React/TanStack Router/shadcn-ui/Catppuccin

**Design doc:** `docs/plans/2026-03-18-iam-identity-platform-design.md`

---

## Phase 1: Proto 层清理与泛化

### Task 1: 删除已废弃的 proto 目录

**Files:**
- Delete: `app/iam/service/api/protos/tenant/` (整个目录)
- Delete: `app/iam/service/api/protos/organization/` (整个目录)
- Delete: `app/iam/service/api/protos/dict/` (整个目录)
- Delete: `app/iam/service/api/protos/position/` (整个目录)
- Delete: `app/iam/service/api/protos/rbac/` (整个目录)
- Delete: `app/iam/service/api/protos/test/` (整个目录)
- Delete: `app/iam/service/api/protos/iam/service/v1/i_tenant.proto`
- Delete: `app/iam/service/api/protos/iam/service/v1/i_organization.proto`
- Delete: `app/iam/service/api/protos/iam/service/v1/i_dict.proto`
- Delete: `app/iam/service/api/protos/iam/service/v1/i_position.proto`
- Delete: `app/iam/service/api/protos/iam/service/v1/i_rbac.proto`
- Delete: `app/iam/service/api/protos/iam/service/v1/i_test.proto`

**Step 1:** 删除上述所有文件和目录。

**Step 2:** 确认 `app/iam/service/api/protos/iam/service/v1/iam_doc.proto` 中是否 import 了已删除的 proto，如果是则移除 import。

**Step 3:** Commit

```bash
git add -A && git commit -m "chore(api/proto): 移除 tenant/org/dict/position/rbac/test proto 定义"
```

### Task 2: 泛化 authz.proto

**Files:**
- Modify: `app/iam/service/api/protos/authz/service/v1/authz.proto`

**Step 1:** 编辑 `authz.proto`：
- `AuthzMode` 枚举简化为 `UNSPECIFIED`、`NONE`、`CHECK`（移除 `ORGANIZATION` 和 `OBJECT`）
- 移除 `ObjectType` 枚举（整个删除）
- 移除 `Relation` 枚举（整个删除）
- `AuthzRule` message 中 `relation` 和 `object_type` 改为 `string` 类型，`mode` 保持 `AuthzMode` 枚举

改后的完整 `AuthzRule` 相关部分：

```protobuf
enum AuthzMode {
  AUTHZ_MODE_UNSPECIFIED = 0;
  AUTHZ_MODE_NONE = 1;
  AUTHZ_MODE_CHECK = 2;
}

message AuthzRule {
  AuthzMode mode = 1;
  string relation = 2;
  string object_type = 3;
  string id_field = 4;
}
```

保留 `extend google.protobuf.MethodOptions` 和 `AuthzService` / `CheckPermission` RPC。

**Step 2:** Commit

```bash
git add -A && git commit -m "refactor(api/proto): 泛化 authz.proto — 枚举改字符串，移除 Tenant/Org 特定类型"
```

### Task 3: 更新 user.proto 和 application.proto

**Files:**
- Modify: `app/iam/service/api/protos/user/service/v1/user.proto`
- Modify: `app/iam/service/api/protos/iam/service/v1/i_user.proto`
- Modify: `app/iam/service/api/protos/application/service/v1/application.proto`
- Modify: `app/iam/service/api/protos/iam/service/v1/i_application.proto`

**Step 1:** `user.proto` 变更：
- 添加 `username` 字段（string）
- 添加 `phone`、`phone_verified` 字段
- 添加 `status` 字段（string: "active" | "disabled"）
- 添加 `profile` 字段（google.protobuf.Struct 或 string/JSON）
- 移除任何 `organization_ids` 相关字段
- 移除任何 `tenant_id` 相关字段

**Step 2:** `i_user.proto` 变更：
- 移除任何 tenant scope 相关的请求参数
- 更新 CreateUser / UpdateUser / ListUsers 的 request/response 以匹配新字段

**Step 3:** `application.proto` 变更：
- 添加 `type` 字段（string: "web" | "native" | "m2m"）
- 移除 `tenant_id` 字段

**Step 4:** `i_application.proto` 变更：
- 移除任何 tenant scope 相关的请求参数

**Step 5:** 更新各 proto 中的 `authz.service.v1.rule` 注解，使用新的字符串语法：
- 原来引用 `OBJECT_TYPE_TENANT` 的改为简单的 `mode: AUTHZ_MODE_CHECK, object_type: "platform", relation: "admin"` 或 `mode: AUTHZ_MODE_NONE`
- 需要管理员权限的接口用 `object_type: "platform", relation: "admin"`

**Step 6:** Commit

```bash
git add -A && git commit -m "refactor(api/proto): 更新 user/application proto — 移除租户依赖，添加新字段"
```

### Task 4: 重新生成 protobuf 代码

**Step 1:** 更新 `protoc-gen-servora-authz` 插件适配新的字符串字段

**Files:**
- Modify: `cmd/protoc-gen-servora-authz/main.go`

生成的 `AuthzRuleEntry` struct 改为：

```go
type AuthzRuleEntry struct {
    Mode       authzpb.AuthzMode
    Relation   string
    ObjectType string
    IDField    string
}
```

对应修改 `generateFile` 函数中的输出逻辑。

**Step 2:** 运行 `make api`（生成 Go + AuthZ 规则代码）

```bash
make api
```

Expected: 编译成功，`api/gen/go/` 下的代码更新，已删 proto 的生成代码消失。

**Step 3:** Commit

```bash
git add -A && git commit -m "refactor(cmd/protoc-gen-servora-authz): 适配 authz 字符串字段并重新生成代码"
```

---

## Phase 2: Ent Schema 清理

### Task 5: 删除已废弃的 Ent Schema

**Files:**
- Delete: `app/iam/service/internal/data/schema/tenant.go`
- Delete: `app/iam/service/internal/data/schema/organization.go`
- Delete: `app/iam/service/internal/data/schema/organization_member.go`
- Delete: `app/iam/service/internal/data/schema/dict_type.go`
- Delete: `app/iam/service/internal/data/schema/dict_item.go`
- Delete: `app/iam/service/internal/data/schema/position.go`
- Delete: `app/iam/service/internal/data/schema/rbac_role.go`
- Delete: `app/iam/service/internal/data/schema/rbac_permission.go`
- Delete: `app/iam/service/internal/data/schema/rbac_permission_group.go`
- Delete: `app/iam/service/internal/data/schema/rbac_permission_api.go`
- Delete: `app/iam/service/internal/data/schema/rbac_permission_menu.go`
- Delete: `app/iam/service/internal/data/schema/rbac_menu.go`
- Delete: `app/iam/service/internal/data/schema/rbac_role_permission.go`
- Delete: `app/iam/service/internal/data/schema/rbac_user_role.go`

**Step 1:** 删除上述 14 个 schema 文件。

**Step 2:** Commit（先提交删除，再做修改，避免混乱）

```bash
git add -A && git commit -m "chore(app/iam): 删除 tenant/org/dict/position/rbac Ent schema"
```

### Task 6: 修改保留的 Ent Schema

**Files:**
- Modify: `app/iam/service/internal/data/schema/user.go`
- Modify: `app/iam/service/internal/data/schema/application.go`

**Step 1:** `user.go` 变更：
- 添加 `username` 字段（string, unique, max 64）
- 添加 `phone` 字段（string, optional, max 32）
- 添加 `phone_verified` 字段（bool, default false）
- 添加 `status` 字段（string, max 32, default "active"）
- 添加 `profile` 字段（JSON, 类型为 `map[string]interface{}`，可选）
- 移除 `role` 字段的 `MaxLen(32).Default("user")`，改为 `MaxLen(32).Default("user")`（保持不变）
- 移除 Edges 中的 `org_memberships` 和 `owned_tenants`
- Edges 返回空切片

**Step 2:** `application.go` 变更：
- 移除 `tenant_id` 字段
- 移除 Edges 中的 tenant edge
- Edges 返回空切片

**Step 3:** 重新生成 Ent 代码

```bash
cd app/iam/service && make gen.ent
```

Expected: 编译成功，`data/ent/` 下的代码重新生成，不再有 tenant/org/rbac 等实体。

**Step 4:** Commit

```bash
git add -A && git commit -m "refactor(app/iam): 更新 user/application schema 并重新生成 Ent 代码"
```

---

## Phase 3: Biz / Data / Service 层清理

### Task 7: 删除已废弃的 Biz 层

**Files:**
- Delete: `app/iam/service/internal/biz/tenant.go`
- Delete: `app/iam/service/internal/biz/organization.go`
- Delete: `app/iam/service/internal/biz/roles.go`（如存在）
- Delete: `app/iam/service/internal/biz/position.go`
- Delete: `app/iam/service/internal/biz/dict.go`
- Delete: `app/iam/service/internal/biz/authz.go`
- Delete: `app/iam/service/internal/biz/entity/tenant.go`
- Delete: `app/iam/service/internal/biz/entity/organization.go`
- Delete: `app/iam/service/internal/biz/entity/rbac.go`
- Delete: `app/iam/service/internal/biz/entity/position.go`
- Delete: `app/iam/service/internal/biz/entity/dict.go`
- Delete: `app/iam/service/internal/biz/tenant_test.go`
- Delete: `app/iam/service/internal/biz/roles_test.go`
- Delete: `app/iam/service/internal/biz/application_test.go`（需要重写，先删）
- Modify: `app/iam/service/internal/biz/biz.go`

**Step 1:** 删除上述文件。

**Step 2:** 修改 `biz/biz.go` ProviderSet：

```go
var ProviderSet = wire.NewSet(NewAuthnUsecase, NewUserUsecase, NewApplicationUsecase)
```

**Step 3:** 修改 `biz/entity/user.go`：
- 添加 `Username`, `Phone`, `PhoneVerified`, `Status`, `Profile` 字段
- 移除 `OrganizationIDs`

**Step 4:** 修改 `biz/entity/application.go`：
- 移除 `TenantID`
- 添加 `Type` 字段（string）

**Step 5:** 修改 `biz/application.go`（ApplicationRepo 接口）：
- `GetByID` — 移除 `tenantID` 参数
- `ListByTenantID` — 改为 `List(ctx, page, pageSize)` 无 tenant 过滤
- `Update` — 移除 `tenantID` 参数
- `Delete` — 移除 `tenantID` 参数
- `UpdateClientSecretHash` — 移除 `tenantID` 参数
- `ApplicationUsecase` 中对应方法同步修改

**Step 6:** Commit

```bash
git add -A && git commit -m "refactor(app/iam): 清理 biz 层 — 移除 tenant/org/dict/position/rbac usecase"
```

### Task 8: 清理 Data 层

**Files:**
- Delete: `app/iam/service/internal/data/rbac.go`
- Delete: `app/iam/service/internal/data/dict.go`
- Delete: `app/iam/service/internal/data/position.go`
- Delete: `app/iam/service/internal/data/authz.go`
- Delete: `app/iam/service/internal/data/seed_rbac_data.go`
- Delete: `app/iam/service/internal/data/seed.go`
- Modify: `app/iam/service/internal/data/data.go`
- Modify: `app/iam/service/internal/data/application.go`

**Step 1:** 删除上述文件。

**Step 2:** 修改 `data/data.go` ProviderSet — 移除所有已删 repo 的 New 函数：

```go
var ProviderSet = wire.NewSet(
    registry.NewDiscovery, NewEntDriver, NewDBClient, NewRedis, NewData,
    NewAuthnRepo, NewUserRepo, NewOTPRepo, NewMailSender,
    NewApplicationRepo, NewOIDCStorage,
)
```

**Step 3:** 修改 `data/application.go` — 移除所有 `tenantID` 参数和 tenant 过滤逻辑。

**Step 4:** 检查 `data/` 目录下是否有引用已删 ent 实体（如 `ent.Tenant`、`ent.Organization`）的其他文件，修复编译错误。重点检查：
- `data/user.go`（可能查询 org memberships）
- `data/oidc_storage.go`（可能引用 tenant）

**Step 5:** Commit

```bash
git add -A && git commit -m "refactor(app/iam): 清理 data 层 — 移除已废弃 repo 和 tenant 依赖"
```

### Task 9: 清理 Service 层

**Files:**
- Delete: `app/iam/service/internal/service/tenant.go`
- Delete: `app/iam/service/internal/service/organization.go`（如存在则搜索确认文件名）
- Delete: `app/iam/service/internal/service/dict.go`
- Delete: `app/iam/service/internal/service/position.go`
- Delete: `app/iam/service/internal/service/test.go`
- Modify: `app/iam/service/internal/service/service.go`
- Modify: `app/iam/service/internal/service/application.go`

**Step 1:** 删除上述文件。

**Step 2:** 修改 `service/service.go` ProviderSet：

```go
var ProviderSet = wire.NewSet(NewAuthnService, NewUserService, NewApplicationService)
```

**Step 3:** 修改 `service/application.go`：
- 移除 `requireTenantScope` 调用
- `CreateApplication` / `ListApplications` 不再需要 tenant 上下文

**Step 4:** 搜索 `service/` 目录下是否有 `requireTenantScope` 或类似的 scope helper 函数定义，若有则删除。

**Step 5:** Commit

```bash
git add -A && git commit -m "refactor(app/iam): 清理 service 层 — 移除已废弃服务实现"
```

---

## Phase 4: Server 层与中间件

### Task 10: 更新 Server 注册与中间件

**Files:**
- Modify: `app/iam/service/internal/server/http.go`
- Modify: `app/iam/service/internal/server/grpc.go`
- Modify: `app/iam/service/internal/server/server.go`
- Modify: `app/iam/service/internal/server/middleware/authz.go`

**Step 1:** `http.go` — `NewHTTPServer` 函数：
- 移除参数：`org`, `tenant`, `rbac`, `position`, `dict`（对应的 `*service.XxxService`）
- 移除 `WithServices` 中的：`RegisterOrganizationServiceHTTPServer`, `RegisterTenantServiceHTTPServer`, `RegisterRbacServiceHTTPServer`, `RegisterPositionServiceHTTPServer`, `RegisterDictServiceHTTPServer`, `RegisterTestServiceHTTPServer`
- `NewHTTPMiddleware` 中：移除 `ScopeFromHeaders()` 中间件（不再有 tenant/org scope），移除白名单中已删服务的 operation
- 简化 `publicWhitelist`

**Step 2:** `grpc.go` — `NewGRPCServer` 函数：
- 移除参数：`org`, `tenant`
- 移除注册：`orgpb.RegisterOrganizationServiceServer`, `tenantpb.RegisterTenantServiceServer`, `testpb.RegisterTestServiceServer`
- 移除 `remapAuthzRulesForGRPC` 函数中对已删 service 的处理（或整体简化）
- 移除已删的 import（`orgpb`, `tenantpb`, `testpb`）
- `NewGRPCMiddleware` 中：移除 `ScopeFromHeaders()` 和白名单中已删 operation

**Step 3:** `middleware/authz.go`：
- `AuthzRuleEntry` 类型改为使用生成代码中的新类型（`Mode: AuthzMode`, `Relation: string`, `ObjectType: string`）
- 移除 `resolveObject` 中的 `AUTHZ_MODE_ORGANIZATION` 分支
- 简化为只处理 `AUTHZ_MODE_NONE`（跳过）和 `AUTHZ_MODE_CHECK`（通用 check）
- 移除 `scopeFromActor` 函数
- `AUTHZ_MODE_CHECK` 分支：直接用 `rule.ObjectType` 和 `rule.Relation` 作为字符串，从 `rule.IDField` 提取 object ID
- 移除 `relationToFGA` 和 `objectTypeToFGA` 转换函数（已经是字符串了）

**Step 4:** `server.go` ProviderSet 不需要改动（不依赖已删模块）。

**Step 5:** Commit

```bash
git add -A && git commit -m "refactor(app/iam): 更新 server 层 — 移除已删服务注册，简化 authz 中间件"
```

---

## Phase 5: pkg/ 层清理

### Task 11: 清理 pkg/actor

**Files:**
- Modify: `pkg/actor/user.go`

**Step 1:** 移除 scope key 常量中的 `ScopeKeyTenantID`、`ScopeKeyOrganizationID`、`ScopeKeyProjectID`（或保留为通用 scope key 供未来其他平台使用 — 取决于是否有其他 pkg 引用它们）。

**Step 2:** 移除便捷方法 `TenantID()`, `OrganizationID()`, `ProjectID()`, `SetTenantID()`, `SetOrganizationID()`, `SetProjectID()`。

**Step 3:** 确认 `pkg/transport/server/middleware/` 中的 `ScopeFromHeaders` 是否需要更新（移除对 tenant/org header 的读取）。

**Step 4:** Commit

```bash
git add -A && git commit -m "refactor(pkg/actor): 移除 Tenant/Org/Project scope 便捷方法"
```

---

## Phase 6: Wire 重新生成与编译验证

### Task 12: 重新生成 Wire 代码并编译

**Step 1:** 重新生成 wire 代码

```bash
cd app/iam/service && make wire
```

Expected: `wire_gen.go` 重新生成，不再引用已删模块。

**Step 2:** 编译验证

```bash
cd app/iam/service && go build ./...
```

Expected: 编译成功，无错误。

**Step 3:** 若编译失败，根据错误信息修复遗漏（可能是某些文件中还引用了已删实体的 import 或函数调用）。

**Step 4:** Commit

```bash
git add -A && git commit -m "chore(app/iam): 重新生成 wire 代码并通过编译验证"
```

---

## Phase 7: OpenFGA Model 更新

### Task 13: 重写 OpenFGA model

**Files:**
- Modify: `manifests/openfga/model/servora.fga`
- Modify: `manifests/openfga/tests/servora.fga.yaml`

**Step 1:** 重写 `servora.fga`：

```fga
model
  schema 1.1

type user

type platform
  relations
    define admin: [user]

type service
  relations
    define caller: [service]
```

**Step 2:** 更新测试文件，编写新 model 的测试用例。

**Step 3:** 验证

```bash
make openfga.model.validate
make openfga.model.test
```

Expected: 验证和测试通过。

**Step 4:** Commit

```bash
git add -A && git commit -m "refactor(manifests/openfga): 重写 model — 仅保留 user/platform/service"
```

---

## Phase 8: 前端清理（web/iam/）

### Task 14: 删除已废弃的前端页面和组件

**Files (delete):**
- `web/iam/src/routes/_app/tenants/` (整个目录)
- `web/iam/src/routes/_app/organizations/` (整个目录)
- `web/iam/src/routes/_app/rbac/` (整个目录)
- `web/iam/src/routes/_app/positions/` (整个目录)
- `web/iam/src/routes/_app/system/` (整个目录)
- `web/iam/src/routes/_app/settings/roles.tsx`
- `web/iam/src/routes/_app/settings/security.tsx`
- `web/iam/src/components/org-context-picker.tsx`
- `web/iam/src/components/scope-switcher.tsx`
- `web/iam/src/stores/scope.ts`
- `web/iam/src/stores/scope.test.ts`
- `web/iam/src/stores/access.ts`
- `web/iam/src/hooks/use-permissions.ts`

**Step 1:** 删除上述文件和目录。

**Step 2:** 修改 `web/iam/src/layout/sidebar.tsx` — 移除已删页面的导航项。

**Step 3:** 修改 `web/iam/src/router.tsx` / `routeTree.gen.ts` — 重新生成路由树（`pnpm --filter iam run generate-routes` 或对应命令）。

**Step 4:** 检查是否有其他组件 import 了已删文件，修复编译错误。

**Step 5:** 验证前端能正常启动

```bash
pnpm --filter iam run dev
```

Expected: 无编译错误，能正常访问 dashboard / users / applications / settings 页面。

**Step 6:** Commit

```bash
git add -A && git commit -m "refactor(web/iam): 移除 tenant/org/rbac/dict/position 页面和组件"
```

---

## Phase 9: 端到端验证

### Task 15: 全栈启动验证

**Step 1:** 重置开发环境（清除旧数据库 schema）

```bash
make compose.reset
make compose.dev
```

**Step 2:** 确认以下功能正常：
- [ ] IAM 服务启动无错误
- [ ] 用户注册（邮箱 + 密码）
- [ ] 用户登录（邮箱 + 密码）→ 返回 token
- [ ] Token 刷新
- [ ] 当前用户信息查询
- [ ] 应用创建（不需要 tenant scope）
- [ ] 应用列表
- [ ] OIDC 登录流程（authorize → login → callback → token）
- [ ] JWKS 端点（`/.well-known/jwks.json`）
- [ ] 前端管理界面可访问

**Step 3:** 记录任何问题并修复。

**Step 4:** Commit（如有修复）

```bash
git add -A && git commit -m "fix(app/iam): 修复端到端验证中发现的问题"
```

---

## Phase 10: 新增能力（后续独立实施）

以下任务作为独立的后续实施计划，不在本次重构中执行：

- **Task A:** M2M Client Credentials Grant — OIDC Storage 实现 `ClientCredentialsTokenRequest`
- **Task B:** OpenFGA service 间授权 — model 中 `type service` 的关系写入与 check 流程
- **Task C:** `web/accounts/` — 独立登录前端应用
- **Task D:** `web/ui/` — 共享 shadcn/ui + Catppuccin 组件包
- **Task E:** `pkg/authn` — 通用 JWT 验签中间件（从 IAM internal 上提）
- **Task F:** `pkg/authz` — 通用 OpenFGA check 中间件（从 IAM internal 上提）
