# Test Cases

## 测试范围

- Feature ID: `SL-ADMIN-FIX-BUILD`
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 前端 lint 通过 | 已安装前端依赖 | 执行 `./harness/verify.sh` | 前端 lint 阶段通过 |
| TC-02 | 前端构建通过 | 已安装前端依赖 | 在 `frontend/` 执行 `pnpm build` | TypeScript 编译与 Vite 打包成功 |
| TC-03 | Review gate 可执行 | feature 文档完整 | 执行 `./harness/review.sh docs/features/SL-ADMIN-FIX-BUILD` | review gate 通过 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 后端本地工具链与 `go.mod` 版本不一致 | 本地 Go 版本低于 `go.mod` 要求 | 执行 `./harness/verify.sh` | 明确暴露后端工具链阻断，便于后续单独处理 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 任意管理员角色 | 打包修复后访问原有菜单与页面 | 不因本次改动引入权限可见性变化 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 后端校验不受前端修复影响 | `./harness/verify.sh` | 后端阶段若失败，应仅为独立环境或既有问题，不是本次前端修复引入 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 前端构建链路恢复 | `pnpm build` | 产物可正常生成 |
| TC-OPS-02 | Release gate 可执行 | `./harness/release.sh docs/features/SL-ADMIN-FIX-BUILD` | release gate 通过或明确暴露独立后端环境阻断 |
