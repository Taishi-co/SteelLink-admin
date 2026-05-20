# Implementation Plan

## Feature 信息

- Feature ID: `SL-ADMIN-WORKBENCH-FLOW-BASELINE`
- 开发负责人: Codex

## 变更范围

- 代码文件: 无
- 文档文件:
  - `docs/ai/project-map.md`
  - `docs/features/SL-ADMIN-WORKBENCH-FLOW-BASELINE/*`
- 配置文件: 无
- 部署/脚本文件:
  - `harness/deploy.sh`

## 实施步骤

1. 用真实目录结构更新管理端项目地图。
2. 让部署 harness 在 `DEPLOY_COMMAND` 之后优先识别 admin 双 Dockerfile。
3. 在 dry-run 中输出明确的后端/前端构建命令，并保留 `DEPLOY_EXECUTE=1` 执行开关。
4. 执行 harness 语法校验、deploy dry-run 和管理端 verify。

## 验证计划

- `bash -n harness/deploy.sh`
- `./harness/deploy.sh staging`
- `ADMIN_FRONTEND_API_BASE_URL=http://127.0.0.1:8080/api/v1 ./harness/deploy.sh staging`
- `./harness/verify.sh`
- 需要的人工验证: 发布前确认镜像 tag、API base URL、容器运行命令和目标环境。

## 风险与回滚关注点

- 风险: 本轮只规划镜像构建，不自动运行容器或推送镜像，真实发布仍需环境侧命令补齐。
- 回滚: 回退 `harness/deploy.sh` 和 `docs/ai/project-map.md`。
