# AI Procurement Vision Config Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG
- 开发负责人: Codex

## 变更范围

- Admin backend:
  - `backend/internal/handler/ai_procurement_config.go`
  - `backend/internal/handler/ai_procurement_config_test.go`
- Admin frontend:
  - `frontend/src/services/api.ts`
  - `frontend/src/pages/ai-procurement/config/index.tsx`
- 文档:
  - `docs/features/SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG/*`

## 实施步骤

1. 扩展 Admin AI 采购配置 DTO，增加视觉识别字段和密钥状态。
2. 规范默认值: `minimax_token_plan`、`https://api.minimax.io/v1/coding_plan/vlm`、`MINIMAX_TOKEN_PLAN_API_KEY`。
3. 前端配置页增加商品截图识别配置块。
4. 增加后端单元测试验证旧配置补默认值和密钥状态。
5. 执行 Admin 后端测试、前端 typecheck/build 和 release gate。

## 验证计划

- `go test ./...`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`
- `./harness/release.sh docs/features/SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG`

## 风险与回滚关注点

- 真实密钥仍只存在环境变量中，Admin 只保存环境变量名。
- 业务后端需要同步支持读取 `model.vision*` 字段，否则 Admin 保存后不会影响截图识别。
- 回滚 Admin 不会删除 Redis 中已保存的字段；需要时可在页面保存旧配置覆盖。
