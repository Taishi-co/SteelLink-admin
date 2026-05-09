# SL-ADMIN-PRODUCT-AUDIT-ENDPOINT Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-PRODUCT-AUDIT-ENDPOINT
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `frontend/src/services/api.ts`
  - `frontend/src/pages/product/index.tsx`
- 文档文件:
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT/spec.md`
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT/implementation-plan.md`
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT/test-cases.md`
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 修改 `productApi.audit`，请求后端已注册的 `/products/:id/audit`。
2. 修改审核弹窗动作值，前端提交 `1=通过、2=驳回`。
3. 保持商品状态展示枚举 `0/1/2/3/4` 不变。
4. 执行前端构建和仓库 gate。

## 验证计划

- `cd frontend && npm run build`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT`
- `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT`
- 需要的人工验证: 部署后在 `/product` 审核商品 4，通过和驳回各验证一次请求路径和状态结果。

## 风险与回滚关注点

- 本修复只改管理端前端请求。如果生产后端不是当前 `SteelLink-admin/backend` 版本，需要确认 `/api/v1/products/:id/audit` 已部署。
