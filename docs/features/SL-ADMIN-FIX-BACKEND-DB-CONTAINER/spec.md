# Backend Container MySQL Config Guard

## 基本信息

- Feature ID: SL-ADMIN-FIX-BACKEND-DB-CONTAINER
- Feature 名称: 修复后端容器缺失 MySQL 主机配置时反复连接 localhost
- 负责人: Codex
- 目标上线环境: prod

## 背景

生产后端容器启动后持续报错：

```text
Failed to connect MySQL: ping mysql: dial tcp [::1]:3306: connect: connection refused
```

当前后端 `DB_HOST` 缺失时会回退到 `localhost`。在容器或 release 部署场景下，这会让服务尝试连接容器自身的 MySQL，而不是生产 MySQL，导致容器反复重启且日志不直接指出缺失的部署配置。

## 范围

### In Scope

- 调整后端 MySQL 配置加载逻辑。
- 在 `GIN_MODE=release` 或容器运行时，要求显式提供 `DB_HOST`。
- 增加后端单元测试覆盖配置加载行为。
- 补充部署配置影响和验证路径。

### Out of Scope

- 不修改后端业务接口、handler、路由或数据库表结构。
- 不修改前端页面、路由、组件或状态管理。
- 不提交生产数据库密码或服务器脚本。

## 页面与交互影响

- 受影响页面: 无
- 受影响路由: 无
- 受影响布局或导航: 无
- 受影响组件: 无
- 受影响交互状态: 无

## 接口与数据影响

- 受影响接口: 无接口契约变化
- 请求参数或响应结构变化: 无
- 受影响服务层: 无
- 受影响后端 handler/service/repository: `backend/internal/database/mysql.go`
- 受影响数据库或缓存: 无 schema 变化
- 配置或环境变量变化: 生产和容器部署必须显式注入 `DB_HOST`，通常通过 `docker run -e DB_HOST=...` 或 `docker run --env-file .env`。
- 构建或部署影响: 后端容器缺失 `DB_HOST` 时会快速失败并提示配置缺失，不再静默连接容器内 `localhost`。

## 权限与合规影响

- 受影响角色: 无应用内角色影响
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 不影响应用内审核与风控；部署失败日志会更明确指出缺少 `DB_HOST`。

## 验收标准

1. 本地开发未设置 `GIN_MODE=release` 且非容器环境时，后端仍可默认使用 `localhost`。
2. `GIN_MODE=release` 且未设置 `DB_HOST` 时，后端返回明确的 `DB_HOST is required` 错误。
3. 容器运行且未设置 `DB_HOST` 时，后端返回明确的 `DB_HOST is required` 错误。
4. 显式设置 `DB_HOST` 后，后端按该主机生成 MySQL 连接配置。
5. `./harness/verify.sh` 通过。
6. README 的 Docker 构建示例与仓库根目录的 `Dockerfile.admin-*` 保持一致。
