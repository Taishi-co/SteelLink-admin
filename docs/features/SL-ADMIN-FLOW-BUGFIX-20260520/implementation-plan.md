# SL-ADMIN-FLOW-BUGFIX-20260520 Implementation Plan

## Feature 信息

- Feature ID: `SL-ADMIN-FLOW-BUGFIX-20260520`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/router/router.go`
  - `backend/internal/handler/handler.go`
  - `backend/internal/handler/enterprise_service.go`
  - `frontend/src/router/index.tsx`
  - `frontend/src/layouts/AdminLayout.tsx`
  - `frontend/src/utils/permissions.ts`
  - `frontend/src/pages/*`
- 文档文件:
  - `docs/domain/role-permission-matrix.md`
  - `docs/features/SL-ADMIN-FLOW-BUGFIX-20260520/*`
- 配置文件:
  - `scripts/init-db.sql`
- 部署/脚本文件: 无

## 实施步骤

1. 将审核和订单运营写接口按角色分组，避免普通登录账号越权执行。
2. 修复订单解冻、违约处理和服务权限落库错误处理。
3. 校准待办统计与订单详情时间线。
4. 修复初始化菜单、基础表字段和历史菜单路由兼容。
5. 补齐默认角色菜单分配，避免运营/审核员没有主流程入口。
6. 前端写操作按钮按角色显隐，空菜单权限不回退全量菜单。
7. 将物流无运力状态的管理端标签、告警和时间线改为履约终止。
8. 增加资金风控拒绝识别和订单详情顶部终止告警。
9. 后端提取统一 `terminationReason`，覆盖卖方拒单、买方取消、资金驳回和物流无运力。
10. 前端订单列表和详情改用统一终止原因字段展示。
11. 补齐权限矩阵并运行验证。

## 验证计划

- `go test ./...`
- `go build ./cmd/server`
- `pnpm test:ci`
- `pnpm build`
- 人工验证: 登录不同角色后 smoke 审核接口、订单详情、菜单跳转、写操作显隐。

## 风险与回滚关注点

- 历史数据库中的旧菜单不会被删除，只会补齐/校准默认菜单；如需彻底清理需单独迁移。
- 默认角色菜单会自动补齐缺失入口；若生产环境对 role_id 2/3 做过刻意收敛，需要发布前复核。
- 解冻接口对缺少冻结事件的旧数据会返回 409，需要人工判定恢复状态。
