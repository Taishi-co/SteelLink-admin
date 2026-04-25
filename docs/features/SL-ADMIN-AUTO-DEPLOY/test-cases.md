# Production Auto Deploy Test Cases

## 测试范围

- Feature ID: SL-ADMIN-AUTO-DEPLOY
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | PR 合并触发生产部署 | PR 通过 review 后合并到 `main`，GitHub Secrets 已配置 | 观察 GitHub Actions `Deploy Production` workflow | workflow 通过 SSH 登录生产服务器，先执行 `./deploy_admin_backend.sh`，再执行 `./deploy_admin_frontend.sh` |
| TC-02 | 手动触发生产部署 | GitHub Secrets 已配置 | 在 GitHub Actions 页面手动运行 workflow | workflow 使用同一远程部署流程完成部署 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 缺少服务器认证信息 | `PROD_SSH_KEY` 与 `PROD_PASSWORD` 均未配置 | 运行 workflow | workflow 在 SSH 连接阶段失败，不会执行部署脚本 |
| TC-EX-02 | 服务器部署脚本失败 | 远程 `./deploy_admin_backend.sh` 或 `./deploy_admin_frontend.sh` 返回非 0 | 运行 workflow | workflow 失败，GitHub Actions 日志保留失败阶段 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | GitHub Actions deploy key 用户 | SSH 登录生产服务器执行部署 | 仅具备部署所需服务器权限，不需要应用内角色权限 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 后端工程校验 | `./harness/verify.sh` | 后端测试/构建通过，接口改动无阻断 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 项目构建可执行 | `./harness/verify.sh` | 前后端关键校验通过，无阻断 |
| TC-OPS-02 | Review gate | `./harness/review.sh docs/features/SL-ADMIN-AUTO-DEPLOY` | feature 文档完整且验证通过 |
| TC-OPS-03 | Release gate | `./harness/release.sh docs/features/SL-ADMIN-AUTO-DEPLOY` | 发布前检查通过 |
