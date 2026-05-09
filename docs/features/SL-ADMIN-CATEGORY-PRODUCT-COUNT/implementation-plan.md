# SL-ADMIN-CATEGORY-PRODUCT-COUNT Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-CATEGORY-PRODUCT-COUNT
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/handler/data.go`
  - `backend/internal/handler/handler.go`
  - `frontend/src/services/api.ts`
  - `frontend/src/pages/product/category/index.tsx`
- 文档文件:
  - `docs/features/SL-ADMIN-CATEGORY-PRODUCT-COUNT/spec.md`
  - `docs/features/SL-ADMIN-CATEGORY-PRODUCT-COUNT/implementation-plan.md`
  - `docs/features/SL-ADMIN-CATEGORY-PRODUCT-COUNT/test-cases.md`
  - `docs/features/SL-ADMIN-CATEGORY-PRODUCT-COUNT/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 扩展类目响应结构，新增 `productCount` 字段。
2. 在类目列表查询后统计 `products` 表中与当前类目匹配的商品数量，兼容 `category_id`、`category`、`sub_category` 字段。
3. 删除类目前复用同一统计逻辑，数量大于 0 时返回 `409`。
4. 前端类目表格展示商品数量，并对有关联商品的类目禁用删除。
5. 补齐 feature 文档并运行 harness 验证。

## 验证计划

- `go test ./...`
- `npm run build`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-CATEGORY-PRODUCT-COUNT`
- `./harness/release.sh docs/features/SL-ADMIN-CATEGORY-PRODUCT-COUNT`
- 需要的人工验证: 在 `/product/category` 检查商品数量列；尝试删除有关联商品和无关联商品的类目。

## 风险与回滚关注点

- 商品历史数据可能同时使用 `category_id` 或文本类目字段，统计逻辑需要保持兼容。
- 回滚时需同时回滚前端展示和后端删除保护，避免接口响应字段与页面预期不一致。
