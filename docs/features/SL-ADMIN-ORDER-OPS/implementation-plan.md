# SL-ADMIN-ORDER-OPS Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-ORDER-OPS
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/router/router.go`
  - `backend/internal/handler/data.go`
  - `backend/internal/handler/handler.go`
  - `frontend/src/services/api.ts`
  - `frontend/src/pages/order/index.tsx`
- 文档文件:
  - `docs/features/SL-ADMIN-ORDER-OPS/spec.md`
  - `docs/features/SL-ADMIN-ORDER-OPS/test-cases.md`
  - `docs/features/SL-ADMIN-ORDER-OPS/implementation-plan.md`
  - `docs/features/SL-ADMIN-ORDER-OPS/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 补齐 feature 文档与分支守卫要求。
2. 后端新增订单详情聚合查询与运营处置接口。
3. 前端订单页接入完整详情、待运营介入筛选与处置表单。
4. 执行格式化、构建与 harness 验证。

## 验证计划

- `./harness/verify.sh`
- 后端 `go test ./...`
- 前端 `npm run build`
- 需要的人工验证: 使用真实管理端账号打开订单详情并执行冻结/解冻、违约、诉讼处置 smoke。

## 风险与回滚关注点

- 运营动作直接写既有业务表，需确认管理端连接的数据库与业务端一致。
- 解冻依赖最近一次 `order_frozen` 事件中的原状态，历史冻结事件缺失时默认回到异常状态。
- 回滚时移除新增路由与前端操作入口即可，不涉及表结构回滚。
