# SL-ADMIN-FLOW-BUGFIX-20260520 Spec

## 基本信息

- Feature ID: `SL-ADMIN-FLOW-BUGFIX-20260520`
- Feature 名称: 管理端主流程配合度缺陷修复
- 负责人: Codex
- 目标上线环境: staging / prod

## 背景

对照 RAG 产品流程图和后台实现后，发现管理端已覆盖商品、企业、服务资源、订单运营处置和 AI 采购配置，但在动作权限、订单解冻、违约处理、待办统计、履约时间线、初始化菜单与内部配置接口上存在流程配合风险。

## 范围

### In Scope

- 审核类接口限定为超管、运营、审核员。
- 订单运营处置类接口限定为超管、运营。
- 解冻订单必须处于冻结状态，且必须存在有效冻结事件才能恢复。
- 违约处理只更新违约记录和事件，不直接把异常订单改为已完成。
- 待办统计改为真实数字状态枚举。
- 订单时间线改为仓库、合同、支付/融资、物流、验收、对账、归档主流程。
- 初始化菜单和兼容路由与真实前端路由对齐。
- 默认运营/审核员菜单与角色矩阵对齐。
- 前端按角色隐藏无权执行的审核、运营处置、类目维护和站内信发送动作。
- 菜单接口成功返回空权限时不再回退展示全量菜单。

### Out of Scope

- 不重构业务后端订单状态机。
- 不新增资金方或物流方审批页面。
- 不新增真实合同签署和支付网关能力。
- 不自动清理历史错误菜单或线上异常订单数据。

## 页面与交互影响

- 受影响页面: 订单管理、商品管理、企业审核、服务权限审核、服务资源审核、AI 采购配置。
- 受影响路由: `/user/list`、`/enterprise/list`、`/product/list` 增加兼容跳转。
- 受影响布局或导航: 初始化菜单补齐服务审核、订单、AI 采购、日志、站内信入口。
- 受影响组件: 订单详情 Drawer 时间线。
- 受影响交互状态: 无权限账号不显示对应写操作按钮；直接调用审核/运营接口仍返回 403。

## 接口与数据影响

- 受影响接口:
  - `PUT /api/v1/users/:id/identity`
  - `PUT /api/v1/enterprises/:id/audit`
  - `PUT /api/v1/enterprise-service-applications/:id/audit`
  - `PUT /api/v1/products/:id/audit`
  - `PUT /api/v1/*-services/:id/audit`
  - `POST /api/v1/orders/:id/freeze`
  - `POST /api/v1/orders/:id/unfreeze`
  - `POST /api/v1/orders/breach-records/:id/mediate`
  - `POST /api/v1/orders/breach-records/:id/resolve`
  - `PUT /api/v1/orders/litigation-records/:id`
- 请求参数或响应结构变化: 无。
- 受影响服务层: `frontend/src/router/index.tsx`。
- 受影响前端权限工具: `frontend/src/utils/permissions.ts`。
- 受影响状态管理: 无。
- 受影响后端 handler/service/repository: `backend/internal/router`、`backend/internal/handler`。
- 受影响数据库或缓存: 初始化 SQL 与 AutoMigrate 菜单种子；不做强制数据迁移。
- 配置或环境变量变化: 无。
- 构建或部署影响: 需要同时部署 admin 后端和前端。

## 权限与合规影响

- 受影响角色: 超管、运营、审核员、普通管理端账号。
- 新增/调整权限: 审核员可做审核动作；运营可做审核和订单运营处置；普通已登录账号只保留查看类能力。
- 是否影响审核、风控、日志留痕: 是。订单冻结、解冻、违约、诉讼继续写入 `order_events` 与 `admin_operation_logs`。

## 验收标准

1. 普通管理端账号不能直接调用审核和订单运营处置接口。
2. 非冻结订单调用解冻接口不再被改成异常。
3. 违约处理不再把异常订单直接改成已完成。
4. 订单详情时间线不再把待重新选仓、融资中、物流竞价等状态误显示为已提货/已出库。
5. 新库初始化菜单能进入真实页面，历史 `/user/list` 等路径能跳转到真实页面。
6. 运营和审核员默认菜单能覆盖其主流程页面。
7. 审核员看不到订单冻结/解冻、企业能力直接控制、站内信发送等运营按钮。
8. 菜单接口成功返回空数组时，前端展示无权访问而不是全量 fallback 菜单。
9. 后端测试、后端构建和前端构建通过。
