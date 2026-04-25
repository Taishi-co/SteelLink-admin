# Production Auto Deploy Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-AUTO-DEPLOY
- 开发负责人: Codex

## 变更范围

- 代码文件: 无应用代码变更
- 文档文件: `docs/features/SL-ADMIN-AUTO-DEPLOY/spec.md`、`docs/features/SL-ADMIN-AUTO-DEPLOY/test-cases.md`、`docs/features/SL-ADMIN-AUTO-DEPLOY/implementation-plan.md`、`docs/features/SL-ADMIN-AUTO-DEPLOY/release-checklist.md`
- 配置文件: `.github/workflows/deploy-prod.yml`
- 部署/脚本文件: 新增 GitHub Actions 生产部署入口；不修改服务器部署脚本

## 实施步骤

1. 新建 feature 文档，记录自动部署的触发条件、Secrets、影响范围与验收标准。
2. 新增 GitHub Actions workflow，监听 `main` 分支 push 与手动触发。
3. 在 workflow 中通过 SSH 登录服务器，支持私钥或密码认证。
4. 在服务器上先执行 `./deploy_admin_backend.sh`，再执行 `./deploy_admin_frontend.sh`。
5. 执行 `branch-guard`、`verify`、`review` 与 `release` 检查。

## 验证计划

- `./harness/branch-guard.sh local`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-AUTO-DEPLOY`
- `./harness/release.sh docs/features/SL-ADMIN-AUTO-DEPLOY`
- 需要的人工验证: 在 GitHub repo 配置生产 Secrets 后，通过一次 `workflow_dispatch` 或合并测试 PR 验证 SSH 连接与两个服务器部署脚本执行结果。

## 风险与回滚关注点

- 风险: SSH Secret、服务器路径或部署命令配置错误会导致 workflow 失败。
- 风险: 远程部署脚本如果不是幂等的，重复触发可能产生不一致状态。
- 回滚: 禁用或回滚 `.github/workflows/deploy-prod.yml`，恢复人工登录服务器执行部署脚本。
