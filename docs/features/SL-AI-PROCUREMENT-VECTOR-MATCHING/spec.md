# SL-AI-PROCUREMENT-VECTOR-MATCHING Spec

## 背景

AI 采购管理端需要能配置商品、供应商、融资产品的向量召回策略。当前商品匹配配置只有品类权重、供应商优先级、价格容忍和推荐数量，无法表达 Milvus 集合、embedding 模型、TopK 和 hybrid 权重。

## 目标

- 在 AI 采购配置页补充向量检索配置。
- 支持商品、供应商、融资产品三个集合名。
- 支持 Milvus endpoint、embedding 模型、向量维度、TopK、最低相似度、关键词/向量权重。
- 默认关闭向量召回，保留规则匹配。

## 非范围

- 不在管理端保存 Milvus 密码。
- 不在本次实现向量库健康检查。
- 不在本次实现索引重建按钮。

## 推荐配置

- 匹配模式: `hybrid`
- 启用向量检索: 灰度开启
- 向量数据库: `milvus`
- 向量库地址: `localhost:19530` 或生产 Milvus 地址
- 商品集合: `ai_product_embeddings`
- 供应商集合: `ai_supplier_embeddings`
- 融资产品集合: `ai_finance_product_embeddings`
- Embedding 模型: `embo-01`
- 向量维度: `1536`
- 向量 TopK: `10`
- 最低相似度: `0.65`
- 关键词权重: `0.35`
- 向量权重: `0.65`
