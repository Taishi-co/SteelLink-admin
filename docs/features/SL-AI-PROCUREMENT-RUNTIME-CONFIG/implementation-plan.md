# SL-AI-PROCUREMENT-RUNTIME-CONFIG Implementation Plan

## 实施内容

1. 扩展 Admin AI采购配置结构，补齐 pgvector DSN 环境变量名、Postgres database、用户名、密码环境变量名、embedding provider、embedding endpoint 和 embedding API Key 环境变量名。
2. 扩展前端配置页对应表单字段。
3. 新增内部运行时配置接口，复用同一份持久化配置。
4. 在 `.env.example` 中记录可选共享同步 token。

## 数据与权限影响

- 复用 `admin_system_configs` 表，无新增表。
- 配置接口仍由超级管理员管理。
- 内部运行时接口在配置 `AI_PROCUREMENT_CONFIG_SYNC_TOKEN` 后需要内部 token。

## 回滚

- 回滚本 feature 代码即可。
- 已保存 JSON 中多出的字段不会影响旧版本读取。
