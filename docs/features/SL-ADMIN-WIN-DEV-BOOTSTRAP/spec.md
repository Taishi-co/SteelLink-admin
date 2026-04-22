# Feature Spec

## 基本信息

- Feature ID: `SL-ADMIN-WIN-DEV-BOOTSTRAP`
- Feature 名称: 管理端本地 Windows 启动与默认管理员登录补齐
- 负责人: 二号开发者
- 目标上线环境: 本地开发环境

## 背景

当前管理端前端与后端已有基础骨架，但本地 Windows 开发存在两类阻塞：

- 后端登录仅返回 mock token，无法稳定支撑管理端前端联调。
- Windows 本机缺少统一的启动入口，重复输入命令成本高，也容易与非 Windows 同事的默认工作流混淆。

本次改动目标是补齐本地开发必需能力，不触碰服务器部署逻辑。

## 范围

### In Scope

- 管理端后端补齐默认管理员登录、JWT 校验、当前管理员信息接口。
- 为 Windows 本地开发新增独立启动脚本。
- 在 `.env.example` 中增加本地开发默认管理员环境变量说明。

### Out of Scope

- 真实数据库管理员体系。
- 服务器部署、容器重启或线上配置修改。
- 完整 RBAC、管理员 CRUD 真正落库。

## 页面与交互影响

- 受影响页面: 管理端登录页
- 受影响路由: 登录成功后的管理员信息获取链路
- 受影响布局或导航: 无
- 受影响组件: 登录态相关请求链路
- 受影响交互状态:
  - `admin / admin123` 可完成本地登录
  - 登录后可读取当前管理员信息

## 接口与数据影响

- 受影响接口:
  - `POST /api/v1/login`
  - `POST /api/v1/logout`
  - `GET /api/v1/admin/info`
- 请求参数或响应结构变化:
  - 登录响应由 mock token 改为真实 JWT 和 `adminInfo`
- 受影响服务层: 管理端 handler / middleware / router
- 受影响状态管理: 管理端前端登录态可使用真实返回数据
- 受影响后端 handler/service/repository:
  - `internal/handler`
  - `internal/middleware`
  - `internal/router`
  - 新增 `internal/auth`
- 受影响数据库或缓存: 无
- 配置或环境变量变化:
  - `ADMIN_DEFAULT_USERNAME`
  - `ADMIN_DEFAULT_PASSWORD`
  - `ADMIN_DEFAULT_REAL_NAME`
- 构建或部署影响:
  - 仅新增 Windows 本地启动脚本

## 权限与合规影响

- 受影响角色: 本地开发默认管理员
- 新增/调整权限: 无新增业务权限，仅补齐本地鉴权闭环
- 是否影响审核、风控、日志留痕: 否

## 验收标准

1. 本地启动管理端后端后，`/health` 返回成功。
2. 使用 `admin / admin123` 调用登录接口可获取 JWT 与管理员信息。
3. 携带登录 token 访问 `/api/v1/admin/info` 返回当前管理员信息。
