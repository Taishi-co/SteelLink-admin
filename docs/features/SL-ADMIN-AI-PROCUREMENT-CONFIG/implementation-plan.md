# SL-ADMIN-AI-PROCUREMENT-CONFIG Implementation Plan

## 设计

- 后端沿用当前 `handler` + `database/sql` 风格。
- 配置以 JSON 写入 `admin_system_configs`，没有配置时返回默认值。
- 前端沿用 Ant Design、React Query 和现有布局菜单。

## 交付项

- 后端受保护配置读写接口。
- 前端导航、路由、API 类型和配置页面。
- 功能文档和验证清单。
