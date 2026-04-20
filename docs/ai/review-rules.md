# Review Rules

Review 不是只看代码风格，而是确认“实现是否符合设计、验证是否足够、上线风险是否可控”。

## Review 输入

发起 review 前必须提供：

- feature 文档目录
- `spec.md`
- `test-cases.md`
- `implementation-plan.md`
- `./harness/verify.sh` 的结果

## 阻断项

- 缺少 feature 文档
- 未识别到前端或后端工程入口
- 前端 `build` 未通过
- 已配置的前端 `lint`、`typecheck`、`test:ci` 或 `test` 未通过
- 后端 `go test ./...` 未通过
- 后端 `go build` 未通过
- 需求实现与 `spec.md` 不一致
- 测试用例未覆盖主要 happy path、异常 path、权限 path、关键兼容 path
- 涉及权限、审核或统计口径改动但未核对角色矩阵
- 涉及发布、环境变量、构建或代理改动但未更新 `release-checklist.md`

## Review 关注点

- 是否改动了错误的页面、组件、服务层、状态层边界
- 是否改动了错误的 handler、middleware、service、repository、config 边界
- 是否遗漏 loading、empty、error、permission denied 等关键界面状态
- 是否遗漏接口失败、重复提交、并发点击、分页筛选重置等管理端常见异常场景
- 是否引入菜单错乱、路由死链、白屏、越权可见等风险
- 是否引入接口越权、错误码漂移、配置失配或数据库副作用
- 是否保留足够的日志、埋点、错误提示与人工回滚信息
