# AI 生图后端服务

这是一个使用 Go 编写的 AI 生图后端服务，提供用户注册、登录、JWT 鉴权、图片生成、图片记录持久化和生成图片文件保存能力。

服务默认监听 `3000` 端口，数据库使用 MySQL，图片生成请求由后端统一代理转发到环境变量配置的生图服务，避免前端泄露 `APPID`、`BASE_URL` 和 `API_KEY`。

## 功能说明

- JWT 用户认证
- 手机号、用户账号、密码注册
- 用户账号或手机号登录
- 持久登录 Token 校验
- 生图任务提交和线程池执行
- 默认线程池数量为 `8`
- 生成图片 Base64 或 URL 结果保存为 PNG 文件
- MySQL 持久化用户信息和生成图片记录
- 结构化 JSON 日志输出
- 支持 CORS

## 项目结构

```text
backend
├── cmd/api/main.go                # 服务启动入口
├── internal/config                # 环境变量配置读取
├── internal/database              # MySQL 连接和表结构迁移
├── internal/handlers              # HTTP 接口处理
├── internal/middleware            # 鉴权、CORS、日志、中间件
├── internal/models                # 数据模型
├── internal/repositories          # 数据库访问层
├── internal/services              # 认证、生图、线程池逻辑
├── internal/utils                 # JWT、密码、响应工具
├── .env.example                   # 环境变量示例
├── API.md                         # 接口文档
└── 部署文档.md                    # 服务器部署说明
```

## MySQL 设计

数据库建议使用 `utf8mb4` 字符集：

```sql
CREATE DATABASE ai_image_demo
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

当前服务启动时会自动创建所需数据表。

### users 用户表

用于保存注册用户的基础信息和密码哈希。

```sql
CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  phone VARCHAR(32) NOT NULL UNIQUE,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT | 用户主键，自增 |
| phone | VARCHAR(32) | 手机号，唯一 |
| username | VARCHAR(64) | 用户账号，唯一 |
| password_hash | VARCHAR(255) | bcrypt 加密后的密码 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### generated_images 图片记录表

用于保存用户每次生图后的记录，包括提示词、图片路径、状态、耗时和失败原因。

```sql
CREATE TABLE IF NOT EXISTS generated_images (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  prompt TEXT NOT NULL,
  image_path VARCHAR(512) NOT NULL,
  status VARCHAR(32) NOT NULL,
  duration_ms BIGINT NOT NULL DEFAULT 0,
  error_message TEXT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_generated_images_user_id (user_id),
  CONSTRAINT fk_generated_images_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE
);
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT | 图片记录主键，自增 |
| user_id | BIGINT | 所属用户 ID |
| prompt | TEXT | 用户输入的提示词 |
| image_path | VARCHAR(512) | 生成图片保存路径 |
| status | VARCHAR(32) | 生成状态，当前为 `success` 或 `failed` |
| duration_ms | BIGINT | 生图耗时，单位毫秒 |
| error_message | TEXT | 失败原因，成功时为空 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

## 环境变量

复制 `.env.example` 为 `.env` 后修改配置：

```env
APP_PORT=3000

JWT_SECRET=please-change-this-secret
JWT_EXPIRE_HOURS=168

MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/ai_image_demo?charset=utf8mb4&parseTime=True&loc=Local

IMAGE_API_BASE_URL=https://www.codexapis.com
IMAGE_API_APP_ID=your-app-id
IMAGE_API_KEY=your-api-key
IMAGE_OUTPUT_DIR=./storage/images
WORKER_POOL_SIZE=8
```

配置说明：

| 配置项 | 说明 |
| --- | --- |
| APP_PORT | 后端服务端口，默认 `3000` |
| JWT_SECRET | JWT 签名密钥，生产环境必须替换 |
| JWT_EXPIRE_HOURS | Token 有效期，默认 `168` 小时 |
| MYSQL_DSN | MySQL 连接字符串 |
| IMAGE_API_BASE_URL | 生图服务基础地址 |
| IMAGE_API_APP_ID | 生图服务 APPID，可为空 |
| IMAGE_API_KEY | 生图服务 API Key |
| IMAGE_OUTPUT_DIR | 图片保存目录 |
| WORKER_POOL_SIZE | 生图线程池数量，默认 `8` |

## 本地运行

```bash
go mod tidy
go run ./cmd/api
```

服务启动后访问：

```text
http://127.0.0.1:3000/api/health
```

## 构建

```bash
go build -o ai-image-api ./cmd/api
```

## 相关文档

- 接口文档：[API.md](./API.md)
- 部署文档：[部署文档.md](./部署文档.md)
