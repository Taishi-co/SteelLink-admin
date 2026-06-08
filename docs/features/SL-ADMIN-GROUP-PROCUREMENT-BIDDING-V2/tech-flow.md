# 集采竞价技术流转说明

## 目标

这份说明用于帮助 review 和回归当前“集采竞价”主流程。当前实现已经统一收口到“采购申请 -> 竞价项目 -> 卖家报价 -> 定标 -> 成单”模型，旧 `procurement_activities` 管理端活动流已下线，不再作为运行入口。

## 主流程

### 1. 买家侧沉淀采购申请

- 买家提交采购需求后，后端落到 `procurement_requests`
- 管理端把这批记录视为“待归并采购申请池”
- 当前管理端入口页面:
  - `frontend/src/pages/group-procurement/index.tsx`

### 2. 管理端归并申请并创建竞价项目

- 管理端从待处理申请中勾选多条记录
- 调用 `POST /group-procurement-projects`
- 后端处理入口:
  - `backend/internal/handler/group_procurement_bidding.go`
  - `CreateGroupProcurementProjectV2`
- 核心动作:
  - 创建 `procurement_projects`
  - 建立 `procurement_project_requests`
  - 将 `procurement_requests.status` 更新为 `selected`
  - 写入 `procurement_logs`

### 3. 管理端邀请卖家参与报价

- 管理端在项目详情中选择卖家商品
- 调用 `POST /group-procurement-projects/:id/invitations`
- 后端核心动作:
  - 读取卖家商品快照
  - 创建 `procurement_supplier_enrollments`
  - 项目状态切到 `quoting`
  - 更新邀请卖家数

### 4. 卖家确认参与并提交报价

- 卖家在商城卖家端查看自己被邀请的项目
- 卖家提交报价后，后端写入 `procurement_supplier_quotes`
- 管理端详情接口 `GET /group-procurement-projects/:id` 会返回:
  - `requests`
  - `enrollments`
  - `quotes`

### 5. 管理端定标

- 管理端在报价对比表中选择一个有效报价
- 调用 `POST /group-procurement-projects/:id/award`
- 后端核心校验:
  - 报价必须属于当前项目
  - 报价必须是最新有效报价
  - 报价可供量必须覆盖项目总需求
- 写入结果:
  - `procurement_project_awards`
  - `procurement_projects.status = awarded`
  - 中标卖家、中标商品、中标价、交期等字段回填到项目表

### 6. 管理端生成订单

- 管理端在已定标项目上点击“生成订单”
- 调用 `POST /group-procurement-projects/:id/order`
- 后端核心动作:
  - 校验项目已定标且未成单
  - 基于中标商品锁库存
  - 生成 1 张母订单 + N 张子订单
  - 将 `procurement_requests.status` 更新为 `accepted`
  - 将 `procurement_projects.status` 更新为 `order_created`
  - 回填 `order_id`、`order_no`

## 当前关键表

- `procurement_requests`
  - 买家采购申请池
- `procurement_projects`
  - 集采竞价项目主表
- `procurement_project_requests`
  - 项目与采购申请映射
- `procurement_supplier_enrollments`
  - 卖家受邀/参与记录
- `procurement_supplier_quotes`
  - 卖家报价记录
- `procurement_project_awards`
  - 定标结果
- `procurement_logs`
  - 项目过程日志

## 当前关键状态

### 项目状态 `procurement_projects.status`

- `collecting`
  - 已建项目，仍在归并需求
- `quoting`
  - 已邀请卖家，等待报价
- `awarded`
  - 已定标
- `order_created`
  - 已生成订单
- `closed`
  - 关闭

### 采购申请状态 `procurement_requests.status`

- `pending`
  - 待管理端处理
- `selected`
  - 已被归并进竞价项目
- `rejected`
  - 被管理端驳回
- `accepted`
  - 已随中标项目成单

## 管理端接口分层

### 采购申请池

- `GET /group-procurement-requests`
- `POST /group-procurement-requests/:id/reject`

### 竞价项目

- `GET /group-procurement-projects`
- `GET /group-procurement-projects/:id`
- `POST /group-procurement-projects`
- `POST /group-procurement-projects/:id/invitations`
- `POST /group-procurement-projects/:id/award`
- `POST /group-procurement-projects/:id/order`

## 代码边界

### 管理端前端

- 页面:
  - `frontend/src/pages/group-procurement/index.tsx`
- 服务:
  - `frontend/src/services/api.ts`
  - `groupProcurementRequestApi`
  - `groupProcurementProjectApi`

### 管理端后端

- 竞价项目主逻辑:
  - `backend/internal/handler/group_procurement_bidding.go`
- 采购申请池逻辑:
  - `backend/internal/handler/group_procurement_requests.go`
- 共享 schema / 订单 helper:
  - `backend/internal/handler/group_procurement.go`

## 本轮清理后的结论

- 旧管理端活动接口 `/group-procurements*` 已移除
- 启动时旧活动自动收尾任务已停用
- 旧活动 handler 和自动成单死代码已删除
- `group_procurement.go` 现仅保留项目制复用的共享基础设施

## 回归建议

- 重点回归 4 个动作:
  - 采购申请列表加载与驳回
  - 多申请创建项目
  - 邀请卖家后卖家侧是否能看到项目并报价
  - 定标后生成母子订单
- 重点看 3 类状态:
  - `procurement_requests.status`
  - `procurement_projects.status`
  - 订单主子单是否都成功生成
