# SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520 Release Checklist

## 发布基本信息

- Feature ID: `SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520`
- 发布负责人: Codex
- 目标环境: staging / prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520` 已通过
- [ ] 前端 `build` 已通过
- [ ] 前端 `lint` 已通过或仅存在非阻断 warning
- [ ] 前端 `typecheck` 已通过
- [ ] 前端 `test:ci` 已通过
- [ ] 后端 `go test ./...` 与 `go build ./cmd/server` 已通过
- [ ] 已确认无环境变量、接口域名、代理或静态资源路径变更
- [ ] 已确认 `ai_procurements` 表不存在时监控页空态可用
- [ ] 已确认超管和运营可访问监控页，审核员和普通账号不可访问
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520 --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 上一个 admin 后端和前端版本。
- 回滚步骤: 回滚本 PR 的前后端发布版本；无数据库结构需要回滚。
- 回滚后验证: 登录管理端，确认 `/ai-procurement/config` 和既有核心菜单仍可访问。
