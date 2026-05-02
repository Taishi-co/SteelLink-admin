# SL-AI-PROCUREMENT-RUNTIME-CONFIG Spec

## 背景

运营需要在 Admin 平台直接调整 AI 采购开关，例如是否启用模型解析、是否启用 Milvus 向量召回、Milvus 集合名、embedding endpoint 和 hybrid 权重。仅依赖业务后端 `.env` 会导致每次调整都需要改服务器配置并重启。

## 目标

- 在 Admin AI采购配置页补齐运行时开关和连接元数据字段。
- Admin 后端保存模型解析、向量召回和 embedding 配置，但不保存真实密钥。
- Admin 后端提供内部只读运行时配置接口，供 SteelLink-backend 拉取。

## 范围

- 扩展 `GET/PUT /api/v1/ai-procurement/config` 的 `matching` 字段。
- 新增 `GET /api/v1/internal/ai-procurement/runtime-config`。
- 通过 `AI_PROCUREMENT_CONFIG_SYNC_TOKEN` 支持内部接口鉴权。

## 非范围

- 不保存 MiniMax API Key 或 Milvus 密码明文。
- 不实现 Milvus 健康检查、索引重建或商品向量同步 job。

## 安全规则

- Admin 页面只保存 `MINIMAX_API_KEY`、`MILVUS_PASSWORD` 这类环境变量名。
- 真实密钥仍由服务器环境变量提供。
- 生产环境应配置 `AI_PROCUREMENT_CONFIG_SYNC_TOKEN`，业务后端拉取配置时带 `X-Internal-Token`。
