# SL-ADMIN-PRODUCT-STATUS-MAPPING Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-PRODUCT-STATUS-MAPPING
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/handler/data.go`
  - `backend/internal/handler/handler.go`
- 文档文件:
  - `docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING/spec.md`
  - `docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING/implementation-plan.md`
  - `docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING/test-cases.md`
  - `docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 调整商品状态映射函数，整数状态直接按 `0/1/2/3/4` 返回。
2. 调整字符串状态兼容映射，`approved` 对应 `2`，`rejected` 对应 `4`，`offline` 对应 `3`。
3. 调整商品列表 `status` 筛选，按真实状态值精确查询。
4. 调整首页待审核商品统计，只统计真实待审核状态。

## 验证计划

- `go test ./...`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING`
- `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING`
- 需要的人工验证: 生产部署后分别打开 `/product` 的待审核和已上架筛选。

## 风险与回滚关注点

- 该修复会改变当前错误筛选结果，运营看到的待审核数量会下降到真实值。
- 如需回滚，回滚本 PR 并重新部署 admin backend。
