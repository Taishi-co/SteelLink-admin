# Release Checklist

## 发布基本信息

- Feature ID: `SL-ADMIN-WORKBENCH-FLOW-BASELINE`
- 发布负责人: Codex
- 目标环境: 本地开发 / 测试联调环境

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-WORKBENCH-FLOW-BASELINE` 已通过
- [ ] `bash -n harness/deploy.sh` 已通过
- [ ] `./harness/deploy.sh staging` dry-run 已确认
- [ ] `ADMIN_FRONTEND_API_BASE_URL=... ./harness/deploy.sh staging` dry-run 已确认
- [ ] `./harness/verify.sh` 已通过或已记录既有阻断
- [ ] 已确认本轮无业务接口、数据库、权限变更
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 发布前版本
- 回滚步骤: 回退部署 harness、项目地图和 feature 文档
- 回滚后验证: 执行 `bash -n harness/deploy.sh` 与 `./harness/deploy.sh staging`
