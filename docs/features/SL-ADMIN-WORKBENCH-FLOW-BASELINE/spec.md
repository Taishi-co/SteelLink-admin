# Feature Spec

## 基本信息

- Feature ID: `SL-ADMIN-WORKBENCH-FLOW-BASELINE`
- Feature 名称: 管理端仓库地图与部署 harness 基线校准
- 负责人: Codex
- 目标上线环境: 本地开发 / 测试联调环境

## 背景

管理端仓库已经具备 `frontend/` React 应用、`backend/` Go/Gin 服务，以及 `Dockerfile.admin-backend`、`Dockerfile.admin-frontend` 两个镜像入口。但 `docs/ai/project-map.md` 仍写着业务代码尚未初始化，`harness/deploy.sh` 也只识别根目录 `Dockerfile` 或 compose 文件，无法准确产出当前管理端的部署计划。

本轮目标是让 agent 文档和部署 harness 与真实仓库结构一致，减少后续优化五个工作台时的流程误判。

## 范围

### In Scope

- 更新 `docs/ai/project-map.md`，反映当前前后端同仓结构。
- 更新 `harness/deploy.sh`，识别 `Dockerfile.admin-backend` 与 `Dockerfile.admin-frontend`。
- dry-run 模式输出后端、前端两个镜像构建计划。
- 保留 `DEPLOY_COMMAND` 和通用 Dockerfile/Compose 兜底能力。

### Out of Scope

- 不新增或修改管理端业务页面。
- 不实现 `/admin/notifications` 缺失接口。
- 不调整数据库结构、JWT、菜单权限模型。
- 不执行真实部署。

## 页面与交互影响

- 受影响页面: 无
- 受影响路由: 无
- 受影响布局或导航: 无
- 受影响组件: 无
- 受影响交互状态: 无

## 接口与数据影响

- 受影响接口: 无
- 请求参数或响应结构变化: 无
- 受影响服务层: 无
- 受影响状态管理: 无
- 受影响后端 handler/service/repository: 无
- 受影响数据库或缓存: 无
- 配置或环境变量变化:
  - 可选读取 `ADMIN_BACKEND_IMAGE`
  - 可选读取 `ADMIN_FRONTEND_IMAGE`
  - 可选读取 `ADMIN_FRONTEND_API_BASE_URL`
- 构建或部署影响:
  - `./harness/deploy.sh <staging|prod>` dry-run 可识别两个 admin Dockerfile。

## 权限与合规影响

- 受影响角色: 管理端运维/发布执行者
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 不影响

## 验收标准

1. `docs/ai/project-map.md` 不再描述仓库“尚未初始化具体业务代码”。
2. `./harness/deploy.sh staging` 在 dry-run 模式输出 admin backend/frontend 两个镜像构建计划。
3. 保留 `DEPLOY_COMMAND` 优先级，不破坏自定义部署入口。
4. `./harness/verify.sh` 的 shell syntax 校验通过。
