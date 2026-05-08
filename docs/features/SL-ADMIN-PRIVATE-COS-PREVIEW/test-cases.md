# SL-ADMIN-PRIVATE-COS-PREVIEW Test Cases

## 测试范围

- Feature ID: SL-ADMIN-PRIVATE-COS-PREVIEW
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 身份证照片换签展示 | 管理员已登录，用户存在 `steellink-user-pri-1321917249` 身份证 URL，后端已配置 COS 密钥 | 打开用户详情/审核抽屉 | 页面调用 `/files/private-url`，图片使用带签名参数的 URL 展示 |
| TC-02 | 签名有效期 | 已配置 `COS_PRIVATE_URL_EXPIRES_SECONDS=600` | 请求私有 URL 签名接口 | 返回 `expiresIn=600`，URL 包含 COS 签名查询参数 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 未配置 COS 密钥 | `COS_SECRET_ID` 或 `COS_SECRET_KEY` 为空 | 请求私有 URL 签名接口 | 返回 503 和明确配置错误 |
| TC-EX-02 | 非允许桶 | 传入其他 COS bucket URL | 请求私有 URL 签名接口 | 返回 403，不生成签名 |
| TC-EX-03 | 非 HTTPS URL | 传入 `http://` 或非 COS URL | 请求私有 URL 签名接口 | 返回 400，不生成签名 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 请求 `/api/v1/files/private-url` | 被现有 Auth 中间件拦截 |
| TC-AUTH-02 | 已登录管理端用户 | 请求已配置私有桶对象 | 返回短期签名 URL |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 后端测试 | `cd backend && go test ./...` | 后端测试通过 |
| TC-BE-02 | 后端构建 | `cd backend && go build -o /tmp/steellink-admin-api ./cmd/server` | 后端构建通过 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 管理端构建 | `cd frontend && pnpm build` | 前端构建通过 |
| TC-OPS-02 | Review gate | `./harness/review.sh docs/features/SL-ADMIN-PRIVATE-COS-PREVIEW` | feature 文档完整且验证通过 |
