# SL-ADMIN-FLOW-BUGFIX-20260520 Release Checklist

## 发布基本信息

- Feature ID: `SL-ADMIN-FLOW-BUGFIX-20260520`
- 发布负责人: Codex
- 目标环境: staging / prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-FLOW-BUGFIX-20260520` 已通过
- [ ] 前端 `pnpm build` 已通过
- [ ] 后端 `go test ./...` 与 `go build ./cmd/server` 已通过
- [ ] 已确认超管、运营、审核员、普通账号的接口权限 smoke 范围
- [ ] 已确认订单冻结、解冻、违约处理 smoke 范围
- [ ] 已确认菜单和历史路径跳转 smoke 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-ADMIN-FLOW-BUGFIX-20260520 --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 上一版 admin 后端和前端镜像。
- 回滚步骤: 回滚本次代码部署。
- 回滚后验证: 登录、菜单、商品审核、订单详情和 AI 采购配置页面可访问。
