# SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF Release Checklist

- [ ] `cd backend && go test ./...` 已通过
- [ ] `cd frontend && npm run build` 已通过
- [ ] `./harness/verify.sh` 已通过
- [ ] `./harness/review.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF` 已通过
- [ ] `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-SEARCH-DIFF` 已通过
- [ ] 部署后确认 `/product` 搜索和筛选可用
- [ ] 部署后确认商品详情和审核弹窗展示字段核对表
