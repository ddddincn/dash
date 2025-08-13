# Dash 博客系统

一个基于 Go + Gin 框架开发的现代化博客系统，支持文章管理、分类标签、主题配置等功能。

## ✨ 功能特性

- 📝 **文章管理**：支持文章的创建、编辑、发布、删除等完整生命周期管理
- 🏷️ **分类标签**：灵活的分类和标签系统，便于内容组织
- 🎨 **主题系统**：支持自定义主题配置
- 📊 **统计面板**：提供文章、分类、标签等数据统计
- 🔐 **用户认证**：基于 JWT 的安全认证系统
- 🌐 **RESTful API**：完整的 REST API 接口
- 📱 **响应式前端**：现代化的管理界面
- 🐳 **容器化部署**：支持 Docker 容器化部署
- 💾 **多数据库支持**：支持 MySQL 和 SQLite3
- 🚀 **高性能缓存**：集成 Redis 缓存系统

## 🛠️ 技术栈

### 后端
- **框架**：Gin (Go Web 框架)
- **数据库 ORM**：GORM v2
- **数据库**：MySQL / SQLite3
- **缓存**：Redis
- **认证**：JWT (golang-jwt/jwt/v5)
- **配置管理**：Viper
- **日志**：Zap + Lumberjack
- **依赖注入**：Google Wire
- **代码生成**：GORM Gen

### 前端
- **框架**：React 18
- **路由**：React Router DOM
- **UI 组件库**：Ant Design

## 📋 系统要求

- Go 1.24+ 
- MySQL 5.7+ 或 SQLite3
- Redis 6.0+
- Docker (可选，用于容器化部署)

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd dash
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

#### 使用 MySQL
确保 MySQL 服务正在运行，并创建数据库：

```sql
CREATE DATABASE dash_dev_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

#### 使用 SQLite3 (开发环境推荐)
无需额外配置，程序会自动创建 `dash.db` 文件。

### 4. 配置 Redis

确保 Redis 服务正在运行：

```bash
# Linux/macOS
redis-server

# Windows (使用 Redis for Windows)
redis-server.exe
```

### 5. 首次安装

首次运行时，系统会启动安装向导：

```bash
go run main.go
```

访问 `http://localhost:8080` 进入安装界面，按照提示完成以下配置：

- **数据库配置**：选择数据库类型并填写连接信息
- **Redis 配置**：填写 Redis 连接信息
- **管理员账户**：设置管理员用户名、密码、昵称和邮箱
- **站点信息**：设置站点标题和 URL

安装完成后，系统会自动重启并进入正常运行模式。

### 6. 正常启动

```bash
go run main.go
```

服务器启动后，访问：
- **前端界面**：http://localhost:8080
- **管理后台**：http://localhost:8080/admin
- **API 接口**：http://localhost:8080/api

## ⚙️ 配置说明

### 主配置文件 `conf/config.yaml`

```yaml
server:
  host: 0.0.0.0      # 服务监听地址
  port: "8080"       # 服务端口

logging:
  filename: dash.log  # 日志文件名
  level:
    app: info        # 应用日志级别
    gorm: warn       # GORM 日志级别
  maxsize: 10        # 日志文件最大大小 (MB)
  maxage: 30         # 日志保留天数
  compress: false    # 是否压缩日志

sqlite3:
  enable: false      # 是否启用 SQLite3
  filename: dash.db  # SQLite3 数据库文件

mysql:
  dsn: root:123456@tcp(localhost:3306)/dash_dev_db?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true

cache:
  redis:
    addr: localhost:6379  # Redis 地址
    password: ""          # Redis 密码
    db: 0                # Redis 数据库编号
  default_ttl: 5m        # 默认缓存过期时间

dash:
  log_mode: console      # 日志模式: console/file
  mode: production       # 运行模式: development/production
  work_dir: ./          # 工作目录
  log_dir: ./logs       # 日志目录
```

### 安装配置文件 `conf/install.yaml`

安装完成后会自动生成，包含安装状态和配置信息。

## 🐳 Docker 部署

### 构建镜像

```bash
docker build -t dash-blog .
```

### 运行容器

```bash
docker run -d \
  --name dash-blog \
  -p 8080:8080 \
  -v $(pwd)/conf:/app/conf \
  -v $(pwd)/logs:/app/logs \
  dash-blog
```

## 📚 API 文档

### 公开接口

#### 文章相关
- `GET /api/posts` - 获取文章列表
- `GET /api/posts/:slug` - 根据 slug 获取文章详情
- `GET /api/posts/search` - 搜索文章
- `GET /api/posts/archive` - 获取文章归档

#### 分类相关
- `GET /api/categories` - 获取分类列表
- `GET /api/categories/:slug/posts` - 获取分类下的文章

#### 标签相关
- `GET /api/tags` - 获取标签列表
- `GET /api/tags/:slug/posts` - 获取标签下的文章

#### 其他
- `GET /api/menus` - 获取菜单列表
- `GET /api/theme/:themeID` - 获取主题设置
- `GET /ping` - 健康检查

### 管理接口 (需要认证)

#### 认证
- `POST /api/admin/auth/login` - 管理员登录
- `POST /api/admin/auth/refresh` - 刷新令牌

#### 文章管理
- `GET /api/admin/posts` - 获取文章列表
- `POST /api/admin/posts` - 创建文章
- `PUT /api/admin/posts/:id` - 更新文章
- `DELETE /api/admin/posts/:id` - 删除文章
- `PATCH /api/admin/posts/:id/status/:status` - 更新文章状态

#### 分类管理
- `GET /api/admin/categories` - 获取分类列表
- `POST /api/admin/categories` - 创建分类
- `PUT /api/admin/categories/:id` - 更新分类
- `DELETE /api/admin/categories/:id` - 删除分类

#### 标签管理
- `GET /api/admin/tags` - 获取标签列表
- `POST /api/admin/tags` - 创建标签
- `PUT /api/admin/tags/:id` - 更新标签
- `DELETE /api/admin/tags/:id` - 删除标签

#### 统计信息
- `GET /api/admin/statistics` - 获取统计数据

## 🏗️ 开发指南

### 项目结构

```
dash/
├── cmd/              # 命令行工具
│   └── generate/     # 代码生成工具
├── conf/             # 配置文件
├── config/           # 配置模块
├── consts/           # 常量定义
├── controller/       # 控制器层
│   ├── handler/      # 业务处理器
│   └── middleware/   # 中间件
├── dal/              # 数据访问层 (自动生成)
├── log/              # 日志模块
├── model/            # 数据模型
│   ├── dto/          # 数据传输对象
│   ├── entity/       # 数据库实体 (自动生成)
│   ├── param/        # 参数模型
│   ├── property/     # 属性配置
│   └── vo/           # 视图对象
├── service/          # 业务逻辑层
│   ├── assembler/    # 数据装配器
│   └── impl/         # 业务实现
├── utils/            # 工具函数
├── cache/            # 缓存模块
├── injection/        # 依赖注入 (Wire)
└── resource/         # 静态资源
```

### 代码生成

项目使用 GORM Gen 进行代码生成，当数据库表结构发生变化时，需要重新生成代码：

```bash
go run cmd/generate/generate.go
```

### 依赖注入

项目使用 Google Wire 进行依赖注入，修改依赖关系后需要重新生成：

```bash
go generate ./injection/
```

### 开发模式

在开发模式下，修改配置文件中的 `dash.mode` 为 `development`：

```yaml
dash:
  mode: development
  log_mode: console
```

## ✅ TODO

- [ ] 优化登录认证策略
- [ ] 增强主题自定义能力
- [ ] 优化响应格式
