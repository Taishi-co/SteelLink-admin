# SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS Spec

## 基本信息

- Feature ID: SL-ADMIN-SERVICE-OFFER-AUDIT-STATUS
- Feature 名称: 修复服务资源审核通过被落为驳回
- 负责人: Codex
- 目标上线环境: prod

## 背景

仓储方在 `/warehouse/` 新增仓库资源后，管理端审核通过，但仓储工作台显示“已驳回”。根因是管理端前端按资源状态枚举提交 `2=已通过、3=已驳回`，而管理端后端审核接口仍把 `1=通过、2=驳回` 当成动作枚举，导致 `status=2` 被写成资源状态 `3`。

## 范围

### In Scope

- 管理端服务资源审核接口按真实资源状态 `2=已通过、3=已驳回` 落库。
- 前端审核请求补充明确的 `auditAction`，避免动作枚举与状态枚举再次混用。
- 已驳回资源允许重新审核，便于修复误驳回记录。
- 覆盖仓储、物流、金融三类服务资源审核。

### Out of Scope

- 不调整仓储工作台资源列表接口。
- 不修改服务资源数据库表结构。
- 不处理线上已有误驳回数据的自动迁移。

## 页面与交互影响

- 受影响页面: 服务资源审核、仓储资源审核、物流资源审核、金融服务审核
- 受影响路由: `/service-offer/warehouse`、`/service-offer/logistics`、`/service-offer/finance`
- 受影响布局或导航: 无
- 受影响组件: `ServiceOfferAudit`
- 受影响交互状态: 已驳回资源可重新打开审核弹窗

## 接口与数据影响

- 受影响接口:
  - `PUT /warehouse-services/:id/audit`
  - `PUT /logistics-services/:id/audit`
  - `PUT /finance-services/:id/audit`
- 请求参数或响应结构变化: 请求兼容新增 `auditAction` / `audit_action`，保留 `status`
- 受影响服务层: `serviceOfferApi`
- 受影响状态管理: 无
- 受影响后端 handler/service/repository: `backend/internal/handler/service_offer.go`
- 受影响数据库或缓存: 只修正写入的 `status` 与 `approved_at`
- 配置或环境变量变化: 无
- 构建或部署影响: 需要同时部署管理端前后端

## 权限与合规影响

- 受影响角色: 管理员、运营、超管
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 影响服务资源审核结果写入，修复误写为驳回的问题。

## 验收标准

1. 管理端点击“通过”后，服务资源表 `status` 写入 `2`，仓储工作台显示“已通过”。
2. 管理端点击“驳回”后，服务资源表 `status` 写入 `3`，仓储工作台显示“已驳回”。
3. 已驳回服务资源可以在管理端重新审核通过。
