# SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/handler/data.go`
  - `backend/internal/handler/handler.go`
  - `frontend/src/services/api.ts`
  - `frontend/src/pages/product/index.tsx`
- 文档文件:
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF/spec.md`
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF/implementation-plan.md`
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF/test-cases.md`
  - `docs/features/SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 后端 `ListProducts` 支持关键字、状态、企业、品类、货品类型、仓库资源筛选。
2. 后端商品状态返回改为真实状态值，不再压缩成旧审核态。
3. 前端商品列表增加筛选表单和重置操作。
4. 前端商品详情抽屉和审核弹窗增加关键字段核对表。
5. 执行构建和 harness 验证。

## 验证计划

- `gofmt -w backend/internal/handler/data.go backend/internal/handler/handler.go`
- `cd backend && go test ./...`
- `cd frontend && npm run build`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF`
- `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF`
- 需要的人工验证: 部署后在 `/product` 使用筛选组合查找商品，并打开详情/审核弹窗核对字段表。

## 风险与回滚关注点

- 当前系统没有完整历史字段快照，所以前后对比中的“审核基准”会明确显示“暂无历史快照”。后续如要真实 before/after，需要卖家端或后端保存商品提交快照。
