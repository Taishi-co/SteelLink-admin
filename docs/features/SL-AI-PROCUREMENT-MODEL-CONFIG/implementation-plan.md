# SL-AI-PROCUREMENT-MODEL-CONFIG Implementation Plan

## 变更范围

- `backend/internal/handler/ai_procurement_config.go`
- `frontend/src/services/api.ts`
- `frontend/src/pages/ai-procurement/config/index.tsx`
- `.env.example`

## 实施步骤

1. 扩展后端模型配置 DTO 和默认值。
2. GET/PUT 时统一 normalize 配置，并用 `os.Getenv(apiKeyEnv)` 计算密钥状态。
3. 扩展前端 API 类型。
4. 扩展模型配置表单，增加 MiniMax endpoint、环境变量名和运行参数。
5. 补充 `.env.example` 和 feature 文档。
6. 执行 `./harness/review.sh` 与 `./harness/release.sh`。

## 风险与回滚

- 风险：旧配置缺少新增字段。缓解：后端 normalize 补默认值。
- 风险：运营误以为页面保存了 API Key。缓解：页面只填写环境变量名并显示密钥状态。
- 回滚：revert 本次配置字段和页面改动。
