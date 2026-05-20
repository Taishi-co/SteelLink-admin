# SL-ADMIN-FLOW-AUDIT-20260520 Implementation Plan

## Feature 信息

- Feature ID: `SL-ADMIN-FLOW-AUDIT-20260520`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/handler/handler.go`
  - `backend/internal/handler/enterprise_service.go`
  - `backend/internal/handler/service_offer.go`
  - `backend/internal/handler/audit_validation_test.go`
  - `frontend/src/pages/user/index.tsx`
  - `frontend/src/pages/enterprise/index.tsx`
  - `frontend/src/pages/enterprise/service-applications.tsx`
  - `frontend/src/pages/service-offer/index.tsx`
  - `frontend/src/pages/product/index.tsx`
- 文档文件:
  - `docs/features/SL-ADMIN-FLOW-AUDIT-20260520/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 修复前端企业审核默认值和重复审核入口。
2. 修复用户实名驳回后本地详情状态更新。
3. 前端审核/控制表单增加驳回或暂停原因校验。
4. 后端审核接口增加明确状态和原因校验。
5. 补充后端状态校验单元测试并运行验证。

## 验证计划

- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-FLOW-AUDIT-20260520`
- `./harness/release.sh docs/features/SL-ADMIN-FLOW-AUDIT-20260520`
- 需要的人工验证: 超管/运营/审核员分别 smoke 企业、商品、实名、服务资源驳回路径。

## 风险与回滚关注点

- 上线后驳回/暂停必须填写原因，可能改变部分运营习惯；回滚代码即可恢复旧行为。
- 历史无原因驳回数据不会被补齐。
