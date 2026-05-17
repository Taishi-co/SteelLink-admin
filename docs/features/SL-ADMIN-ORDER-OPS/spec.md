# SL-ADMIN-ORDER-OPS Spec

## 基本信息

- Feature ID: SL-ADMIN-ORDER-OPS
- Feature 名称: 管理端订单详情与运营处置
- 负责人: Codex
- 目标上线环境: staging

## 背景

管理端订单页面已调用 `GET /orders/:id` 加载完整履约详情，但 admin 后端未注册该路由，导致详情 Drawer 只能依赖列表数据兜底。业务后端已有订单事件、融资、物流、违约、冻结与诉讼能力，管理端需要补齐查看与运营介入入口。

## 范围

### In Scope

- 新增 admin 后端订单详情接口。
- 返回订单事件、融资、物流竞价、运输任务、质押、违约、诉讼等详情。
- 新增 admin 后端冻结/解冻、违约协调/处理、诉讼状态更新接口。
- 前端订单详情 Drawer 展示运营处置区域与异常记录。
- 校准订单状态文案与待运营介入筛选。

### Out of Scope

- 不重构业务后端订单状态机。
- 不新增数据库表结构。
- 不实现跨服务 token 代理调用。

## 页面与交互影响

- 受影响页面: `frontend/src/pages/order/index.tsx`
- 受影响路由: `/order`
- 受影响布局或导航: 无
- 受影响组件: 订单详情 Drawer、订单筛选器
- 受影响交互状态: 加载详情、冻结/解冻、违约协调、违约处理、诉讼状态更新

## 接口与数据影响

- 受影响接口:
  - `GET /api/v1/orders/:id`
  - `POST /api/v1/orders/:id/freeze`
  - `POST /api/v1/orders/:id/unfreeze`
  - `POST /api/v1/orders/breach-records/:id/mediate`
  - `POST /api/v1/orders/breach-records/:id/resolve`
  - `PUT /api/v1/orders/litigation-records/:id`
- 请求参数或响应结构变化: 订单详情新增履约与争议聚合字段。
- 受影响服务层: `frontend/src/services/api.ts`
- 受影响状态管理: 无
- 受影响后端 handler/service/repository: `backend/internal/handler/handler.go`、`backend/internal/router/router.go`
- 受影响数据库或缓存: 读取既有订单相关表，运营动作更新既有订单、违约与诉讼表并写入 `order_events`。
- 配置或环境变量变化: 无
- 构建或部署影响: 无

## 权限与合规影响

- 受影响角色: 已登录管理端用户。
- 新增/调整权限: 复用管理端鉴权保护路由，未新增菜单权限点。
- 是否影响审核、风控、日志留痕: 影响风控/运营留痕，动作写入 `order_events` 与 `admin_operation_logs`。

## 验收标准

1. 订单详情 Drawer 能通过 `GET /orders/:id` 获取完整履约信息。
2. 管理员可在订单详情中冻结/解冻订单。
3. 管理员可在违约记录中发起协调、完成处理，并可更新诉讼状态。
4. 订单列表支持待运营介入筛选并显示业务后端当前状态文案。
