# SL-ADMIN-PRODUCT-STATUS-MAPPING Release Checklist

## 发布基本信息

- Feature ID: SL-ADMIN-PRODUCT-STATUS-MAPPING
- 发布负责人: Codex
- 目标环境: Admin production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING` 已通过
- [ ] 后端 `go test ./...` 已通过
- [ ] 已确认无新增环境变量、接口域名、代理、数据库或静态资源路径变更
- [ ] 已确认 `/product` 待审核、已上架、审核拒绝筛选 smoke 范围
- [ ] 已确认 `GET /api/v1/products?status=1` 与 `status=2` 返回状态语义正确
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 上一个稳定 admin backend 版本
- 回滚步骤: 回滚本 PR 对应提交并重新部署 admin backend
- 回滚后验证: `/product` 商品列表接口恢复到回滚版本行为
