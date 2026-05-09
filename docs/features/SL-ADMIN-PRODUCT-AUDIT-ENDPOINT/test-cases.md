# SL-ADMIN-PRODUCT-AUDIT-ENDPOINT Test Cases

## 测试范围

- Feature ID: SL-ADMIN-PRODUCT-AUDIT-ENDPOINT
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 商品审核通过 | 存在待审核商品 | 在 `/product` 点击审核，选择通过上架 | 请求 `/products/:id/audit`，payload `status=1`，商品状态变为已上架 |
| TC-02 | 商品审核驳回 | 存在待审核商品 | 在 `/product` 点击审核，选择驳回 | 请求 `/products/:id/audit`，payload `status=2`，商品状态变为审核拒绝 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 商品不存在 | 对不存在商品 ID 审核 | 提交审核 | 后端返回 product not found |
| TC-EX-02 | 未登录 | 未登录管理端 | 提交审核 | 请求被鉴权拦截 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 管理员 | 审核商品 | 可进入既有 `AuditProduct` handler |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 路由对齐 | 查看 `backend/internal/router/router.go` | 商品审核路由为 `/products/:id/audit` |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 前端构建 | `cd frontend && npm run build` | 构建通过 |
| TC-OPS-02 | Release gate | `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-AUDIT-ENDPOINT` | release gate 通过 |
