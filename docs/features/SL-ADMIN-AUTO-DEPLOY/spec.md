# Production Auto Deploy

## 基本信息

- Feature ID: SL-ADMIN-AUTO-DEPLOY
- Feature 名称: main 分支合并后自动部署管理端前后端
- 负责人: Codex
- 目标上线环境: prod

## 背景

当前项目需要在 GitHub PR 合并到 `main` 后自动触发生产部署，并在生产服务器上运行既有部署脚本，减少人工登录服务器执行部署命令的步骤。

## 范围

### In Scope

- 新增 GitHub Actions workflow，监听 `main` 分支 `push`。
- 通过 SSH 登录生产服务器。
- 在服务器仓库目录拉取最新 `main`。
- 依次执行管理端后端与前端部署脚本。
- 支持手动 `workflow_dispatch` 触发同一部署流程。

### Out of Scope

- 不修改前端页面、路由、组件或状态管理。
- 不修改后端接口、数据库、权限或业务逻辑。
- 不新增服务器部署脚本本身。
- 不在仓库中保存生产服务器地址、私钥或部署路径。

## 页面与交互影响

- 受影响页面: 无
- 受影响路由: 无
- 受影响布局或导航: 无
- 受影响组件: 无
- 受影响交互状态: 无

## 接口与数据影响

- 受影响接口: 无
- 请求参数或响应结构变化: 无
- 受影响服务层: 无
- 受影响状态管理: 无
- 受影响后端 handler/service/repository: 无
- 受影响数据库或缓存: 无
- 配置或环境变量变化: 需要在 GitHub Actions Secrets 中配置 `PROD_HOST`、`PROD_USER`、`PROD_SSH_KEY`、`PROD_APP_DIR`，可选配置 `PROD_BACKEND_DEPLOY_COMMAND`、`PROD_FRONTEND_DEPLOY_COMMAND`。
- 构建或部署影响: PR 合并到 `main` 后会触发 `.github/workflows/deploy-prod.yml`，远程执行后端部署脚本后再执行前端部署脚本。

## 权限与合规影响

- 受影响角色: 无应用内角色影响
- 新增/调整权限: GitHub Actions 需要具备通过 SSH 登录生产服务器并执行部署脚本的权限。
- 是否影响审核、风控、日志留痕: 不影响应用内审核与风控；部署日志会记录在 GitHub Actions run 与服务器部署脚本日志中。

## 验收标准

1. `.github/workflows/deploy-prod.yml` 存在，并仅在 `main` 分支 push 或手动触发时运行生产部署。
2. workflow 使用 GitHub Secrets 读取服务器 SSH 信息与仓库路径，不提交生产密钥。
3. workflow 在服务器上拉取最新 `main` 后，依次执行 `deploy_admin_backend` 与 `deploy_admin_frontend`，并允许通过 Secrets 覆盖命令。
