# Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-NOTIFICATIONS
- 开发负责人: Codex

## 变更范围

- 代码文件: `frontend/src/pages/notification/index.tsx`、`frontend/src/services/api.ts`、`frontend/src/router/index.tsx`、`frontend/src/layouts/AdminLayout.tsx`
- 文档文件: `docs/features/SL-ADMIN-NOTIFICATIONS/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 新增 notification API 封装。
2. 新增发送页面和表单。
3. 注册路由和 fallback 菜单。
4. 执行管理端前端 build。

## 验证计划

- `pnpm build` in `frontend/`
- 需要的人工验证: 使用管理员账号提交一次站内信。

## 风险与回滚关注点

- 主后端未发布时提交会失败。
