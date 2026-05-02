# SL-AI-PROCUREMENT-RUNTIME-CONFIG Test Cases

## Admin 页面

- AI采购配置页可切换“启用模型解析”。
- AI采购配置页可切换“启用向量检索”。
- 页面可编辑 Milvus endpoint、database、用户名、密码环境变量名和三个集合名。
- 页面可编辑 embedding provider、endpoint、API Key 环境变量名、模型、维度、TopK 和 hybrid 权重。
- 页面不提供真实 API Key 或 Milvus 密码输入框。

## Admin 后端

- `GET /api/v1/ai-procurement/config` 返回新增字段默认值。
- `PUT /api/v1/ai-procurement/config` 能保存新增字段。
- `GET /api/v1/internal/ai-procurement/runtime-config` 返回同一份运行时配置。
- 设置 `AI_PROCUREMENT_CONFIG_SYNC_TOKEN` 后，未带 `X-Internal-Token` 或 token 错误时返回 401。

## 回归

- 已有 AI采购配置保存、操作日志和菜单权限不受影响。
