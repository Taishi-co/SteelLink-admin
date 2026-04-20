# AGENTS.md

本文件为 Codex、Claude Code 以及后续接入 agent 在 `SteelLink-admin` 中工作的统一入口。

当前仓库目标是 SteelLink 管理端前后端同仓项目，已知前提为“前端使用 React，后端使用 Go/Gin”。如果后续实际脚手架、构建工具或目录结构与本文不一致，应优先更新文档，不要让 agent 沿用过时假设。

## 1. 启动协议

进入仓库后按以下顺序建立上下文：

1. 阅读 `AGENTS.md`
2. 阅读 `docs/ai/entrypoint.md`
3. 阅读 `docs/ai/project-map.md`
4. 阅读 `docs/ai/engineering-rules.md`
5. 阅读 `docs/ai/git-workflow.md`
6. 如处理具体需求，再阅读对应 `docs/features/<feature-id>/`

如果以上文档与代码不一致，应先更新文档或在结果中明确指出差异，不要自行假设。

## 2. 标准工作流

所有变更必须按以下阶段推进：

1. 设计阶段：先补齐 `spec.md` 与 `test-cases.md`
2. 开发阶段：按 `implementation-plan.md` 实施
3. Review 阶段：执行 `./harness/review.sh`
4. 准备上线阶段：执行 `./harness/release.sh`
5. 部署阶段：统一通过 `./harness/deploy.sh <staging|prod>`

禁止直接在 `main` 上开发或绕过 PR 合并。

## 3. 必用命令入口

- `./harness/bootstrap.sh`
- `./harness/branch-guard.sh local`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/<feature-id>`
- `./harness/release.sh docs/features/<feature-id>`
- `./harness/deploy.sh staging`
- `./harness/deploy.sh prod`

## 4. 质量底线

- 前端部分如存在 `package.json`，必须可被 Node.js 工具链识别，并通过 `build`
- 前端如存在 `lint`、`typecheck`、`test:ci` 或 `test` 脚本，必须通过对应验证
- 后端部分如存在 `go.mod`，必须通过 `go test ./...`
- 后端如存在明确服务入口，必须完成 `go build`
- 如仓库存在 Docker Compose 配置，必须确认 `docker compose config` 可解析
- 涉及登录、菜单、路由守卫、角色权限、审批、风控时，必须核对 `docs/domain/role-permission-matrix.md`
- 涉及页面路由、组件结构、接口契约、状态管理、后端接口、数据库、环境变量、构建配置或部署改动时，必须在 feature 文档中写明影响范围
- 提交 PR 前必须确认当前分支来自最新 `main`
- 分支命名必须符合 `feat|fix|docs|chore|refactor|test|hotfix/<feature-id>-<slug>`
- 如果当前目录尚未初始化为 Git worktree 或尚未初始化前后端基础工程，先补齐仓库和项目基础结构，再执行 `branch-guard` 与 `verify`

## 5. 输出要求

提交结果时至少说明：

- 当前对应的 feature 文档路径
- 已执行的 harness 命令
- 是否完成前端 build / lint / typecheck / test 验证
- 是否完成后端 test / build 验证
- 是否存在遗留风险、人工验证点或部署前依赖
