# AGENTS.md - web/iam

<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-03-15 -->

## 目录定位

`web/iam` 是 IAM 的前端应用（Vite + React + TanStack Router + TanStack Query），与 `app/iam/service` 后端对接。

## 业务模型与类型

**如需使用业务/数据模型（请求体、响应体、列表项、分页等），请优先使用代码生成类型，不要手写类型。**

- 生成代码有两处，均由根目录 `make api-ts` 生成：
  - **公共 API 类型**（`api/gen/ts/`）：分页等跨前端复用类型，在 web/iam 中通过 `#/api-gen/*` 引用。
  - **IAM 服务**（`src/service/gen/`）：HTTP 客户端与 IAM 领域类型。
- 可按包引用，例如：
  - `#/api-gen/pagination/v1`：分页请求/响应（公共，优先用此）
  - `#/service/gen/iam/service/v1`：HTTP 客户端与 IAM 相关类型
  - `#/service/gen/authn/service/v1`：认证相关类型
  - `#/service/gen/organization/service/v1`：组织模型
  - `#/service/gen/user/service/v1`：用户模型
- 发请求使用单例 `iamClients`（`#/api`），见 `src/api.ts` 与 `src/service/request/`.

## 常用命令

- 开发：`pnpm dev`
- 构建：`pnpm build`
- 生成 TS 客户端与类型（在仓库根目录执行）：`make api-ts`
