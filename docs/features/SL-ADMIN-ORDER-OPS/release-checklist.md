# SL-ADMIN-ORDER-OPS Release Checklist

## 发布基本信息

- Feature ID: SL-ADMIN-ORDER-OPS
- 发布负责人: Codex
- 目标环境: staging

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-ORDER-OPS` 已通过
- [ ] 如存在前端工程，`build` 脚本已通过
- [ ] 如存在前端 `lint` 脚本，已通过
- [ ] 如存在前端 `typecheck` 脚本，已通过
- [ ] 如存在前端 `test:ci` 或 `test` 脚本，已通过
- [ ] 如存在后端工程，`go test ./...` 与 `go build` 已通过
- [ ] 已确认环境变量、接口域名、代理、数据库或静态资源路径变更
- [ ] 已确认登录、首页、核心菜单、核心列表页 smoke 范围
- [ ] 已确认关键后端接口、权限与依赖配置 smoke 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-ADMIN-ORDER-OPS --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 上一稳定版本
- 回滚步骤: 回滚本 feature 对 admin 前后端的代码变更并重新发布。
- 回滚后验证: 登录管理端，确认订单列表可访问，详情不再显示新增处置入口。
