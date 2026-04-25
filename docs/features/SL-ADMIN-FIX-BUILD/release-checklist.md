# Release Checklist

## 发布基本信息

- Feature ID: `SL-ADMIN-FIX-BUILD`
- 发布负责人: Codex
- 目标环境: 管理端前端构建链路

## 发布前检查

- [x] `./harness/release.sh docs/features/SL-ADMIN-FIX-BUILD` 已通过
- [x] 如存在前端工程，`build` 脚本已通过
- [x] 如存在前端 `lint` 脚本，已通过
- [x] 如存在前端 `typecheck` 脚本，已通过
- [x] 如存在前端 `test:ci` 或 `test` 脚本，已通过
- [x] 如存在后端工程，`go test ./...` 与 `go build` 已通过
- [x] 已确认环境变量、接口域名、代理、数据库或静态资源路径变更
- [x] 已确认登录、首页、核心菜单、核心列表页 smoke 范围
- [x] 已确认关键后端接口、权限与依赖配置 smoke 范围
- [x] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 本次修复前的前端构建失败版本
- 回滚步骤:
  - 回滚本次前端静态检查与打包修复提交
  - 重新执行 `./harness/verify.sh` 与前端 `pnpm build` 确认回滚结果
- 回滚后验证:
  - 确认回到修复前状态，仅影响构建链路，不影响线上业务数据
