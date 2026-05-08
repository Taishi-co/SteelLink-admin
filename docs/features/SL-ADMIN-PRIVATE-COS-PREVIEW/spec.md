# SL-ADMIN-PRIVATE-COS-PREVIEW Spec

## 基本信息

- Feature ID: SL-ADMIN-PRIVATE-COS-PREVIEW
- Feature 名称: 管理端私有 COS 资料预览
- 负责人: Codex
- 目标上线环境: production

## 背景

实名认证身份证照片上传到 `steellink-user-pri-1321917249` 私有 COS 桶。管理端审核页直接把私有桶 URL 传给浏览器 `<img>`，没有签名参数，COS 会返回 403 Forbidden，导致审核人员无法查看证件照片。

## 范围

### In Scope

- 新增已登录管理端可调用的私有 COS URL 临时签名接口。
- 用户审核详情页展示身份证正反面前，先将原始私有 URL 换成短期签名 URL。
- 补充管理端生产环境 COS 读取配置示例。

### Out of Scope

- 不调整上传路径和 COS 桶 ACL。
- 不把私有身份证材料改成公开可读。
- 不新增独立文件审计日志表。

## 页面与交互影响

- 受影响页面: 用户管理 / 实名认证审核详情抽屉
- 受影响路由: `/user`
- 受影响布局或导航: 无
- 受影响组件: 身份证正反面 `Image`
- 受影响交互状态: 图片预览从原始私有 URL 切换为短期签名 URL

## 接口与数据影响

- 受影响接口: 新增 `GET /api/v1/files/private-url?url=<cos-url>`
- 请求参数或响应结构变化: 返回 `{ url, expiresIn }`
- 受影响服务层: 无独立 service
- 受影响状态管理: 无
- 受影响后端 handler/service/repository: `backend/internal/handler/private_file.go`
- 受影响数据库或缓存: 无
- 配置或环境变量变化: `COS_REGION`, `COS_SECRET_ID`, `COS_SECRET_KEY`, `COS_PRIVATE_BUCKET`, `COS_ADMIN_BUCKET`, `COS_PRIVATE_URL_EXPIRES_SECONDS`
- 构建或部署影响: 管理端后端需注入 COS 读取密钥；密钥不写入仓库

## 权限与合规影响

- 受影响角色: 已登录管理端用户
- 新增/调整权限: 新增接口复用现有管理端登录鉴权，不开放匿名访问
- 是否影响审核、风控、日志留痕: 影响实名认证审核资料读取；不改变审核结果写入逻辑

## 验收标准

1. 私有桶身份证 URL 通过管理端后端换签后可在审核详情中展示。
2. 非配置桶或非 COS HTTPS URL 不能被换签。
3. 未配置 COS 密钥时接口返回明确配置错误，不静默公开私有资料。
