# AI Procurement Vision Config Release Checklist

## 发布基本信息

- Feature ID: SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG
- 发布负责人: Codex
- 目标环境: Admin API 与 Admin 前端

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG` 已通过
- [ ] 已确认 Admin API 环境变量中存在真实 Token Plan Key
- [ ] 已确认业务后端环境变量中存在同名真实 Token Plan Key
- [ ] 已确认业务后端已合入读取 `model.vision*` runtime config 的变更
- [ ] 已在 Admin 页面保存一次配置，并确认 Redis `ai_procurement_config` 已更新
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG --pr-url <github-pr-url> --pr-title "<pr-title>"`

## 回滚信息

- 回滚目标版本: 上一版 Admin API 与 Admin 前端镜像
- 回滚步骤: 回滚 Admin 代码；如需回滚配置，在 Admin 页面保存旧字段或删除 Redis runtime config。
- 回滚后验证: `/ai-procurement/config` 可正常打开，AI 采购配置保存仍可用。
