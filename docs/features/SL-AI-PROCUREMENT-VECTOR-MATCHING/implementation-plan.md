# SL-AI-PROCUREMENT-VECTOR-MATCHING Implementation Plan

## 当前变更

- 扩展管理端 AI 采购匹配配置结构。
- 前端配置页增加 pgvector/embedding/hybrid 参数。
- 补充 feature 文档。

## 后续实施步骤

1. 后端提供向量库状态接口，返回表记录数量、最近同步时间和 embedding 模型。
2. 管理端增加向量库状态卡片。
3. 管理端增加手动重建索引入口。
4. 审计记录索引重建和配置变更。

## 回滚

- 关闭向量检索开关。
- revert 本次管理端字段扩展。
