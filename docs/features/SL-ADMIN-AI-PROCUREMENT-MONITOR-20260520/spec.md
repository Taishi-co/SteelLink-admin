# SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520 Spec

## 基本信息

- Feature ID: `SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520`
- Feature 名称: AI 采购运营监控页
- 负责人: Codex
- 目标上线环境: staging / prod

## 背景

AI 采购已经覆盖指令解析、商品匹配、订单生成、融资匹配和审核提交，但管理端目前只有配置页。运营无法在 Admin 侧判断近期 AI 采购是否正常流转、失败或停滞在哪个阶段，也看不到当前运行配置是否启用模型解析、向量召回和融资匹配。

## 范围

### In Scope

- 新增管理端 `/ai-procurement/monitor` 页面。
- 新增 `GET /api/v1/ai-procurement/monitor` 只读监控接口。
- 展示总量、今日新增、进行中、审核提交、异常和待介入统计。
- 展示阶段分布、状态分布和当前运行配置摘要。
- 展示最近 AI 采购记录，包含采购单号、企业、创建人、商品、阶段、状态、匹配数量、订单数量、融资方案数量和待介入原因。
- 初始化菜单和默认运营角色菜单补齐监控页入口。

### Out of Scope

- 不新增 AI 采购单审核、强制推进、重试或关闭动作。
- 不改业务后端 `ai_procurements` 状态机。
- 不改变现有 AI 采购配置保存逻辑。
- 不移除 `AI_PROCUREMENT_CONFIG_SYNC_TOKEN` 强制校验。

## 页面与交互影响

- 受影响页面: AI 采购管理。
- 新增路由: `/ai-procurement/monitor`。
- 受影响布局或导航: AI 采购管理下新增“运营监控”菜单。
- 受影响组件: 新增监控看板页面。
- 受影响交互状态: 数据表不存在时展示空监控状态，不阻断配置页。

## 接口与数据影响

- 受影响接口:
  - 新增 `GET /api/v1/ai-procurement/monitor`
- 请求参数:
  - `page`、`pageSize`、`keyword`、`status`、`stage`、`intervention`
- 响应结构变化:
  - 新增监控摘要、阶段分布、状态分布、配置摘要和分页记录。
- 受影响服务层: `frontend/src/services/api.ts`。
- 受影响状态管理: 无。
- 受影响后端 handler/service/repository: `backend/internal/router`、`backend/internal/handler`。
- 受影响数据库或缓存: 只读 `ai_procurements`、`admin_system_configs`；不新增表。
- 配置或环境变量变化: 无。
- 构建或部署影响: 需要部署 admin 后端和前端。

## 权限与合规影响

- 受影响角色: 超管、运营。
- 新增/调整权限: 运营可查看 AI 采购运营监控；AI 采购配置仍仅超管可查看/修改。
- 是否影响审核、风控、日志留痕: 不新增写操作；仅提供运营监控和异常排查视图。

## 验收标准

1. 超管和运营能打开 `/ai-procurement/monitor`，普通账号和审核员不能直接读取监控接口。
2. 监控页能展示统计卡片、阶段分布、状态分布、配置摘要和最近记录。
3. 支持按关键词、状态、阶段和待介入筛选最近记录。
4. `ai_procurements` 表不存在时接口返回空数据和 `tableReady=false`，前端展示空状态。
5. 后端测试、后端构建、前端测试和前端构建通过。
