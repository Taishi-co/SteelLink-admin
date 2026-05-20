# SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520 Implementation Plan

## Feature 信息

- Feature ID: `SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `backend/internal/handler/ai_procurement_monitor.go`
  - `backend/internal/handler/ai_procurement_monitor_test.go`
  - `backend/internal/router/router.go`
  - `backend/internal/handler/data.go`
  - `backend/internal/router/router_test.go`
  - `frontend/src/services/api.ts`
  - `frontend/src/router/index.tsx`
  - `frontend/src/layouts/AdminLayout.tsx`
  - `frontend/src/pages/ai-procurement/monitor/index.tsx`
- 文档文件:
  - `docs/domain/role-permission-matrix.md`
  - `docs/features/SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520/*`
- 配置文件:
  - `scripts/init-db.sql`
- 部署/脚本文件: 无

## 实施步骤

1. 后端新增监控 DTO、统计查询、记录列表查询和待介入判定。
2. 后端路由增加 `/ai-procurement/monitor`，权限为超管和运营。
3. 初始化菜单和默认角色菜单补齐“运营监控”。
4. 前端服务层新增监控接口类型。
5. 前端新增监控页面、路由、菜单和页面元信息。
6. 补齐角色矩阵、单元测试并运行验证门禁。

## 验证计划

- `./harness/branch-guard.sh local`
- `go test ./...`
- `go build ./cmd/server`
- `pnpm test:ci`
- `pnpm build`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520`
- `./harness/release.sh docs/features/SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520`
- 人工验证: 超管或运营登录后打开 `/ai-procurement/monitor`，确认统计、筛选和空态展示。

## 风险与回滚关注点

- 监控接口只读，不改变 AI 采购主流程数据。
- 如果生产库尚未有 `ai_procurements` 表，页面显示空态；部署业务后端表后自动有数据。
- 回滚本 PR 后新增菜单可由 AutoMigrate 下次覆盖或在菜单配置中隐藏。
