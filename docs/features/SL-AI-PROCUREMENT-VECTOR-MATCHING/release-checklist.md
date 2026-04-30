# SL-AI-PROCUREMENT-VECTOR-MATCHING Release Checklist

## 发布基本信息

- Feature ID: SL-AI-PROCUREMENT-VECTOR-MATCHING
- 发布负责人:
- 目标环境: admin

## 发布前检查

- [ ] `./harness/review.sh docs/features/SL-AI-PROCUREMENT-MODEL-CONFIG` 已通过
- [ ] `./harness/release.sh docs/features/SL-AI-PROCUREMENT-MODEL-CONFIG` 已通过
- [ ] 已确认不在管理端保存 Milvus 密码
- [ ] 已确认默认关闭向量召回

## 回滚信息

- 回滚目标版本: 合入前版本
- 回滚步骤: 关闭向量检索或 revert 本次 UI 字段扩展
- 回滚后验证: `GET /api/v1/ai-procurement/config`
