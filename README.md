# **短链接服务**

> 基于 Go + Gin + GORM + Redis + MySQL 的高性能短链接生成与重定向服务

[![Go Version](https://img.shields.io/badge/Go-1.25-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## 📖 简介

短链接服务是一个将长 URL 转换为短 URL 的在线工具。用户提交长链接后，系统生成唯一的短码，访问短链接时会自动 302 重定向到原始长链接。同时提供点击统计、Redis 缓存加速、IP 限流、异步统计写入等功能。

**项目特点**：
- 🚀 **高性能**：单机 QPS 可达 4.4 万（16核64G 云服务器），P99 延迟 < 20ms
- 🧠 **智能缓存**：Cache-Aside 模式，热点短链自动缓存到 Redis
- 🛡️ **安全防护**：令牌桶限流、URL 格式校验、可达性重试
- 📊 **异步统计**：点击计数先记 Redis，定时批量回写 MySQL，降低数据库压力
- 🐳 **一键部署**：Docker Compose 快速启动完整服务栈
- 📝 **结构化日志**：JSON 格式日志，支持级别控制

---

## 🛠️ 技术栈

| 组件 | 技术 |
|------|------|
| 编程语言 | Go 1.25.4 |
| Web 框架 | Gin v1.12.0 |
| ORM | GORM v1.31.1 |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis 7.2 |
| 限流 | golang.org/x/time/rate |
| 日志 | log/slog |
| 部署 | Docker + Docker Compose |

---

## 🚀 快速开始

### 前置条件
- Go 1.25+（本地运行）
- Docker 和 Docker Compose（容器运行，推荐）

### 方式一：使用 Docker Compose（推荐）

```bash
# 1. 克隆仓库
git clone https://github.com/alac19/se-go-url-shortener-2026.git
cd se-go-url-shortener-2026

# 2. 启动所有服务（MySQL + Redis + 短链接服务）
docker-compose up -d

# 3. 测试生成短链
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'

# 预期返回：
# {"short_url":"http://localhost:8080/1"}
```

### 方式二：本地源码运行

```bash
# 1. 安装依赖
go mod download

# 2. 启动 MySQL 和 Redis（需提前安装）
# 推荐使用 Docker 快速启动：
docker run -d --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 -e MYSQL_DATABASE=shortlink_db mysql:8
docker run -d --name redis -p 6379:6379 redis:7.2

# 3. 修改 configs/config.toml 中的数据库地址（若使用 Docker，地址为 localhost）
# 4. 运行服务
go run cmd/server/main.go
```

---

## 📚 API 文档

### 1. 生成短链接

**请求**：
```http
POST /api/links
Content-Type: application/json

{
  "url": "https://example.com/very/long/url"
}
```

**响应**：
```json
{
  "short_url": "http://localhost:8080/abc123"
}
```

**错误码**：
- `400`：URL 格式无效或不可达
- `429`：请求频率超限（每分钟最多 5 次）
- `500`：服务器内部错误

### 2. 重定向

**请求**：
```http
GET /{code}
```

**响应**：`302 Found`，`Location` 头指向原始长链接。

---

## ⚙️ 配置说明

配置文件位于 `configs/config.toml`，主要参数：

| 配置段 | 字段 | 说明 |
|--------|------|------|
| `[mysql]` | `dsn` | MySQL 连接字符串 |
| `[redis]` | `addr` | Redis 地址 |
| `[server]` | `port`, `domain` | 服务端口和短链域名前缀 |
| `[ratelimit]` | `every_seconds`, `burst` | 令牌桶速率（秒）和桶容量 |
| `[asyncflush]` | `interval_seconds`, `scan_count` | 异步写入间隔和 SCAN 数量 |
| `[urlcheck]` | `timeout_seconds`, `max_retries` | URL 可达性检查超时和重试 |
| `[cache]` | `ttl_seconds` | Redis 缓存 TTL |
| `[log]` | `level`, `file_path` | 日志级别和输出文件 |

---

## 🧪 测试

```bash
# 运行所有单元测试
go test -v ./...

# 查看覆盖率
go test -coverprofile=c.out ./...
go tool cover -html=c.out
```

**压测结果**（16 核 64G 云服务器）：
- **QPS**：44,000+
- **P99 延迟**：< 20ms
- **错误率**：0%

---

## 📂 项目结构

```
.
├── cmd/server/          # 程序入口
├── internal/            # 内部模块
│   ├── config/          # 配置加载与校验
│   ├── handler/         # HTTP 处理层
│   ├── service/         # 业务逻辑层
│   ├── repository/      # 数据访问层（MySQL + Redis）
│   ├── middleware/      # 限流中间件
│   └── model/           # 数据模型
├── pkg/                 # 可复用工具
│   ├── base62/          # base62 编码
│   ├── limiter/         # 限流器
│   ├── logger/          # 日志初始化
│   └── urlcheck/        # URL 校验与可达性检查
├── configs/             # 配置文件
├── docs/                # 文档
├── docker-compose.yml
├── Dockerfile
└── README.md
```

---

## 🐳 Docker 部署

项目提供了完整的容器化方案：

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker logs shortlink-service

# 停止
docker-compose down
```

**容器说明**：
- `shortlink-mysql`：MySQL 8.0（数据持久化）
- `shortlink-redis`：Redis 7.2（数据持久化）
- `shortlink-service`：短链接服务（端口 8080）

---

## 📄 许可证

MIT © [刘灿阳](https://github.com/alac19)

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request。

---
