# SL-ADMIN-FLOW-AUDIT-20260520 Release Checklist

## 发布基本信息

- Feature ID: `SL-ADMIN-FLOW-AUDIT-20260520`
- 发布负责人: Codex
- 目标环境: staging / prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-FLOW-AUDIT-20260520` 已通过
- [ ] 前端 `pnpm lint`、`pnpm typecheck`、`pnpm test:ci`、`pnpm build` 已通过
- [ ] 后端 `go test ./...` 与 `go build ./cmd/server` 已通过
- [ ] 已确认企业待审默认通过、已驳回可重新审核、已通过无重复审核入口
- [ ] 已确认实名驳回后详情状态显示正确
- [ ] 已确认各驳回/暂停路径无原因时被阻断
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-ADMIN-FLOW-AUDIT-20260520 --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 上一版 admin 前后端镜像。
- 回滚步骤: 回滚本次 PR。
- 回滚后验证: 企业、商品、实名、服务资源审核页面可打开并提交。
