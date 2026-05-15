# AI Procurement Vision Config Test Cases

## 测试范围

- Feature ID: SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 默认配置加载 | 超管已登录 | 打开 `/ai-procurement/config` | 商品截图识别默认供应方为 `minimax_token_plan` |
| TC-02 | 保存图片识别配置 | Redis 可用 | 修改 VLM endpoint 和 Token Plan Key 环境变量名后保存 | GET 配置和 Redis runtime config 返回新字段 |
| TC-03 | 密钥状态 | Admin API 环境存在 `MINIMAX_TOKEN_PLAN_API_KEY` | 打开配置页 | 图片密钥状态显示已配置 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 密钥不存在 | 环境变量未设置 | 打开配置页 | 图片密钥状态显示未检测到 |
| TC-EX-02 | 旧配置缺少视觉字段 | 数据库中存在旧版 JSON | GET 配置 | 后端自动补默认值，不影响旧字段 |
| TC-EX-03 | Redis 不可用 | Redis 连接失败 | 保存配置 | 返回 503，避免只写 MySQL 但业务后端不生效 |

## 验证命令

| 编号 | 命令 | 期望结果 |
| --- | --- | --- |
| TC-OPS-01 | `go test ./...` | Admin backend 测试通过 |
| TC-OPS-02 | `pnpm --dir frontend typecheck` | 前端类型检查通过 |
| TC-OPS-03 | `pnpm --dir frontend build` | Admin frontend 构建通过 |
| TC-OPS-04 | `./harness/release.sh docs/features/SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG` | release gate 通过 |
