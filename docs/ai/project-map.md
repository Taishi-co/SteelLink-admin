# Admin Project Map

## 当前仓库状态

`SteelLink-admin` 当前目录尚未初始化具体业务代码。本文件先定义 SteelLink 管理端前后端同仓项目的目标结构与协作边界；项目初始化后，如果实际结构与本文不同，应同步更新本文。

## 项目定位

`SteelLink-admin` 是 SteelLink 管理后台前后端同仓仓库。已知前提为前端使用 React、后端使用 Go/Gin，统一承载登录、权限、菜单导航、列表页、表单页、详情页、数据看板，以及管理端专属接口、鉴权、配置和部署能力。

## 技术栈约束

- 已知：前端框架使用 React
- 已知：后端框架使用 Go + Gin
- 待代码初始化后确认：TypeScript/JavaScript 方案、构建工具、路由库、状态管理、请求库、测试框架、样式方案
- 待代码初始化后确认：后端目录布局、模块拆分、数据库访问层、配置加载方式
- 原则：以仓库实际 `package.json`、锁文件、`go.mod` 和源码目录为准，不要在没有代码时武断指定实现细节

## 推荐目录地图

- `frontend/` 或 `web/`
  React 管理端主目录，若未拆子目录，也可能直接在仓库根目录放置前端代码。
- `frontend/src/main.*`
  前端应用启动入口，负责挂载根组件、注册路由、Provider 与全局样式。
- `frontend/src/router/`
  路由定义、路由守卫、动态菜单与权限拦截逻辑。
- `frontend/src/layouts/`
  管理端整体骨架，如侧边栏、顶部导航、面包屑、标签页、内容区。
- `frontend/src/pages/`
  页面级实现，按业务域划分目录。
- `frontend/src/components/`
  复用组件与通用管理端组件，如表格、搜索栏、弹窗、表单片段。
- `frontend/src/services/`
  请求封装、接口定义、DTO 转换、错误处理。
- `frontend/src/store/`
  全局状态、用户信息、权限、主题、字典缓存等。
- `frontend/src/hooks/`
  复用 hooks。
- `frontend/src/utils/`
  工具函数、格式化、权限判断、常量。
- `frontend/src/styles/`
  全局样式、主题变量、覆盖样式。
- `frontend/public/`
  静态资源与 HTML 模板。
- `backend/` 或 `server/`
  Go/Gin 服务主目录，若未拆子目录，也可能直接在仓库根目录放置后端代码。
- `backend/cmd/`
  后端服务启动入口。
- `backend/internal/`
  业务实现、handler、service、middleware、repository 等内部模块。
- `backend/pkg/`
  可复用公共能力，如响应结构、错误码、日志、认证封装。
- `backend/config/`
  配置结构、样例配置与环境变量映射。
- `harness/`
  项目约束脚本入口，负责 bootstrap、分支检查、验证、review、release、deploy。
- `docs/ai/`
  agent 协作规则与工作流文档。
- `docs/features/`
  feature 设计、测试、实施、发布文档；`_template/` 为模板。

## 关键模块

- 前端启动与 Provider 注册：`frontend/src/main.*` 或根目录同类入口
- 前端路由与权限守卫：`frontend/src/router/`
- 前端页面骨架与导航：`frontend/src/layouts/`
- 前端接口请求与鉴权续期：`frontend/src/services/`
- 前端用户信息与权限态：`frontend/src/store/`
- 后端 HTTP 启动与路由：`backend/cmd/`、`backend/internal/handler/`
- 后端 middleware、鉴权、日志与恢复：`backend/internal/middleware/` 或等价目录
- 后端数据库与配置：`backend/internal/database/`、`backend/config/`
- 工程化配置：`package.json`、`go.mod`、构建配置文件、环境变量文件

## 高风险改动区域

- 前端路由与权限控制
  任何菜单可见性、路由守卫、登录态恢复、token 刷新改动都可能造成越权或页面白屏。
- 前端请求层与接口契约
  请求拦截器、错误码处理、base URL、超时、重试与上传下载逻辑直接影响全站行为。
- 前端全局状态与布局骨架
  用户态、主题态、字典缓存、主布局改动会影响多个页面。
- 后端接口、权限与配置
  handler、middleware、认证、数据库连接、配置字段与默认值变更会直接影响接口可用性与安全边界。
- 环境变量与构建配置
  `.env*`、构建 base path、代理配置、静态资源路径一旦变更，容易导致线上资源加载或接口访问异常。
- 部署配置与静态资源产物
  `Dockerfile`、Nginx 配置、CDN 路径、容器入口脚本、二进制产物与 `dist/` 发布方式需要明确回滚路径。
- `harness/`
  是团队统一流程入口，修改时必须同步更新 `AGENTS.md` 与 `docs/ai/*.md`。

## 变更前检查

在以下场景必须先补文档再开发：

- 新增或调整页面、路由、布局、菜单、权限守卫
- 新增或修改前端接口请求、错误处理、上传下载、鉴权续期
- 新增或修改后端路由、handler、middleware、数据库访问、配置结构
- 调整全局状态、主题、字典缓存、埋点或日志
- 修改环境变量、构建配置、代理、Dockerfile、Nginx、部署流程
- 迁移 Git 工作流、仓库 remote、默认分支或 PR 规则
