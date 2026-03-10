# IAM Authentication Specification

## ADDED Requirements

### Requirement: The system MUST support user registration with email and password

系统 MUST 支持用户通过邮箱和密码进行注册，注册成功后自动创建关联的 Tenant 和默认 Workspace，并将用户关联到固定的 `platform:root`。

#### Scenario: Successful registration

- **WHEN** 用户提供有效的邮箱、密码和用户名
- **THEN** 系统创建用户记录、创建 Tenant（platform_id=1）、创建默认 Workspace、返回 Access Token 和 Refresh Token

#### Scenario: Registration with duplicate email

- **WHEN** 用户使用已存在的邮箱注册
- **THEN** 系统返回错误 "Email already exists"，HTTP 状态码 409

#### Scenario: Registration with weak password

- **WHEN** 用户提供的密码不符合强度要求（少于 8 个字符）
- **THEN** 系统返回错误 "Password must be at least 8 characters"，HTTP 状态码 400

### Requirement: The system MUST support user login with email and password

系统 MUST 支持用户通过邮箱和密码进行登录，登录成功后签发 JWT Access Token 和 Refresh Token。

#### Scenario: Successful login

- **WHEN** 用户提供正确的邮箱和密码
- **THEN** 系统验证凭证、签发 Access Token（有效期 15 分钟）和 Refresh Token（有效期 7 天）、返回用户信息

#### Scenario: Login with incorrect password

- **WHEN** 用户提供错误的密码
- **THEN** 系统返回错误 "Invalid credentials"，HTTP 状态码 401

#### Scenario: Login with non-existent email

- **WHEN** 用户提供不存在的邮箱
- **THEN** 系统返回错误 "Invalid credentials"，HTTP 状态码 401

#### Scenario: Login with soft-deleted user

- **WHEN** 用户账号已被软删除（deleted_at 不为空）
- **THEN** 系统返回错误 "Account has been deleted"，HTTP 状态码 403

### Requirement: The system MUST issue JWT tokens with kid

系统 MUST 在签发 JWT Token 时包含 `kid`（Key ID）字段，用于支持密钥轮换。

#### Scenario: Access Token contains kid

- **WHEN** 系统签发 Access Token
- **THEN** Token 的 Header 必须包含 `kid` 字段，值为当前使用的密钥 ID

#### Scenario: Access Token contains tenant and workspace claims

- **WHEN** 系统签发 Access Token
- **THEN** Token 的 Payload 必须包含 `tenant_id` 和 `workspace_id` 字段

#### Scenario: Access Token expiration

- **WHEN** Access Token 签发后超过 15 分钟
- **THEN** Token 验证失败，返回错误 "Token expired"

### Requirement: The system MUST enforce a refresh token mechanism

系统 MUST 支持使用 Refresh Token 刷新 Access Token，Refresh Token 存储在 Redis 中。

#### Scenario: Successful token refresh

- **WHEN** 用户提供有效的 Refresh Token
- **THEN** 系统验证 Refresh Token、签发新的 Access Token、返回新 Token

#### Scenario: Refresh with expired token

- **WHEN** Refresh Token 已过期（超过 7 天）
- **THEN** 系统返回错误 "Refresh token expired"，HTTP 状态码 401

#### Scenario: Refresh with revoked token

- **WHEN** Refresh Token 已被撤销（从 Redis 中删除）
- **THEN** 系统返回错误 "Refresh token revoked"，HTTP 状态码 401

### Requirement: The system MUST prevent refresh token replay after rotation

系统 MUST 在 Refresh Token 轮换后立即使旧 token 失效，防止同一 token 被重复使用。

#### Scenario: Old refresh token reuse after rotation

- **WHEN** 用户使用旧 Refresh Token 刷新成功后再次复用同一旧 token
- **THEN** 系统拒绝请求并返回错误 "Refresh token replay detected"，HTTP 状态码 401

### Requirement: The system MUST provide a JWKS endpoint for public key distribution

系统 MUST 提供 JWKS Endpoint（`/.well-known/jwks.json`），供其他服务验证 JWT Token。

#### Scenario: JWKS Endpoint returns valid JWK Set

- **WHEN** 客户端访问 `/.well-known/jwks.json`
- **THEN** 系统返回 JSON 格式的 JWK Set，包含当前所有有效的公钥

#### Scenario: JWKS response includes Cache-Control header

- **WHEN** 客户端访问 JWKS Endpoint
- **THEN** 响应必须包含 `Cache-Control: public, max-age=3600` Header

#### Scenario: JWKS contains multiple keys during rotation

- **WHEN** 系统正在进行密钥轮换
- **THEN** JWKS 必须同时包含旧密钥和新密钥，直到旧密钥过期

### Requirement: The system MUST support key rotation

系统 MUST 支持密钥轮换，采用三阶段流程（分发 → 切换 → 清理），确保服务不中断。

#### Scenario: New key distribution phase

- **WHEN** 管理员执行密钥轮换脚本的分发阶段
- **THEN** 新密钥添加到 JWKS，但系统仍使用旧密钥签发 Token

#### Scenario: Key switch phase

- **WHEN** 管理员执行密钥轮换脚本的切换阶段
- **THEN** 系统开始使用新密钥签发 Token，旧密钥仍保留在 JWKS 中用于验证

#### Scenario: Old key cleanup phase

- **WHEN** 旧密钥保留时间超过 Access Token 最大有效期（15 分钟）
- **THEN** 管理员可以执行清理阶段，从 JWKS 中移除旧密钥

### Requirement: The system MUST provide token verification for other services

系统 MUST 提供 gRPC 接口供其他服务验证 JWT Token。

#### Scenario: Successful token verification

- **WHEN** 其他服务通过 gRPC 调用 `VerifyToken` 并提供有效的 Access Token
- **THEN** 系统验证 Token 签名和有效期、返回用户信息和权限上下文

#### Scenario: Token verification with invalid signature

- **WHEN** 其他服务提供的 Token 签名无效
- **THEN** 系统返回错误 "Invalid token signature"

#### Scenario: Token verification with unknown kid

- **WHEN** 其他服务提供的 Token 包含未知的 `kid`
- **THEN** 系统返回错误 "Unknown key ID"
