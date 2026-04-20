# Implementation Plan

## Feature 信息

- Feature ID: `SL-ADMIN-WIN-DEV-BOOTSTRAP`
- 开发负责人: 二号开发者

## 变更范围

- 代码文件:
  - `backend/internal/handler/handler.go`
  - `backend/internal/middleware/auth.go`
  - `backend/internal/router/router.go`
  - `backend/internal/auth/jwt.go`
- 文档文件:
  - `docs/features/SL-ADMIN-WIN-DEV-BOOTSTRAP/*`
- 配置文件:
  - `.env.example`
- 部署/脚本文件:
  - `scripts/windows/start-backend.ps1`
  - `scripts/windows/start-frontend.ps1`

## 实施步骤

1. 为管理端后端补齐 JWT 生成与解析能力。
2. 将登录接口从 mock token 升级为本地开发可用的默认管理员登录。
3. 新增获取当前管理员信息与登出接口，接上路由和鉴权中间件。
4. 增加 Windows 独立启动脚本，避免影响非 Windows 同事命令。
5. 补充 feature 文档与本地环境变量说明。

## 验证计划

- `go build ./...`
- `pnpm build`
- 需要的人工验证:
  - 启动后访问 `/health`
  - 调用登录接口
  - 携带 token 调用 `/api/v1/admin/info`

## 风险与回滚关注点

- 默认管理员能力仅用于本地开发，不能误认为生产方案。
- Windows 启动脚本必须保持隔离，不应污染跨平台默认命令。
