# AI Procurement Vision Config Spec

## 基本信息

- Feature ID: SL-ADMIN-AI-PROCUREMENT-VISION-CONFIG
- Feature 名称: AI 采购配置支持商品截图识别
- 负责人: Codex
- 目标上线环境: Admin 前端与 Admin API

## 背景

卖家商品 AI 批量导入的截图识别使用 MiniMax Token Plan VLM。运营希望像 AI 采购模型配置一样，在 Admin 页面配置图片识别供应方、VLM endpoint 和 Token Plan Key 的环境变量名，而不是改业务后端配置文件后重启。

## 范围

### In Scope

- 扩展 `GET/PUT /api/v1/ai-procurement/config` 的 `model` 配置结构。
- Admin 页面 `/ai-procurement/config` 增加商品截图识别配置项。
- 保存后继续写入 `admin_system_configs` 和 Redis key `ai_procurement_config`。
- 页面只保存环境变量名，不保存真实 MiniMax Token Plan Key。

### Out of Scope

- 不在 Admin 数据库保存真实密钥。
- 不新增单独的商品导入配置页。
- 不改变现有 AI 采购聊天模型、向量召回和融资审核配置语义。

## 接口与数据影响

- 受影响接口:
  - `GET /api/v1/ai-procurement/config`
  - `PUT /api/v1/ai-procurement/config`
  - `GET /api/v1/internal/ai-procurement/runtime-config`
- 新增 JSON 字段:
  - `model.visionProvider`
  - `model.visionBaseUrl`
  - `model.visionApiKeyEnv`
  - `model.visionApiKeyConfigured`
- 持久化位置:
  - MySQL `admin_system_configs.config_value`
  - Redis `ai_procurement_config`

## 权限与合规影响

- 仅超管可访问现有 AI 采购配置页。
- 密钥状态只根据 Admin API 进程环境变量是否存在返回布尔值，不返回真实值。

## 验收标准

1. 进入 `/ai-procurement/config` 可以看到商品截图识别配置。
2. 保存配置后刷新页面，图片识别供应方、endpoint、环境变量名不丢失。
3. Redis runtime config 中包含新增字段，业务后端可读取。
4. 环境变量存在时，页面显示图片密钥已配置。
