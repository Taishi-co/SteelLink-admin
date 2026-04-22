# Test Cases

## 测试范围

- Feature ID: `SL-ADMIN-WIN-DEV-BOOTSTRAP`
- 测试负责人: 二号开发者

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 默认管理员登录成功 | 后端已启动 | 调用 `POST /api/v1/login`，用户名密码为 `admin / admin123` | 返回 `code=0`，包含 `token` 和 `adminInfo` |
| TC-02 | 读取当前管理员信息 | 已获取登录 token | 调用 `GET /api/v1/admin/info` 并带 `Authorization` | 返回 `code=0`，管理员用户名为 `admin` |
| TC-03 | Windows 启动脚本可执行 | Windows 本机具备 PowerShell | 执行 `scripts/windows/start-backend.ps1` 与 `start-frontend.ps1` | 服务可正常启动 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 错误密码登录失败 | 后端已启动 | 使用错误密码调用登录接口 | 返回 `401` |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录请求 | 请求 `/api/v1/admin/info` | 返回未授权 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 管理端后端可构建 | `go build ./...` | 构建通过 |
| TC-BE-02 | 管理端前端可构建 | `pnpm build` | 构建通过 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | Windows 附加改动隔离 | 检查新增脚本目录 | 仅在 `scripts/windows/` 下新增入口，不影响默认命令 |
