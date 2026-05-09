# Feature Spec

## 基本信息

- Feature ID: SL-ADMIN-NOTIFICATIONS
- Feature 名称: 管理端站内信发送
- 负责人: Codex
- 目标上线环境: staging / prod

## 背景

运营人员需要从管理端向用户发送平台通知类站内信。

## 范围

### In Scope

- 新增管理端站内信发送页面。
- 支持按用户 ID、企业 ID、用户角色或全量发送。
- 调用主后端 `/api/v1/admin/notifications`。

### Out of Scope

- 不做历史发送记录查询。
- 不做富文本和附件。

## 页面与交互影响

- 受影响页面: AdminLayout
- 受影响路由: `/notification`
- 受影响布局或导航: fallback 菜单新增站内信入口
- 受影响组件: 无
- 受影响交互状态: 表单提交、重置、发送中 loading

## 接口与数据影响

- 受影响接口: `POST /admin/notifications`
- 请求参数或响应结构变化: 新增 `notificationApi.send`
- 受影响服务层: frontend service
- 受影响数据库或缓存: 由主后端处理
- 配置或环境变量变化: 无
- 构建或部署影响: 无

## 权限与合规影响

- 受影响角色: 管理员、运营人员
- 新增/调整权限: 无独立前端权限码，菜单回退模式可访问；真实权限由后端校验。
- 是否影响审核、风控、日志留痕: 管理员发送会写入用户站内信。

## 验收标准

1. 管理端可打开 `/notification`。
2. 表单可提交站内信请求。
3. 发送成功后提示创建消息数量。
