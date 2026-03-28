# YT Video Downloader

这是一个基于 `yt-dlp` 的视频下载网站 monorepo，当前采用以下技术方案：

- 前端：`Vite + React`
- 后端：`Go API + Go Worker`
- 数据库：`PostgreSQL`
- 队列/缓存：`Redis`
- 对象存储：`Cloudflare R2`
- 部署：`Docker Compose`
- 运行环境：`VPS + Cloudflare proxy`

## 目录结构

```text
apps/
  api/        Go API
  worker/     Go Worker
  web/        Vite + React
db/
  init/       数据库初始化 SQL
deploy/
  caddy/      Caddy 配置
  compose/    Docker Compose
  docker/     各服务 Dockerfile
internal/     Go 共享代码
packages/     前端共享包
```

## 本地启动

1. 复制环境变量模板：

```bash
cp .env.example .env
```

2. 启动基础设施和应用：

```bash
docker compose -f deploy/compose/docker-compose.yml up --build
```

3. 访问：

- Web: `http://localhost`
- API 健康检查: `http://localhost/api/v1/healthz`

## 当前状态

当前仓库已完成基础工程骨架：

- Monorepo 结构
- Web/API/Worker 最小可运行入口
- Postgres/Redis/R2 配置约定
- Docker Compose 和容器镜像定义

下一步建议：

- 接入真实的 Postgres / Redis 客户端
- 实现 `yt-dlp` probe 和下载任务
- 接入 R2 上传
- 增加登录与限流

