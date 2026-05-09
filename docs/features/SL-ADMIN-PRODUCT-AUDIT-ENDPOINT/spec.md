# SL-ADMIN-PRODUCT-AUDIT-ENDPOINT Spec

## 基本信息

- Feature ID: SL-ADMIN-PRODUCT-AUDIT-ENDPOINT
- Feature 名称: 修复管理端商品审核接口路径和动作值
- 负责人: Codex
- 目标上线环境: prod

## 背景

管理端商品审核页 `/product` 提交审核时请求 `PUT /api/v1/products/:id/status`，但管理端后端实际注册的是 `PUT /api/v1/products/:id/audit`，导致生产环境返回 404。

同时，管理端后端 `AuditProduct` 使用审核动作值 `1=通过、2=驳回`，再落库为商品状态 `2=已上架、4=审核拒绝`。前端当前把商品状态值 `2/4` 直接作为审核动作提交，会造成通过/驳回语义错位。

## 范围

### In Scope

- 商品审核请求路径从 `/products/:id/status` 改为 `/products/:id/audit`。
- 商品审核弹窗提交动作值改为后端约定的 `1=通过、2=驳回`。
- 保留审核备注字段兼容 camelCase 和 snake_case。

### Out of Scope

- 不调整商品列表、详情、类目接口。
- 不修改商品状态枚举和数据库表结构。
- 不修改主站后端 `SteelLink-backend` 的 `/api/v1/products/:id/status` 路由。

## 页面与交互影响

- 受影响页面: 商品管理
- 受影响路由: `/product`
- 受影响布局或导航: 无
- 受影响组件: 商品审核弹窗
- 受影响交互状态: 通过上架、驳回

## 接口与数据影响

- 受影响接口: `PUT /api/v1/products/:id/audit`
- 请求参数或响应结构变化: 前端提交 `status=1|2` 审核动作，附带 `auditRemark`
- 受影响服务层: `frontend/src/services/api.ts`
- 受影响状态管理: React Query 商品列表和详情缓存刷新逻辑不变
- 受影响后端 handler/service/repository: 无代码变更，仅对齐既有 `AuditProduct`
- 受影响数据库或缓存: 无
- 配置或环境变量变化: 无
- 构建或部署影响: 需要重新部署管理端前端

## 权限与合规影响

- 受影响角色: 管理员、运营、超管
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 修复商品审核请求进入既有审核 handler，恢复审核日志写入。

## 验收标准

1. 在 `/product` 审核商品通过时，请求 `PUT /api/v1/products/:id/audit` 返回成功，商品状态变为已上架。
2. 在 `/product` 审核商品驳回时，请求 `PUT /api/v1/products/:id/audit` 返回成功，商品状态变为审核拒绝。
3. 不再出现 `PUT /api/v1/products/:id/status` 的 404。
