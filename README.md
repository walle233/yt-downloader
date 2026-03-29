# YTDownloader

`YTDownloader` 是一个面向 YouTube 的下载产品 monorepo。

当前版本已经具备完整的探测、登录、创建下载、任务详情和历史记录流程，并包含基础的免费计划限制：

- 匿名探测 YouTube 链接
- 登录后创建下载任务
- 每个账号最多 `3` 次免费下载
- 支持 `MP4 360p / 480p / 720p / 1080p`
- 支持 `仅音频 MP3`
- 下载历史按账号隔离
- 支持页、下载详情页、移动端导航

## 技术栈

- 前端：`Vite + React + Tailwind CSS + React Router + Clerk`
- 后端：`Go API + Go Worker`
- 数据库：`PostgreSQL`
- 队列/缓存：`Redis`
- 对象存储：S3 兼容对象存储，当前环境变量以 `R2_*` 命名
- 部署：`Docker Compose + Caddy`
- 域名入口：自定义域名

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
.github/
  workflows/  GitHub Actions
```

## 当前产品行为

- 仅支持 `youtube.com/watch` 和 `youtu.be`
- `POST /api/v1/videos/probe` 可匿名访问
- 创建下载、查看历史、查看任务详情、获取结果链接都必须登录
- 免费计划按账号限制为 `3` 次成功创建下载任务
- 所有规格都计入免费额度，包括 `audio-mp3`
- 免费额度用尽后，前端会展示“订阅即将上线”的提示
- 下载历史、状态、结果链接都只属于创建该任务的 Clerk 用户

## 环境变量

先复制模板：

```bash
cp .env.example .env
```

关键变量：

- `APP_DOMAIN`：站点域名，本地默认 `localhost`，生产环境填写你自己的正式域名
- `VITE_CLERK_PUBLISHABLE_KEY`：前端 Clerk publishable key
- `VITE_GA_MEASUREMENT_ID`：前端 GA4 Measurement ID，例如 `G-XXXXXXXXXX`
- `CLERK_SECRET_KEY`：后端校验 Clerk session token
- `DATABASE_URL`：Postgres 连接串
- `REDIS_ADDR`：Redis 地址
- `R2_*`：对象存储配置
- `DOWNLOAD_ROOT`：worker 临时工作目录
- `YTDLP_PROXY`：可选，给 `yt-dlp` 配置住宅代理或其他可用代理出口

## 本地启动

```bash
docker compose -f deploy/compose/docker-compose.yml up -d --build
```

启动后访问：

- Web：`http://localhost`
- 健康检查：`http://localhost/api/v1/healthz`
- 支持页：`http://localhost/support`

## 前端页面

- `/`：首页，包含 URL 输入、可用规格探测、剩余额度和最近下载
- `/downloads/:jobId`：下载详情页，显示任务状态和结果链接
- `/support`：FAQ / About / 产品说明页

GA4 接入后，前端会自动发送基础 `page_view`，并上报探测、登录、下载创建与下载链接打开等关键事件。

## API 概览

- `POST /api/v1/videos/probe`
  - 请求：`{ "url": "https://www.youtube.com/watch?v=..." }`
  - 返回：视频信息和可用 `profiles[]`
- `GET /api/v1/billing`
  - 需登录
  - 返回当前账号的免费额度摘要
- `POST /api/v1/downloads`
  - 需登录
  - 请求：`{ "url": "...", "profileId": "video-720p" }`
  - 免费额度用尽后返回 `402`
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

## 支付计划

Stripe 订阅方案已整理在 [BILLING_PLAN.md](./BILLING_PLAN.md)。

当前版本只实现免费额度与 paywall 提示，不包含真实支付链路。

## 自动部署

仓库包含 `push main` 后自动部署的 GitHub Actions 工作流。

部署目标：

- 服务器：你的 Linux 服务器
- 目录：推荐使用 `/opt/yt-downloader`
- 域名：你自己的正式域名

GitHub Secrets 需要配置：

- `DEPLOY_HOST`
- `DEPLOY_USER`
- `DEPLOY_SSH_KEY`
- `PROD_ENV_FILE`
- `YOUTUBE_COOKIES_B64`（可选，YouTube cookies 文件的 base64 内容）

服务器首次初始化需要：

- 安装 Docker Engine 和 Compose plugin
- 创建 `/opt/yt-downloader`
- 首次部署前在 GitHub Secrets 中配置完整的生产 `.env` 内容

推荐做法：

- 仓库根目录 `.env` 仅用于本地开发
- 生产环境变量统一放在 GitHub Secrets 的 `PROD_ENV_FILE`
- YouTube cookies 文件通过 `YOUTUBE_COOKIES_B64` 下发到服务器 `.runtime-secrets/youtube-cookies.txt`
- 若接入代理出口，可在 `PROD_ENV_FILE` 中加入 `YTDLP_PROXY`，格式例如 `http://user:pass@host:port`
- `CLERK_SECRET_KEY`、数据库密码、对象存储密钥等敏感值不要提交到仓库
- GitHub Actions 会直接把当前提交的源码同步到服务器，不依赖服务器访问 GitHub 仓库
- Workflow 会在部署时将 `PROD_ENV_FILE` 写入服务器 `/opt/yt-downloader/.env`

## Cloudflare 上线说明

- 域名注册商不限，可将 DNS 托管迁到 Cloudflare
- 推荐将根域 `A @` 指向你的源站 IP
- `CNAME www -> 你的主域名`
- 首次证书签发时，Cloudflare 先使用 `DNS only`
- 验证源站 HTTPS 正常后，再切到 `Proxied`
- Cloudflare SSL/TLS 模式使用 `Full (strict)`

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
