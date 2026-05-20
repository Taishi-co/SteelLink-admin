# Role Permission Matrix

本文件沉淀管理端角色、菜单、页面动作与权限点矩阵。前端可见性与后端接口授权必须保持一致；涉及审核、风控、财务、导出、批量操作时必须明确到动作级别。

## 角色定义

| 角色 | role_id | 说明 |
| --- | --- | --- |
| 超级管理员 | 1 | 系统管理、权限配置、全部业务查看、审核和运营处置 |
| 运营管理员 | 2 | 业务查看、审核、订单运营处置、站内信 |
| 审核员 | 3 | 业务查看、企业/实名/商品/服务资源审核 |
| 普通管理端账号 | 其他 | 默认只保留登录后的查看类能力，不能执行审核或运营处置 |

## 权限矩阵

| 模块 | 页面/菜单 | 动作 | 普通账号 | 审核员 | 运营管理员 | 超级管理员 | 后端约束 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 数据概览 | `/dashboard` | 查看 | 是 | 是 | 是 | 是 | 登录态 |
| 用户管理 | `/user` | 查看 | 是 | 是 | 是 | 是 | 登录态 |
| 用户管理 | `/user` | 启用/停用 | 否 | 否 | 是 | 是 | `RequireRole(1,2)` |
| 用户管理 | `/user` | 实名审核 | 否 | 是 | 是 | 是 | `RequireRole(1,2,3)` |
| 企业审核 | `/enterprise` | 查看 | 是 | 是 | 是 | 是 | 登录态 |
| 企业审核 | `/enterprise` | 企业资质审核 | 否 | 是 | 是 | 是 | `RequireRole(1,2,3)` |
| 服务权限审核 | `/enterprise/services` | 查看 | 是 | 是 | 是 | 是 | 登录态 |
| 服务权限审核 | `/enterprise/services` | 申请审核 | 否 | 是 | 是 | 是 | `RequireRole(1,2,3)` |
| 服务权限审核 | `/enterprise/:id/services` | 直接调整权限状态 | 否 | 否 | 是 | 是 | `RequireRole(1,2)` |
| 商品管理 | `/product` | 查看 | 是 | 是 | 是 | 是 | 登录态 |
| 商品管理 | `/product` | 商品审核 | 否 | 是 | 是 | 是 | `RequireRole(1,2,3)` |
| 商品管理 | `/product` | 编辑/删除 | 否 | 否 | 是 | 是 | `RequireRole(1,2)` |
| 服务资源审核 | `/service-offers/warehouse` | 仓储资源审核 | 否 | 是 | 是 | 是 | `RequireRole(1,2,3)` |
| 服务资源审核 | `/service-offers/logistics` | 物流资源审核 | 否 | 是 | 是 | 是 | `RequireRole(1,2,3)` |
| 服务资源审核 | `/service-offers/finance` | 金融服务审核 | 否 | 是 | 是 | 是 | `RequireRole(1,2,3)` |
| 订单管理 | `/product/order` | 查看 | 是 | 是 | 是 | 是 | 登录态 |
| 订单管理 | `/product/order` | 冻结/解冻/违约协调/诉讼状态 | 否 | 否 | 是 | 是 | `RequireRole(1,2)` |
| 类目配置 | `/product/category` | 查看 | 是 | 是 | 是 | 是 | 登录态 |
| 类目配置 | `/product/category` | 新增/编辑/删除 | 否 | 否 | 否 | 是 | `RequireRole(1)` |
| AI采购配置 | `/ai-procurement/config` | 查看/修改 | 否 | 否 | 否 | 是 | `RequireRole(1)` |
| 站内信 | `/notification` | 发送 | 否 | 否 | 是 | 是 | `RequireRole(1,2)` |
| 系统管理 | `/system/*` | 管理员、角色、菜单管理 | 否 | 否 | 否 | 是 | `RequireRole(1)` |
| 操作日志 | `/log` | 查看 | 是 | 是 | 是 | 是 | 登录态 |

## 默认菜单分配

| 角色 | 默认菜单 |
| --- | --- |
| 超级管理员 | 全部菜单 |
| 运营管理员 | `/dashboard`、业务审核页、服务资源审核页、`/product`、`/product/order`、`/product/category`、`/log`、`/notification` |
| 审核员 | `/dashboard`、业务审核页、服务资源审核页、`/product`、`/product/order`、`/product/category`、`/log` |
| 普通管理端账号 | 不自动分配菜单，由超级管理员显式配置 |

前端动作按钮按同一矩阵显隐；后端接口仍是最终授权边界。

## 待补充信息

- 菜单按钮级 permission code 与后端角色分组的统一生成规则。
- 普通管理端账号是否应继续保留所有查看类接口，或进一步收敛为菜单授权。
- 导出、批量操作、资金审批代理等后续动作级权限。
