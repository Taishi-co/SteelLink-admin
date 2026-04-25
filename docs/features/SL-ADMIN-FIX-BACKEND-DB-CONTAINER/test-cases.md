# Backend Container MySQL Config Guard Test Cases

## 测试范围

- Feature ID: SL-ADMIN-FIX-BACKEND-DB-CONTAINER
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 本地开发默认连接 localhost | 未设置 `GIN_MODE=release`，非容器运行，未设置 `DB_HOST` | 执行配置加载单元测试 | 配置 host 为 `localhost` |
| TC-02 | 生产显式指定 DB_HOST | `GIN_MODE=release`，设置 `DB_HOST=172.16.0.6` | 执行配置加载单元测试 | 配置 host 为 `172.16.0.6` |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | release 缺少 DB_HOST | `GIN_MODE=release`，未设置 `DB_HOST` | 执行配置加载单元测试 | 返回 `DB_HOST is required` 错误 |
| TC-EX-02 | 容器缺少 DB_HOST | 容器运行，未设置 `DB_HOST` | 执行配置加载单元测试 | 返回 `DB_HOST is required` 错误 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 任意应用角色 | 数据库配置校验 | 不改变应用内权限判断 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 后端单元测试 | `cd backend && go test ./...` | 配置加载测试通过 |
| TC-BE-02 | 后端构建 | `cd backend && go build ./cmd/server` | 服务入口构建通过 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 项目验证 | `./harness/verify.sh` | 前后端关键校验通过 |
| TC-OPS-02 | Review gate | `./harness/review.sh docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER` | feature 文档完整且验证通过 |
| TC-OPS-03 | Release gate | `./harness/release.sh docs/features/SL-ADMIN-FIX-BACKEND-DB-CONTAINER` | 发布前检查通过 |
| TC-OPS-04 | README 部署示例 | 人工检查 `README.md` 生产部署章节 | Docker build 使用 `Dockerfile.admin-backend` 和 `Dockerfile.admin-frontend` |
