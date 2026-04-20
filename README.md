# SteelLink Admin

SteelLink B2B 钢铁商城运营管理后台，用于支撑日常运营工作，包括用户管理、企业审核、商品审核等功能。

## 技术栈

### 前端
- React 18 + TypeScript
- Vite 6
- Ant Design 5.x
- Zustand (状态管理)
- React Query (服务端状态)
- React Router DOM 7

### 后端
- Go 1.23+
- Gin Web Framework
- JWT 认证
- MySQL 8.0

## 目录结构

```
SteelLink-admin/
├── frontend/                 # React 前端
│   ├── src/
│   │   ├── main.tsx         # 入口文件
│   │   ├── App.tsx          # 根组件
│   │   ├── router/          # 路由配置
│   │   ├── layouts/         # 布局组件
│   │   ├── pages/           # 页面
│   │   ├── services/        # API 服务
│   │   ├── store/           # 状态管理
│   │   └── styles/          # 样式
│   ├── Dockerfile
│   └── nginx.conf
│
├── backend/                  # Go 后端
│   ├── cmd/server/          # 服务入口
│   ├── internal/
│   │   ├── handler/         # HTTP handlers
│   │   ├── middleware/      # 中间件
│   │   └── router/          # 路由
│   ├── config/              # 配置
│   └── Dockerfile
│
├── scripts/
│   └── init-db.sql          # 数据库初始化
│
├── docker-compose.prod.yml  # 生产部署配置
└── Makefile                 # 常用命令
```

## 功能模块

### 系统管理
- 管理员管理 (CRUD)
- 角色管理 (CRUD)
- 菜单管理 (CRUD)

### 业务管理
- 用户管理 (查看、状态控制)
- 企业管理 (入驻审核、资质审核)
- 商品管理 (上架审核、分类管理)

## 快速开始

### 环境要求
- Node.js 18+
- Go 1.21+
- MySQL 8.0
- pnpm (推荐)

### 本地开发

**1. 克隆项目**
```bash
git clone https://github.com/Jijiang-co/SteelLink-admin.git
cd SteelLink-admin
```

**2. 初始化数据库**
```bash
mysql -h <host> -u root -p steellink < scripts/init-db.sql
```

**3. 配置环境变量**
```bash
# 后端配置
cp .env.example .env
# 编辑 .env 填入数据库连接信息

# 前端配置
cp frontend/.env.example frontend/.env
```

**4. 启动后端**
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

**5. 启动前端**
```bash
cd frontend
pnpm install
pnpm dev
```

访问 http://localhost:3000

## 生产部署

### Docker 部署（推荐）

**1. 创建环境变量文件**
```bash
cp .env.production .env
# 编辑 .env，配置数据库连接和 JWT 密钥
```

**2. 构建并启动**
```bash
make deploy
```

或使用 docker compose：
```bash
docker compose -f docker-compose.prod.yml up -d --build
```

**3. 查看日志**
```bash
make logs
```

**4. 访问地址**
- 管理平台前端：`http://服务器IP:6666`
- 后端 API：`http://服务器IP:6667`

### 常用命令

| 命令 | 说明 |
|------|------|
| `make build` | 构建 Docker 镜像 |
| `make up` | 启动服务 |
| `make down` | 停止服务 |
| `make logs` | 查看日志 |
| `make deploy` | 构建并启动 |

### 手动部署

**前端构建**
```bash
cd frontend
pnpm build
# 将 dist/ 目录部署到 Nginx
```

**后端构建**
```bash
cd backend
go build -o server ./cmd/server
./server
```

## 数据库设计

### 标准字段约定

所有业务表统一包含以下字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | BIGINT UNSIGNED | 自增主键 |
| `created_by` | BIGINT UNSIGNED | 创建者 ID |
| `created_at` | DATETIME | 创建时间 |
| `updated_by` | BIGINT UNSIGNED | 修改者 ID |
| `updated_at` | DATETIME | 修改时间 |
| `deleted_at` | DATETIME | 软删除时间 |

### 主要数据表

| 表名 | 说明 |
|------|------|
| `admin_users` | 管理员账号 |
| `admin_roles` | 角色 |
| `admin_menus` | 菜单配置 |
| `admin_role_menus` | 角色-菜单关联 |
| `admin_operation_logs` | 操作日志 |
| `users` | 用户账号（共享） |
| `enterprises` | 企业信息（共享） |
| `products` | 商品信息（共享） |
| `product_categories` | 商品分类（共享） |

## API 接口

### 公开接口
- `POST /api/v1/login` - 管理员登录

### 认证接口（需要 JWT Token）
- `GET /health` - 健康检查
- `GET/POST/PUT/DELETE /api/v1/admins` - 管理员管理
- `GET/POST/PUT/DELETE /api/v1/roles` - 角色管理
- `GET/POST/PUT/DELETE /api/v1/menus` - 菜单管理
- `GET /api/v1/users` - 用户列表
- `GET /api/v1/enterprises` - 企业列表
- `GET /api/v1/products` - 商品列表

## 默认账号

首次部署后，系统会创建默认管理员账号：

- 用户名：`admin`
- 密码：`admin123`（生产环境请立即修改）

## 开发规范

详见 [AGENTS.md](./AGENTS.md) 和 [docs/ai/](./docs/ai/) 目录。

## License

MIT
