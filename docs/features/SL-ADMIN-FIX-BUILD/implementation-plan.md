# Implementation Plan

## Feature 信息

- Feature ID: `SL-ADMIN-FIX-BUILD`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `frontend/src/pages/order/index.tsx`
  - `frontend/src/pages/user/index.tsx`
  - `frontend/src/pages/system/menu/index.tsx`
  - `frontend/src/pages/system/role/index.tsx`
  - `frontend/src/pages/enterprise/index.tsx`
  - `frontend/src/router/index.tsx`
  - `frontend/src/App.tsx`
- 文档文件:
  - `docs/features/SL-ADMIN-FIX-BUILD/spec.md`
  - `docs/features/SL-ADMIN-FIX-BUILD/test-cases.md`
  - `docs/features/SL-ADMIN-FIX-BUILD/implementation-plan.md`
  - `docs/features/SL-ADMIN-FIX-BUILD/release-checklist.md`
- 配置文件:
  - 无
- 部署/脚本文件:
  - 无

## 实施步骤

1. 复现前端 `pnpm build` 与 `./harness/verify.sh` 的阻断项，确认修复边界。
2. 以最小改动修复页面中的未使用导入、宽泛树节点类型与路由导出告警。
3. 执行 `verify`、`review`、`release` 相关命令，记录剩余风险或环境阻断。

## 验证计划

- `./harness/verify.sh`
- `pnpm build`
- `./harness/review.sh docs/features/SL-ADMIN-FIX-BUILD`
- `./harness/release.sh docs/features/SL-ADMIN-FIX-BUILD`
- 需要的人工验证:
  - 打开主要菜单页面，确认基础渲染与交互未变化

## 风险与回滚关注点

- 本次修改覆盖多个前端页面入口，需避免误改现有表格和弹窗交互。
- 若出现回归，可直接回滚本次前端静态检查修复提交，不涉及数据迁移或接口回滚。
