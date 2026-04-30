# SL-AI-PROCUREMENT-MODEL-CONFIG Release Checklist

## 发布基本信息

- Feature ID: SL-AI-PROCUREMENT-MODEL-CONFIG
- 发布负责人:
- 目标环境: admin

## 发布前检查

- [ ] `./harness/review.sh docs/features/SL-AI-PROCUREMENT-MODEL-CONFIG` 已通过
- [ ] `./harness/release.sh docs/features/SL-AI-PROCUREMENT-MODEL-CONFIG` 已通过
- [ ] 已确认 `.env.example` 不包含真实密钥
- [ ] 已确认生产环境可设置 `MINIMAX_API_KEY`

## 回滚信息

- 回滚目标版本: 合入前版本
- 回滚步骤: revert 本次配置字段和页面改动
- 回滚后验证: `GET /api/v1/ai-procurement/config`
