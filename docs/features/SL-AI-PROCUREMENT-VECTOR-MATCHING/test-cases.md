# SL-AI-PROCUREMENT-VECTOR-MATCHING Test Cases

## UI

- 页面可编辑匹配模式和是否启用向量检索。
- 页面可编辑 Milvus 地址、商品集合、供应商集合、融资产品集合。
- 页面可编辑 embedding 模型、维度、TopK、最低相似度和 hybrid 权重。
- 保存后再次 GET 应返回新增字段。

## 安全

- 页面不提供 Milvus 密码输入。
- `.env.example` 只说明 `MILVUS_PASSWORD`，不包含真实密码。

## 回归

- 原商品匹配策略字段仍可保存。
- 模型配置、融资配置和审核配置不受影响。
