# SL-ADMIN-PRIVATE-COS-PREVIEW Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-PRIVATE-COS-PREVIEW
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/handler/private_file.go`
  - `backend/internal/router/router.go`
  - `frontend/src/services/api.ts`
  - `frontend/src/pages/user/index.tsx`
- 文档文件:
  - `docs/features/SL-ADMIN-PRIVATE-COS-PREVIEW/*`
- 配置文件:
  - `.env.example`
  - `.env.production`
- 部署/脚本文件: 无

## 实施步骤

1. 增加私有 COS URL 解析、bucket 白名单校验和短期 GET 签名接口。
2. 将接口挂到现有管理端 Auth 保护路由下。
3. 前端用户审核详情打开时解析身份证正反面私有 URL，并使用签名 URL 展示图片。
4. 补充生产环境需要注入的 COS 配置。
5. 执行 Go 测试/构建、前端构建、review/release gate。

## 验证计划

- `cd backend && go test ./...`
- `cd backend && go build -o /tmp/steellink-admin-api ./cmd/server`
- `cd frontend && pnpm build`
- `./harness/review.sh docs/features/SL-ADMIN-PRIVATE-COS-PREVIEW`
- `./harness/release.sh docs/features/SL-ADMIN-PRIVATE-COS-PREVIEW`
- 需要的人工验证: 生产环境注入 COS 密钥后，在实名认证审核页打开真实身份证照片确认不再 403。

## 风险与回滚关注点

- COS 密钥缺失时图片仍不能展示，但接口会返回明确错误。
- 签名接口只允许配置的私有桶/管理桶，避免成为任意 COS URL 签名器。
- 回滚时移除新增接口和前端换签调用，页面会回到直接访问私有 URL 的 403 状态。
