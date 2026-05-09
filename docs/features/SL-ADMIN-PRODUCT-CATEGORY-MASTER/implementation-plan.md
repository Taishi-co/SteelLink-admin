# SL-ADMIN-PRODUCT-CATEGORY-MASTER Implementation Plan

1. 扩展 Admin 后端类目 DTO 和 SQL，支持 `display_name`、`image_url`。
2. 补齐 `product_categories` 表自动建表逻辑。
3. 增加管理端图片上传接口，将类目导航图片上传到公开 COS bucket 并返回 URL。
4. 扩展 Admin 前端类型、列表列、表单字段和上传控件。
5. 运行 Admin 前端 build 和后端 test/build。
