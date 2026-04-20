# Engineering Rules

## 开发原则

- 设计先于编码：没有 `spec.md` 与 `test-cases.md` 的需求，不应直接开始开发
- 小步提交：每次变更聚焦一个 feature 或一个子问题
- 影响可追溯：涉及权限、接口、数据库、状态、环境、构建、部署时，必须写入 feature 文档
- 统一入口：验证、评审、发布、部署统一走 `harness/`
- 分支规范：所有开发分支必须从最新 `main` 拉出，并通过 PR 合入 `main`
- 仓库切换时先校验当前目录是否为正确的 Git worktree，不要沿用旧 remote 或旧仓库假设

## 代码约束

### 通用规范

- 优先复用现有页面、路由、布局、组件、服务层、状态管理、handler、middleware、database 边界
- 权限改动必须同步核对 `docs/domain/role-permission-matrix.md`
- 涉及登录态、审批、风控、统计口径的改动，必须显式说明影响范围和回滚路径
- 涉及接口契约改动时，必须写明前后端联调影响、错误码与兼容策略
- 涉及数据库、缓存、JWT、权限、审核流的改动时，必须写明迁移顺序、影响面与回滚策略
- 涉及环境变量、构建、代理、部署改动时，必须先更新文档再改 CI 或脚本
- 不要在页面层硬编码生产环境域名、token、角色判断或后端错误兜底逻辑

### React / 前端规范

- 页面层负责组合，业务逻辑优先沉淀到 hooks、services、store 或通用组件
- 统一处理加载、空态、错误态、无权限态，不要只实现 happy path
- 表单、列表、筛选、分页、导出等管理端高频能力优先抽象复用
- 若项目使用 TypeScript，新增接口与状态必须补齐类型定义
- 若项目暂时是 JavaScript，新增代码也要保持清晰的模块边界与命名一致性
- 接口请求统一经过 `services/` 或请求封装层，不要在页面里直接散落 `fetch`/`axios` 调用

### Go / 后端规范

- 优先复用现有 Gin router、handler、middleware、service、repository 分层
- 错误响应、日志、鉴权与中间件链保持统一，不要在单个 handler 中临时各写一套
- 涉及数据库访问时，优先通过既有数据访问层收口，避免在 handler 中直接拼 SQL
- 配置与环境变量统一从配置层读取，不要把默认值散落在多个业务文件中

## 命令约束

```bash
./harness/bootstrap.sh
./harness/branch-guard.sh local
./harness/verify.sh
./harness/review.sh docs/features/<feature-id>
./harness/release.sh docs/features/<feature-id>
./harness/deploy.sh <staging|prod>
```

## 通过标准

1. 至少识别到前端或后端中的一个工程入口
2. 若存在前端 `package.json`，`build` 脚本执行成功
3. 若前端存在 `lint`、`typecheck`、`test:ci` 或 `test` 脚本，必须执行成功
4. 若存在后端 `go.mod`，`go test ./...` 执行成功
5. 若存在后端服务入口，`go build` 执行成功
6. 如环境安装了 Docker Compose 且仓库存在 Compose 文件，`docker compose config` 可解析
7. feature 文档四件套完整
8. 变更说明覆盖页面、路由、接口、数据库、权限、环境、回滚关注点
9. 分支名、PR base branch 与 merge 方式符合 `docs/ai/git-workflow.md`

## 禁止事项

- 禁止跳过 `verify` 直接发起 review
- 禁止直接向 `main` push 业务改动
- 禁止只在对话中描述测试结果而不运行命令
- 禁止把当前同仓项目误判为纯前端仓库或纯后端仓库
- 禁止把临时 mock、测试 token、调试开关直接带入发布路径
