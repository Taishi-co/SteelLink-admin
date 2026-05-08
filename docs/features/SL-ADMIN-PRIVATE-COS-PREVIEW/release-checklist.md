# SL-ADMIN-PRIVATE-COS-PREVIEW Release Checklist

## 发布基本信息

- Feature ID: SL-ADMIN-PRIVATE-COS-PREVIEW
- 发布负责人: Codex
- 目标环境: production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-PRIVATE-COS-PREVIEW` 已通过
- [ ] 前端 `pnpm build` 已通过
- [ ] 后端 `go test ./...` 与 `go build` 已通过
- [ ] 已确认生产环境配置 `COS_REGION`
- [ ] 已确认生产环境配置 `COS_SECRET_ID`
- [ ] 已确认生产环境配置 `COS_SECRET_KEY`
- [ ] 已确认生产环境配置 `COS_PRIVATE_BUCKET=steellink-user-pri-1321917249`
- [ ] 已确认登录后用户管理/实名认证审核详情 smoke 范围
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-ADMIN-PRIVATE-COS-PREVIEW --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 上一版管理端后端/前端镜像
- 回滚步骤: 回滚本次代码部署，保留 COS 私有桶配置不变
- 回滚后验证: 登录管理端，确认用户管理页面可正常打开；私有身份证图片将恢复为直接访问私有 URL 的旧行为
