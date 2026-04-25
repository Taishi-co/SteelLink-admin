# Backend Container MySQL Config Guard Implementation Plan

## Feature 信息

- Feature ID: SL-ADMIN-FIX-BACKEND-DB-CONTAINER
- 开发负责人: Codex

## 变更范围

- 代码文件: `backend/internal/database/mysql.go`、`backend/internal/database/mysql_test.go`
- 文档文件: `README.md`、`docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER/spec.md`、`docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER/test-cases.md`、`docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER/implementation-plan.md`、`docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER/release-checklist.md`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 新建 feature 文档，记录容器部署缺少 `DB_HOST` 的故障背景、影响范围与验收标准。
2. 将 MySQL 配置加载收口为可测试 helper。
3. 在 `GIN_MODE=release` 或容器运行时校验 `DB_HOST` 必须显式设置。
4. 增加单元测试覆盖本地默认值、release 缺失配置、容器缺失配置、显式生产配置。
5. 修正 README 中 Docker 独立部署示例，使用仓库根目录的 `Dockerfile.admin-*` 并强调容器必须注入 `DB_HOST`。
6. 执行 `branch-guard`、`verify`、`review` 与 `release` 检查。

## 验证计划

- `./harness/branch-guard.sh local`
- `cd backend && go test ./...`
- `cd backend && go build ./cmd/server`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER`
- `./harness/release.sh docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER`

## 风险与回滚关注点

- 风险: 生产部署脚本如果之前依赖默认 `localhost`，升级后会直接失败；这符合容器生产部署的显式配置要求。
- 风险: 部署脚本必须确认 `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME` 被注入后端容器。
- 回滚: 回退 `backend/internal/database/mysql.go` 中 release/container 的 `DB_HOST` 校验逻辑。
