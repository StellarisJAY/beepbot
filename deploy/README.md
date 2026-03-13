# BeepBot 部署指南

## 快速部署

### 1. 环境准备

确保已安装：
- Docker
- Docker Compose

### 2. 配置

复制并编辑配置文件（如需修改默认配置）：

```bash
# 配置文件已提供默认值，可直接使用
# 如需修改数据库密码等配置，编辑以下文件：
# - config.json: 后端服务配置
# - .env: 环境变量（可选）
```

创建 `.env` 文件来自定义配置（可选）：

```bash
# .env 文件示例
POSTGRES_DB=beepbot
POSTGRES_USER=beepbot
POSTGRES_PASSWORD=your_secure_password
API_PORT=8888
DASHBOARD_PORT=80
```

### 3. 启动服务

```bash
cd deploy
docker compose up -d
```

### 4. 访问应用

- **前端 Dashboard**: http://localhost (默认端口 80)
- **后端 API**: http://localhost:8888

### 5. 默认管理员账号

首次启动会创建默认管理员账号：
- 用户名: `admin`
- 密码: `admin123`

**请登录后立即修改密码！**

## 常用命令

```bash
# 查看日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f backend
docker compose logs -f dashboard

# 停止服务
docker compose down

# 停止并删除数据卷（重置）
docker compose down -v

# 重新构建镜像
docker compose build --no-cache

# 更新服务
docker compose pull
docker compose up -d
```

## 数据持久化

Docker Compose 配置了以下数据卷：

| 卷名 | 用途 |
|------|------|
| `beepbot_postgres_data` | PostgreSQL 数据库数据 |
| `beepbot_data` | 应用数据（技能、工作空间） |
| `beepbot_keys` | 加密密钥、JWT 密钥 |

## 配置说明

### config.json 配置项

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `port` | API 服务端口 | 8888 |
| `beepbot_data_dir` | 数据目录 | /data/beepbot |
| `database.host` | 数据库主机 | database |
| `database.port` | 数据库端口 | 5432 |
| `database.user` | 数据库用户 | beepbot |
| `database.password` | 数据库密码 | beepbot123 |
| `database.dbname` | 数据库名称 | beepbot |

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `POSTGRES_DB` | PostgreSQL 数据库名 | beepbot |
| `POSTGRES_USER` | PostgreSQL 用户名 | beepbot |
| `POSTGRES_PASSWORD` | PostgreSQL 密码 | beepbot123 |
| `API_PORT` | API 服务映射端口 | 8888 |
| `DASHBOARD_PORT` | Dashboard 映射端口 | 80 |

## 生产环境建议

1. **修改默认密码**: 修改 `.env` 中的 `POSTGRES_PASSWORD`
2. **配置 HTTPS**: 使用反向代理（如 Nginx、Caddy）配置 SSL
3. **定期备份**: 备份数据库和 `beepbot_data` 卷
4. **日志管理**: 配置日志轮转或使用日志收集服务

## 故障排查

### 后端无法连接数据库

```bash
# 检查数据库状态
docker compose logs database

# 确保数据库已启动
docker compose ps
```

### 前端无法访问 API

```bash
# 检查后端健康状态
curl http://localhost:8888/api/v1/health

# 检查网络连接
docker compose exec dashboard ping backend
```

### 重置服务

```bash
# 警告：这将删除所有数据！
docker compose down -v
docker compose up -d
```