# SL-ADMIN-CATEGORY-PRODUCT-COUNT Spec

## 基本信息

- Feature ID: SL-ADMIN-CATEGORY-PRODUCT-COUNT
- Feature 名称: 类目商品数量与删除保护
- 负责人: Codex
- 目标上线环境: Admin production

## 背景

管理端「商品管理 / 类目配置」可以直接删除类目，但页面没有展示每个类目下已有多少商品。运营在整理类目时无法判断删除影响，容易误删仍被商品使用的类目。

## 范围

### In Scope

- 类目列表接口返回每条类目的关联商品数量。
- 类目配置表格新增「商品数量」列。
- 有关联商品的类目在页面上禁用删除，并展示不能删除的原因。
- 后端删除接口增加保护，类目下仍有关联商品时返回 `409`。

### Out of Scope

- 不迁移现有商品类目字段。
- 不提供批量类目合并或商品改绑能力。
- 不调整买家端类目展示逻辑。

## 页面与交互影响

- 受影响页面: 商品管理 / 类目配置
- 受影响路由: `/product/category`
- 受影响布局或导航: 无
- 受影响组件: 类目表格、删除操作
- 受影响交互状态: 商品数量大于 0 时删除按钮禁用，悬停展示提示；数量为 0 时保留二次确认。

## 接口与数据影响

- 受影响接口: `GET /api/v1/product/categories`、`DELETE /api/v1/product/categories/:id`
- 请求参数或响应结构变化: 类目响应新增 `productCount`
- 受影响服务层: `frontend/src/services/api.ts`
- 受影响状态管理: React Query 类目列表缓存刷新逻辑保持不变
- 受影响后端 handler/service/repository: `backend/internal/handler/handler.go`
- 受影响数据库或缓存: 无 schema 变更；读取 `products` 表统计关联数量
- 配置或环境变量变化: 无
- 构建或部署影响: 需要部署 admin backend 与 admin frontend

## 权限与合规影响

- 受影响角色: 管理端商品运营、管理员
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 不影响审核流程；删除操作增加服务端约束。

## 验收标准

1. 类目配置列表每条记录能看到当前关联商品数量。
2. 关联商品数量大于 0 的类目不能从页面发起删除。
3. 即使绕过页面直接调用删除接口，仍会收到 `409` 且不会删除有关联商品的类目。
