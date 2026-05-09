# SL-ADMIN-PRODUCT-STATUS-MAPPING Spec

## 基本信息

- Feature ID: SL-ADMIN-PRODUCT-STATUS-MAPPING
- Feature 名称: 商品审核状态口径修正
- 负责人: Codex
- 目标上线环境: Admin production

## 背景

管理端商品列表使用 `status=1` 筛选“待审核”，但后端把该参数映射成数据库 `status=2`，再把返回状态 `2` 映射成前端状态 `1`。结果是已经上架的商品会在管理端显示为待审核，造成运营误判。

## 范围

### In Scope

- 商品列表筛选参数按真实商品状态 `0/1/2/3/4` 查询。
- 商品列表与详情响应返回真实商品状态，不再压缩成审核动作状态。
- 首页待审核统计只统计真实待审核商品。

### Out of Scope

- 不修改买家端商城商品展示过滤。
- 不自动清洗生产商品数据。
- 不修改商品审核提交接口的动作值约定。

## 页面与交互影响

- 受影响页面: 商品管理、首页待处理事项
- 受影响路由: `/product`
- 受影响组件: 商品列表状态筛选、审核状态标签、商城可见性标签
- 受影响交互状态: `status=1` 只返回待审核，`status=2` 返回已上架。

## 接口与数据影响

- 受影响接口: `GET /api/v1/products`、`GET /api/v1/products/:id`、`GET /api/v1/dashboard/stats`
- 请求参数或响应结构变化: 无新增字段；修正 `status` 筛选与响应值语义。
- 受影响服务层: 无
- 受影响状态管理: React Query 缓存 key 不变
- 受影响后端 handler/service/repository: `backend/internal/handler/handler.go`、`backend/internal/handler/data.go`
- 受影响数据库或缓存: 无 schema 变更
- 配置或环境变量变化: 无
- 构建或部署影响: 需要部署 admin backend

## 权限与合规影响

- 受影响角色: 管理员、商品审核员、运营
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 影响审核列表展示口径，不改变审核写入逻辑。

## 验收标准

1. 请求 `/api/v1/products?status=1` 只返回数据库 `status=1` 的待审核商品。
2. 请求 `/api/v1/products?status=2` 返回数据库 `status=2` 的已上架商品，并在页面展示为已上架。
3. 商品详情返回的 `status` 与数据库状态一致。
