# IAM User Management Specification

## ADDED Requirements

### Requirement: The system MUST support user CRUD operations

系统 MUST 支持用户的创建、查询、更新和软删除操作。

#### Scenario: Create user by admin

- **WHEN** Tenant admin 创建用户并提供 email、username 和 password
- **THEN** 系统创建用户记录、将用户添加到当前 Tenant、返回用户信息

#### Scenario: Get user by ID

- **WHEN** 用户请求查询用户信息并提供 user_id
- **THEN** 系统返回用户的基本信息（不包含密码）

#### Scenario: Update user profile

- **WHEN** 用户更新自己的 username、avatar 或 bio
- **THEN** 系统更新用户记录、返回更新后的信息

#### Scenario: Update user email

- **WHEN** 用户更新邮箱
- **THEN** 系统验证新邮箱未被使用、发送验证邮件、更新邮箱后标记为未验证

#### Scenario: List users in tenant

- **WHEN** Tenant member 查询 Tenant 内的用户列表
- **THEN** 系统返回所有 Tenant 成员的用户信息（分页）

### Requirement: The system MUST support user cross-tenant association

系统 MUST 支持用户跨多个 Tenant 的关联，一个用户可以属于多个 Tenant。

#### Scenario: User joins multiple tenants

- **WHEN** 用户被邀请加入第二个 Tenant
- **THEN** 系统创建新的 TenantMember 记录、用户可以在多个 Tenant 之间切换

#### Scenario: List user's tenants

- **WHEN** 用户查询自己所属的 Tenant 列表
- **THEN** 系统返回用户有成员关系的所有 Tenant

#### Scenario: Switch active tenant

- **WHEN** 用户切换到另一个 Tenant
- **THEN** 系统签发新的 JWT Token（包含新的 tenant_id 和默认 workspace_id）

### Requirement: The system MUST support soft delete user

系统 MUST 支持用户的软删除，软删除后用户无法登录但数据保留。

#### Scenario: User soft delete by admin

- **WHEN** Tenant admin 软删除用户
- **THEN** 系统设置 `deleted_at` 字段、撤销用户在当前 Tenant 的所有权限、保留用户数据

#### Scenario: User self-delete

- **WHEN** 用户请求删除自己的账号
- **THEN** 系统设置 `deleted_at` 字段、撤销用户在所有 Tenant 的权限、保留用户数据

#### Scenario: Soft deleted user cannot login

- **WHEN** 软删除的用户尝试登录
- **THEN** 系统返回错误 "Account has been deleted"，HTTP 状态码 403

#### Scenario: Restore soft deleted user

- **WHEN** 系统管理员恢复软删除的用户
- **THEN** 系统清除 `deleted_at` 字段、恢复用户状态为 active

### Requirement: The system MUST support hard delete user

系统 MUST 支持用户的硬删除，物理删除用户数据和所有关联关系。

#### Scenario: User hard delete by system admin

- **WHEN** 系统管理员执行硬删除用户
- **THEN** 系统物理删除用户记录、删除所有 TenantMember 和 WorkspaceMember 记录、删除所有 OAuthAccount 记录、删除 OpenFGA 关系元组

#### Scenario: Hard delete cascades to owned resources

- **WHEN** 用户被硬删除且该用户是某些资源的唯一 owner
- **THEN** 系统返回错误 "Cannot delete user: user is the only owner of resources"，HTTP 状态码 400

#### Scenario: Hard delete removes all traces

- **WHEN** 用户被硬删除
- **THEN** 系统确保用户数据完全删除，满足 GDPR 要求

### Requirement: The system MUST support user password management

系统 MUST 支持用户密码的修改和重置。

#### Scenario: Change password

- **WHEN** 用户提供当前密码和新密码
- **THEN** 系统验证当前密码、更新密码哈希、撤销所有 Refresh Token

#### Scenario: Change password with incorrect current password

- **WHEN** 用户提供错误的当前密码
- **THEN** 系统返回错误 "Current password is incorrect"，HTTP 状态码 401

#### Scenario: Reset password via email

- **WHEN** 用户请求重置密码并提供邮箱
- **THEN** 系统发送重置链接到邮箱（包含临时 token，有效期 1 小时）

#### Scenario: Complete password reset

- **WHEN** 用户通过重置链接提供新密码
- **THEN** 系统验证 token、更新密码哈希、撤销所有 Refresh Token

### Requirement: The system MUST support user email verification

系统 MUST 支持邮箱验证机制。

#### Scenario: Send verification email on registration

- **WHEN** 用户完成注册
- **THEN** 系统发送验证邮件（包含验证链接，有效期 24 小时）

#### Scenario: Verify email

- **WHEN** 用户点击验证链接
- **THEN** 系统验证 token、标记邮箱为已验证、返回成功

#### Scenario: Resend verification email

- **WHEN** 用户请求重新发送验证邮件
- **THEN** 系统生成新的验证 token、发送新邮件

#### Scenario: Access restricted features without verification

- **WHEN** 未验证邮箱的用户尝试访问需要验证的功能
- **THEN** 系统返回错误 "Email verification required"，HTTP 状态码 403

### Requirement: The system MUST support user status management

系统 MUST 支持用户状态管理（active、inactive、suspended）。

#### Scenario: Suspend user

- **WHEN** Tenant admin 暂停用户
- **THEN** 系统设置用户状态为 suspended、撤销所有 Refresh Token、用户无法登录

#### Scenario: Suspended user cannot login

- **WHEN** 被暂停的用户尝试登录
- **THEN** 系统返回错误 "Account has been suspended"，HTTP 状态码 403

#### Scenario: Reactivate suspended user

- **WHEN** Tenant admin 重新激活用户
- **THEN** 系统设置用户状态为 active、用户可以正常登录

### Requirement: The system MUST support user search and filtering

系统 MUST 支持用户的搜索和过滤。

#### Scenario: Search users by email

- **WHEN** Tenant admin 搜索用户并提供邮箱关键词
- **THEN** 系统返回邮箱包含关键词的用户列表

#### Scenario: Search users by username

- **WHEN** Tenant admin 搜索用户并提供用户名关键词
- **THEN** 系统返回用户名包含关键词的用户列表

#### Scenario: Filter users by status

- **WHEN** Tenant admin 过滤用户并指定状态（active/inactive/suspended）
- **THEN** 系统返回指定状态的用户列表

#### Scenario: Filter users by role in tenant

- **WHEN** Tenant admin 过滤用户并指定角色（owner/admin/member）
- **THEN** 系统返回在当前 Tenant 中具有指定角色的用户列表
