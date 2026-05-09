# SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `frontend/src/pages/service-offer/index.tsx`
  - `frontend/src/services/api.ts`
  - `backend/internal/handler/service_offer.go`
  - `backend/internal/handler/service_offer_test.go`
- 文档文件:
  - `docs/features/SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS/spec.md`
  - `docs/features/SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS/implementation-plan.md`
  - `docs/features/SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS/test-cases.md`
  - `docs/features/SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 前端审核请求保留真实状态 `2/3`，并额外传 `auditAction` / `audit_action`。
2. 后端审核接口优先按 `auditAction` 归一化状态，无 action 时兼容真实状态值。
3. 已驳回服务资源允许重新进入审核弹窗。
4. 增加后端状态归一化单元测试。
5. 执行后端测试、前端构建和 harness 验证。

## 验证计划

- `cd backend && go test ./...`
- `cd frontend && npm run build`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS`
- `./harness/release.sh docs/features/SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS`
- 需要的人工验证: 线上部署后对误驳回仓库资源重新点“通过”，再到 `/warehouse/` 确认显示“已通过”。

## 风险与回滚关注点

- 已经被误写成 `status=3` 的线上数据不会自动恢复，需要管理端重新审核或手工 SQL 修正。
- 如果只部署前端不部署后端，旧后端仍可能把 `status=2` 当成驳回；本次修复以后端部署为必要条件。
