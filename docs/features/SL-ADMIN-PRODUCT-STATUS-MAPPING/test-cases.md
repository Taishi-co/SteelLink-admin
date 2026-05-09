# SL-ADMIN-PRODUCT-STATUS-MAPPING Test Cases

## 测试范围

- Feature ID: SL-ADMIN-PRODUCT-STATUS-MAPPING
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 筛选待审核商品 | 存在 `status=1` 商品 | 请求 `/api/v1/products?status=1` | 只返回待审核商品，响应 `status=1` |
| TC-02 | 筛选已上架商品 | 存在 `status=2` 商品 | 请求 `/api/v1/products?status=2` | 返回已上架商品，响应 `status=2` |
| TC-03 | 查看商品详情 | 存在 `status=2` 商品 | 请求 `/api/v1/products/:id` | 详情响应 `status=2`，页面展示已上架 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 非法状态筛选 | 无 | 请求 `/api/v1/products?status=99` | 不附加错误状态条件，接口不报错 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 请求商品管理接口 | 被鉴权中间件拒绝 |
| TC-AUTH-02 | 已登录审核员/管理员 | 筛选商品状态 | 可按真实状态查看结果 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 后端测试 | `go test ./...` | 测试通过 |
| TC-BE-02 | 仓库验证 | `./harness/verify.sh` | 后端测试/构建通过 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 发布检查 | `./harness/release.sh docs/features/SL-ADMIN-PRODUCT-STATUS-MAPPING` | 发布门禁通过 |
