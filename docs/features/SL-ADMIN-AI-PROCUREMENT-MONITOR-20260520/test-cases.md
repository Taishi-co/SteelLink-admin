# SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520 Test Cases

## 测试范围

- Feature ID: `SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520`
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 查看监控总览 | 存在 `ai_procurements` 数据 | 打开 `/ai-procurement/monitor` | 展示总量、今日新增、进行中、审核提交、异常和待介入统计 |
| TC-02 | 查看阶段分布 | AI 采购单处于多个阶段 | 打开监控页 | 阶段分布表展示各阶段数量 |
| TC-03 | 查看状态分布 | AI 采购单处于多个状态 | 打开监控页 | 状态分布表展示各状态数量 |
| TC-04 | 筛选待介入 | 存在审核提交、失败或停滞记录 | 选择“待介入”筛选 | 列表只展示待介入记录和原因 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | AI 采购表不存在 | 当前库未部署业务侧 AI 采购表 | 调用监控接口 | 返回空统计、空列表、`tableReady=false` |
| TC-EX-02 | 空数据 | 表存在但没有记录 | 打开监控页 | 页面展示 0 值统计和空列表 |
| TC-EX-03 | 非法分页参数 | `page` 或 `pageSize` 非数字 | 调用监控接口 | 使用默认分页，不报错 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 运营 | 查看 AI 采购监控 | 允许访问 |
| TC-AUTH-02 | 超管 | 查看 AI 采购监控 | 允许访问 |
| TC-AUTH-03 | 审核员 | 直接调用监控接口 | 返回 403 |
| TC-AUTH-04 | 普通管理端账号 | 直接调用监控接口 | 返回 403 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 监控辅助函数测试 | `go test ./...` | 待介入、JSON 数量和状态标签测试通过 |
| TC-BE-02 | 后端构建 | `go build ./cmd/server` | 构建通过 |

## 兼容与发布验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-FE-01 | 前端测试 | `pnpm test:ci` | 测试通过 |
| TC-OPS-01 | 前端构建 | `pnpm build` | 构建通过 |
| TC-OPS-02 | 发布门禁 | `./harness/release.sh docs/features/SL-ADMIN-AI-PROCUREMENT-MONITOR-20260520` | release gate 通过 |
