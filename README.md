# Archive

`Archive` 是一个面向 YouTube 的下载产品 monorepo。

当前版本已经包含完整的前后端主流程：

- 匿名探测 YouTube 链接
- 登录后创建下载任务
- 支持 `MP4 360p / 480p / 720p / 1080p`
- 支持 `仅音频 MP3`
- 下载历史按账号隔离
- 下载详情页、支持页、移动端导航
- Go API + Go Worker 异步处理下载和上传

## 技术栈

- 前端：`Vite + React + Tailwind CSS + React Router + Clerk`
- 后端：`Go`
- 数据库：`PostgreSQL`
- 队列/缓存：`Redis`
- 对象存储：S3 兼容存储，当前环境变量以 `R2_*` 命名
- 部署：`Docker Compose + Caddy`

## 目录结构

```text
apps/
  api/        Go API
  worker/     Go Worker
  web/        React 前端
db/
  init/       数据库初始化 SQL
deploy/
  caddy/      Caddy 配置
  compose/    Docker Compose
  docker/     各服务 Dockerfile
  nginx/      前端静态资源与运行时配置注入
internal/     Go 共享代码
packages/     前端共享类型
```

## 当前产品行为

- 仅支持 `youtube.com/watch` 和 `youtu.be`
- `POST /api/v1/videos/probe` 可匿名访问
- 创建下载、查看历史、查看任务详情、获取结果链接都必须登录
- 登录后当前版本不做每日下载次数限制
- 下载历史、状态、结果链接都只属于创建该任务的 Clerk 用户

## 环境变量

先复制模板：

```bash
cp .env.example .env
```

关键变量：

- `VITE_CLERK_PUBLISHABLE_KEY`：前端 Clerk publishable key
- `CLERK_SECRET_KEY`：后端校验 Clerk session token
- `DATABASE_URL`：Postgres 连接串
- `REDIS_ADDR`：Redis 地址
- `R2_*`：对象存储配置
- `DOWNLOAD_ROOT`：worker 临时工作目录

## 本地启动

```bash
docker compose -f deploy/compose/docker-compose.yml up -d --build
```

启动后访问：

- Web：`http://localhost`
- 健康检查：`http://localhost/api/v1/healthz`
- 支持页：`http://localhost/support`

## 前端页面

- `/`：首页，包含 URL 输入、可用规格探测、最近下载
- `/downloads/:jobId`：下载详情页，显示任务状态和结果链接
- `/support`：FAQ / About / 产品说明页

## API 概览

- `POST /api/v1/videos/probe`
  - 请求：`{ "url": "https://www.youtube.com/watch?v=..." }`
  - 返回：视频信息和可用 `profiles[]`
- `POST /api/v1/downloads`
  - 需登录
  - 请求：`{ "url": "...", "profileId": "video-720p" }`
- `GET /api/v1/downloads`
  - 需登录
  - 返回当前用户自己的下载历史
- `GET /api/v1/downloads/:jobId`
  - 需登录
  - 返回任务状态
- `GET /api/v1/downloads/:jobId/result`
  - 需登录
  - 任务完成后返回结果下载链接

## 下载规格

- `video-360p`
- `video-480p`
- `video-720p`
- `video-1080p`
- `audio-mp3`

探测接口只会返回当前视频真正可用的规格；前端会展示可选项，后端也会再次校验。

## 开发检查

前端：

```bash
npm run lint --workspace @ytvd/web
npm run build --workspace @ytvd/web
```

后端：

```bash
go test ./...
```
