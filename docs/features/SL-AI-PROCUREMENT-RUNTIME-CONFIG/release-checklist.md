# SL-AI-PROCUREMENT-RUNTIME-CONFIG Release Checklist

## 发布前检查

- [x] Admin backend `go test ./...` 已通过
- [x] Admin backend build 已通过
- [x] Admin frontend build 已通过
- [x] 已确认页面不保存真实 API Key、pgvector DSN 或 Postgres 密码
- [ ] 已确认生产环境配置 `AI_PROCUREMENT_CONFIG_SYNC_TOKEN`

## 发布后验证

- [ ] Admin 页面保存“启用模型解析”后刷新仍保留
- [ ] Admin 页面保存“启用向量检索”后刷新仍保留
- [ ] SteelLink-backend 配置 runtime sync 后能读取 Admin 配置

## 回滚

- 回滚目标版本: 合入前版本
- 回滚步骤: revert 本 feature 或关闭 SteelLink-backend `AI_PROCUREMENT_RUNTIME_CONFIG_ENABLED`
