# Test Cases

## 测试范围

- Feature ID: `SL-ADMIN-WORKBENCH-FLOW-BASELINE`
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 管理端部署 dry-run | 仓库存在两个 admin Dockerfile | 执行 `./harness/deploy.sh staging` | 输出 backend/frontend 两个镜像构建计划 |
| TC-02 | 前端 API base build arg | 设置 `ADMIN_FRONTEND_API_BASE_URL` | 执行 `./harness/deploy.sh staging` | 前端构建计划包含 `--build-arg VITE_API_BASE_URL=...` |
| TC-03 | 自定义部署命令优先 | 设置 `DEPLOY_COMMAND` | 执行 `./harness/deploy.sh staging` | 输出自定义部署入口，不走 Dockerfile 自动分支 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 非法环境 | 无 | 执行 `./harness/deploy.sh dev` | 输出 usage 并失败 |
| TC-EX-02 | 无部署入口 | 移除 Dockerfile/compose 且未设置 `DEPLOY_COMMAND` | 执行部署脚本 | dry-run 提示未检测到部署入口 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 发布执行者 | dry-run 部署 | 不需要应用权限或数据库权限 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 管理端后端无业务改动 | `./harness/verify.sh` | 后端测试/构建通过或暴露既有阻断 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | harness 语法 | `bash -n harness/deploy.sh` | shell syntax 通过 |
| TC-OPS-02 | 发布 gate | `./harness/release.sh docs/features/SL-ADMIN-WORKBENCH-FLOW-BASELINE` | 必需文档存在，verify 无新增阻断 |
