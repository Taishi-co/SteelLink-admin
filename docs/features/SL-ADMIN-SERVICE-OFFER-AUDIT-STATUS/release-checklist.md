# SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS Release Checklist

- [ ] `cd backend && go test ./...` 已通过
- [ ] `cd frontend && npm run build` 已通过
- [ ] `./harness/verify.sh` 已通过
- [ ] `./harness/review.sh docs/features/SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS` 已通过
- [ ] `./harness/release.sh docs/features/SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS` 已通过
- [ ] 部署后已对线上误驳回仓储资源重新审核通过
- [ ] 仓储端 `/warehouse/` 已确认该资源显示为“已通过”
