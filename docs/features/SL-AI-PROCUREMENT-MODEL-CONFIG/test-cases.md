# SL-AI-PROCUREMENT-MODEL-CONFIG Test Cases

## API

- 首次 GET 应返回 MiniMax 默认模型配置。
- PUT 保存配置后，再次 GET 应保留 endpoint、apiKeyEnv、topP、timeoutSec 和 retryCount。
- `apiKeyConfigured` 应由服务端环境变量计算，不依赖客户端提交值。

## UI

- 模型配置区应展示启用模型解析、供应方、模型名称、endpoint、API Key 环境变量、密钥状态、温度、Top P、Token、超时、重试和置信阈值。
- API Key 提示应说明真实密钥不在管理端落库。
- 未设置 `MINIMAX_API_KEY` 时，密钥状态显示未检测到。

## 回归

- AI 采购阶段、商品匹配、融资匹配和审核策略仍可保存。
- 管理端登录、菜单和其他配置页面不受影响。
