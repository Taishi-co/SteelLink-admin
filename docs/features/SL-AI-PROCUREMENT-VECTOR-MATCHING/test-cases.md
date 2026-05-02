# SL-AI-PROCUREMENT-VECTOR-MATCHING Test Cases

## UI

- 页面可编辑匹配模式和是否启用向量检索。
- 页面可编辑 pgvector DSN 环境变量名、商品向量表、供应商向量表、融资产品向量表。
- 页面可编辑 embedding 模型、维度、TopK、最低相似度和 hybrid 权重。
- 保存后再次 GET 应返回新增字段。

## 安全

- 页面不提供 pgvector DSN 或 Postgres 密码输入。
- `.env.example` 只说明 `PGVECTOR_DSN` / `PGVECTOR_PASSWORD`，不包含真实连接串或密码。

## 回归

- 原商品匹配策略字段仍可保存。
- 模型配置、融资配置和审核配置不受影响。
