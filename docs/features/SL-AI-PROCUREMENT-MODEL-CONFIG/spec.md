# SL-AI-PROCUREMENT-MODEL-CONFIG Spec

## 背景

管理端 AI 采购配置页已有模型供应方、模型名称、温度、Token 和置信阈值，但缺少 MiniMax/OpenAI-compatible provider 需要的 endpoint、API Key 引用、Top P、超时、重试和密钥状态。直接在管理端保存 API Key 会增加明文落库风险。

## 目标

- 扩展管理端 AI 采购模型配置字段。
- 支持 MiniMax 推荐配置。
- API Key 不在管理端落库，只保存环境变量名并展示服务端是否检测到密钥。
- 保持现有配置读写接口兼容。

## 范围

- 扩展 `GET/PUT /api/v1/ai-procurement/config` 的 `model` 配置结构。
- 扩展前端 `/ai-procurement/config` 模型配置表单。
- `.env.example` 补充 `MINIMAX_API_KEY` 注入说明。

## 非范围

- 不实现密钥加密落库。
- 不直接调用 MiniMax 验证密钥有效性。
- 不修改买家侧 AI 采购页面。

## 推荐 MiniMax 配置

- 启用模型解析: 开启
- 模型供应方: `minimax`
- 模型名称: `MiniMax-M2.7`
- Endpoint: `https://api.minimaxi.com/anthropic`
- API Key 环境变量: `MINIMAX_API_KEY`
- 温度: `0.2`
- Top P: `0.9`
- 最大 Token: `1200`
- 超时秒数: `20`
- 失败重试: `1`
- 置信阈值: `0.78`
