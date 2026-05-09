# SL-ADMIN-PRODUCT-CATEGORY-MASTER Release Checklist

- [ ] 管理端类目配置可编辑展示名和图片 URL。
- [ ] 管理端类目配置可上传导航图片，上传后自动回填 URL。
- [ ] `product_categories` 表缺失时 Admin 后端可创建表。
- [ ] 生产环境已配置 `COS_REGION`、`COS_SECRET_ID`、`COS_SECRET_KEY`、`COS_PUBLIC_BUCKET` 或 `COS_ADMIN_BUCKET`；如走 CDN，已配置 `COS_PUBLIC_BASE_URL`。
- [ ] 管理端 build 通过。
- [ ] Admin 后端 test/build 通过。
