# Release Checklist

## 发布基本信息

- Feature ID: SL-ADMIN-NOTIFICATIONS
- 发布负责人: Codex
- 目标环境: staging / prod

## 发布前检查

- [x] `pnpm build` 已通过
- [ ] `./harness/release.sh docs/features/SL-ADMIN-NOTIFICATIONS` 已通过
- [ ] 已确认主后端站内信接口已发布
- [ ] 已完成发送站内信 smoke test

## 回滚信息

- 回滚目标版本: 合入前版本
- 回滚步骤: 回滚页面、路由、菜单和 service
- 回滚后验证: 管理端构建通过，既有菜单可访问
