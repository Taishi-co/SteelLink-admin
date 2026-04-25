# Production Auto Deploy Release Checklist

## 发布基本信息

- Feature ID: SL-ADMIN-AUTO-DEPLOY
- 发布负责人: Codex
- 目标环境: prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-AUTO-DEPLOY` 已通过
- [ ] 如存在前端工程，`build` 脚本已通过
- [ ] 如存在前端 `lint` 脚本，已通过
- [ ] 如存在前端 `typecheck` 脚本，已通过
- [ ] 如存在前端 `test:ci` 或 `test` 脚本，已通过
- [ ] 如存在后端工程，`go test ./...` 与 `go build` 已通过
- [ ] 已在 GitHub Actions Secrets 配置 `PROD_HOST`、`PROD_USER`、`PROD_SSH_KEY`、`PROD_APP_DIR`
- [ ] 如服务器部署脚本不在仓库目录下，已配置 `PROD_BACKEND_DEPLOY_COMMAND` 与 `PROD_FRONTEND_DEPLOY_COMMAND`
- [ ] 已确认远程部署用户具备读取仓库、拉取 `main`、执行部署脚本与重启服务的权限
- [ ] 已确认 `deploy_admin_backend` 与 `deploy_admin_frontend` 可重复执行
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 上一个生产稳定版本
- 回滚步骤: 禁用 `Deploy Production` workflow，按服务器既有回滚流程恢复前端产物与后端服务版本
- 回滚后验证: 检查 GitHub Actions 不再触发自动部署，手动验证管理端登录、首页与核心接口可用
