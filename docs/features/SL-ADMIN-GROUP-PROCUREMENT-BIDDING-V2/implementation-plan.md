# Implementation Plan

## Feature 信息

- Feature ID: `SL-ADMIN-GROUP-PROCUREMENT-BIDDING-V2`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `frontend/src/pages/group-procurement/*`
  - `frontend/src/services/*`
  - `backend/internal/handler/group_procurement.go`
  - 相关 router / dto / query 代码
- 文档文件:
  - `docs/features/SL-ADMIN-GROUP-PROCUREMENT-BIDDING-V2/*`
- 配置文件:
  - 无
- 脚本文件:
  - 无

## 实施步骤

1. 重构管理端 group procurement 接口和服务层，接入项目、报名、报价、定标和成单能力。
2. 改造集采管理页为项目制工作台，增加申请归并、卖家邀请、报价对比和定标 UI。
3. 补齐页面状态流转、错误提示、操作留痕展示和成单结果跳转。

## 验证计划

- `./harness/verify.sh`
- 如涉及配套应用仓，补充对应 lint/build/测试命令
- 需要的人工验证:
  - 管理员创建项目并邀请卖家
  - 审核报名、定标、成单

## 风险与回滚关注点

- 页面从活动制切到项目制，旧字段兼容映射不完整会导致详情页空态。
- 定标与成单按钮必须严格依赖后端状态，否则容易误触发。
- 回滚时可恢复旧管理页和旧 handler 查询逻辑，新增后端字段保持兼容。
