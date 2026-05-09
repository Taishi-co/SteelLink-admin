# SL-ADMIN-CATEGORY-PRODUCT-COUNT Release Checklist

## 发布基本信息

- Feature ID: SL-ADMIN-CATEGORY-PRODUCT-COUNT
- 发布负责人: Codex
- 目标环境: Admin production

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-ADMIN-CATEGORY-PRODUCT-COUNT` 已通过
- [ ] 前端 `npm run build` 已通过
- [ ] 后端 `go test ./...` 已通过
- [ ] 已确认无新增环境变量、接口域名、代理、数据库或静态资源路径变更
- [ ] 已确认 `/product/category` 类目列表 smoke 范围
- [ ] 已确认 `GET /api/v1/product/categories` 返回 `productCount`
- [ ] 已确认 `DELETE /api/v1/product/categories/:id` 对有关联商品类目返回 `409`
- [ ] GitHub PR 创建后，已执行 `./harness/lark-pr-doc.sh docs/features/SL-ADMIN-CATEGORY-PRODUCT-COUNT --pr-url <github-pr-url> --pr-title "<pr-title>"` 并写入共享飞书 PR 记录
- [ ] 已确认回滚方式

## 回滚信息

- 回滚目标版本: 上一个稳定 admin backend/frontend 版本
- 回滚步骤: 回滚本 PR 对应提交并重新部署 admin backend 与 admin frontend
- 回滚后验证: `/product/category` 类目列表、编辑、删除流程恢复到回滚版本行为
