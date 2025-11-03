# Hyperlane Backend API

基于 Go + Gin 的社区平台后端服务

## 🚀 快速启动

### 1. 环境要求
- Go 1.23+
- PostgreSQL 数据库

### 2. 配置数据库
编辑 `config.yaml` 配置文件：
```yaml
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your_password"
  dbname: "hyperlane"
  sslmode: "disable"
```

### 3. 启动服务
```bash
# 安装依赖
go mod download

# 运行项目
go run main.go

# 或编译后运行
go build -o hyperlane
./hyperlane
```

服务将在 `http://localhost:8080` 启动

---

## 📡 API 路由

### 🔐 认证
| Method | Endpoint | 说明 | 权限要求 |
|--------|----------|------|----------|
| POST | `/v1/login` | 用户登录 | - |

### 👤 用户管理
| Method | Endpoint | 说明 | 权限要求 |
|--------|----------|------|----------|
| PUT | `/v1/users/:id` | 更新用户信息 | JWT |
| GET | `/v1/users/:id` | 获取用户信息 | - |
| POST | `/v1/users/follow/:id` | 关注用户 | JWT |
| POST | `/v1/users/unfollow/:id` | 取消关注 | JWT |
| POST | `/v1/users/follow/states` | 批量获取关注状态 | JWT |

### 📅 活动管理
| Method | Endpoint | 说明 | 权限要求 |
|--------|----------|------|----------|
| POST | `/v1/events` | 创建活动 | event:write |
| DELETE | `/v1/events/:id` | 删除活动 | event:delete |
| PUT | `/v1/events/:id` | 更新活动 | event:write |
| GET | `/v1/events` | 查询活动列表 | - |
| GET | `/v1/events/:id` | 获取活动详情 | - |
| PUT | `/v1/events/:id/status` | 更新发布状态 | event:review |
| POST | `/v1/events/recap` | 创建活动回顾 | blog:write |
| DELETE | `/v1/events/recap/:id` | 删除回顾 | blog:delete |
| PUT | `/v1/events/recap/:id` | 更新回顾 | blog:write |
| GET | `/v1/events/recap` | 获取回顾 | - |

### 📝 博客管理
| Method | Endpoint | 说明 | 权限要求 |
|--------|----------|------|----------|
| POST | `/v1/blogs` | 创建博客 | blog:write |
| DELETE | `/v1/blogs/:id` | 删除博客 | blog:delete |
| PUT | `/v1/blogs/:id` | 更新博客 | blog:write |
| GET | `/v1/blogs/:id` | 获取博客详情 | - |
| GET | `/v1/blogs` | 查询博客列表 | - |
| PUT | `/v1/blogs/:id/status` | 更新发布状态 | blog:review |

### 📚 教程管理
| Method | Endpoint | 说明 | 权限要求 |
|--------|----------|------|----------|
| POST | `/v1/tutorials` | 创建教程 | tutorial:write |
| DELETE | `/v1/tutorials/:id` | 删除教程 | tutorial:delete |
| PUT | `/v1/tutorials/:id` | 更新教程 | tutorial:write |
| GET | `/v1/tutorials/:id` | 获取教程详情 | - |
| GET | `/v1/tutorials` | 查询教程列表 | - |
| PUT | `/v1/tutorials/:id/status` | 更新发布状态 | tutorial:review |

### 💬 帖子管理
| Method | Endpoint | 说明 | 权限要求 |
|--------|----------|------|----------|
| POST | `/v1/posts` | 创建帖子 | blog:write |
| DELETE | `/v1/posts/:id` | 删除帖子 | blog:delete |
| GET | `/v1/posts/:id` | 获取帖子详情 | - |
| PUT | `/v1/posts/:id` | 更新帖子 | blog:write |
| GET | `/v1/posts` | 查询帖子列表 | - |
| GET | `/v1/posts/stats` | 帖子统计 | - |
| POST | `/v1/posts/:id/like` | 点赞 | JWT |
| POST | `/v1/posts/:id/unlike` | 取消点赞 | JWT |
| POST | `/v1/posts/:id/favorite` | 收藏 | JWT |
| POST | `/v1/posts/:id/unfavorite` | 取消收藏 | JWT |
| GET | `/v1/posts/status` | 获取帖子状态 | JWT |

### 💡 反馈管理
| Method | Endpoint | 说明 | 权限要求 |
|--------|----------|------|----------|
| POST | `/v1/feedbacks` | 提交反馈 | JWT |
| GET | `/v1/feedbacks` | 查询反馈列表 | - |

### 📊 数据统计
| Method | Endpoint | 说明 | 权限要求 |
|--------|----------|------|----------|
| GET | `/v1/stats` | 获取统计概览 | - |

---

## 🔑 权限说明

需要在请求头中携带 JWT Token：
```
Authorization: Bearer <your_token>
```

权限类型：
- `JWT` - 只需登录
- `blog:write` - 博客写权限
- `blog:delete` - 博客删除权限
- `blog:review` - 博客审核权限
- `event:write` - 活动写权限
- `event:delete` - 活动删除权限
- `event:review` - 活动审核权限
- `tutorial:write` - 教程写权限
- `tutorial:delete` - 教程删除权限
- `tutorial:review` - 教程审核权限

---

## 🛠️ 技术栈

- **框架**: Gin
- **数据库**: PostgreSQL + GORM
- **认证**: JWT
- **日志**: Logrus
- **配置**: Viper
- **限流**: uber/ratelimit

---

## 📁 项目结构

```
hyperlane/
├── config/          # 配置模块
├── controllers/     # 控制器（业务逻辑）
├── middlewares/     # 中间件（CORS、JWT、日志、限流）
├── models/          # 数据模型（GORM）
├── routes/          # 路由定义
├── logger/          # 日志系统
├── utils/           # 工具函数
├── config.yaml      # 配置文件
└── main.go          # 入口文件