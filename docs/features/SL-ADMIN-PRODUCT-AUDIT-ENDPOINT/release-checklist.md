# SL-ADMIN-PRODUCT-AUDIT-ENDPOINT Release Checklist

- [ ] `cd frontend && npm run build` 已通过
- [ ] `./harness/verify.sh` 已通过
- [ ] `./harness/review.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT` 已通过
- [ ] `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT` 已通过
- [ ] 部署后确认 `/product` 商品审核不再请求 `/products/:id/status`
- [ ] 部署后确认审核通过落为已上架，驳回落为审核拒绝
