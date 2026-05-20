# SL-ADMIN-PRODUCT-AUDIT-SNAPSHOT-20260520 Implementation Plan

## Feature 信息

- Feature ID: `SL-ADMIN-PRODUCT-AUDIT-SNAPSHOT-20260520`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/handler/data.go`
  - `backend/internal/handler/handler.go`
  - `backend/internal/handler/audit_validation_test.go`
  - `frontend/src/services/api.ts`
  - `frontend/src/pages/product/index.tsx`
- 文档文件:
  - `docs/domain/role-permission-matrix.md`
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-SNAPSHOT-20260520/*`
- 配置文件:
  - `scripts/init-db.sql`
- 部署/脚本文件: 无

## 实施步骤

1. 新增并迁移 `product_audit_logs` 快照字段，兼容已有审核日志表。
2. 审核商品时在同一事务内读取审核前商品、更新审核结果、读取审核后商品并写入 before/after 快照。
3. 商品详情接口返回最近审核历史，列表接口保持轻量。
4. 前端服务层补齐审核历史类型与 snake/camel 兼容映射。
5. 商品详情和审核弹窗用最近一次审核后快照展示字段差异和审核历史。
6. 补齐快照构建单元测试、角色矩阵说明并运行门禁。

## 验证计划

- `./harness/branch-guard.sh local`
- `go test ./...`
- `go build ./cmd/server`
- `pnpm test:ci`
- `pnpm build`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-SNAPSHOT-20260520`
- `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-SNAPSHOT-20260520`
- 人工验证: 在 `/product` 打开商品详情和审核弹窗，确认审核历史与差异表展示。

## 风险与回滚关注点

- 旧商品不会自动拥有历史快照，首次上线后旧数据详情会显示暂无历史快照。
- 如果生产已有 `product_audit_logs`，AutoMigrate 只补缺失字段，不删除旧字段或旧记录。
- 回滚代码后新增审核日志字段可保留，不影响旧版审核接口读取核心字段。
