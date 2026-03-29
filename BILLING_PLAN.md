# Billing Plan

本文记录 `Archive` 下一阶段的 Stripe 订阅设计。当前仓库尚未接入真实支付链路，本文件作为后续实现的唯一设计依据。

## 已确定的产品规则

- 免费用户：每个账号终身可成功创建 `3` 个下载任务
- 免费额度：所有下载都计数，包括视频和 `audio-mp3`
- 扣减时机：任务成功入库并成功入队后计为 1 次
- 付费套餐：单一 `Pro`
  - `$9/月`
  - `$79/年`
- 付费权益：仅提供无限下载
- 订阅方式：`Stripe Checkout + Stripe Customer Portal`
- 取消订阅：到当前计费周期结束后失效

## 计划中的后端能力

### 数据表

后续 Stripe 接入将基于 `billing_accounts` 表扩展，而不是回到旧 `users` 表：

- `clerk_user_id`
- `plan_code`
- `subscription_status`
- `billing_interval`
- `stripe_customer_id`
- `stripe_subscription_id`
- `current_period_end`
- `cancel_at_period_end`
- `free_downloads_limit`
- `free_downloads_used`

同时增加 `stripe_webhook_events` 作为 webhook 幂等去重表。

### 预期 API

- `GET /api/v1/billing`
  - 返回当前账号套餐和剩余额度
- `POST /api/v1/billing/checkout`
  - 创建 Stripe Checkout Session
- `POST /api/v1/billing/portal`
  - 创建 Stripe Customer Portal Session
- `POST /api/v1/stripe/webhook`
  - 接收 Stripe Webhook

### Stripe 状态同步

后续将处理这些事件：

- `checkout.session.completed`
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`

## 计划中的前端能力

- 首页展示 `Free / Pro` 套餐区
- 免费额度用尽后展示 paywall
- 已订阅用户展示 `Manage subscription`
- 支付成功后回跳首页并刷新 billing 状态

## 未来环境变量

- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_PRICE_PRO_MONTHLY_ID`
- `STRIPE_PRICE_PRO_YEARLY_ID`
- `APP_BASE_URL`

## 当前阶段说明

当前阶段只实现：

- billing 文档落地
- 免费用户 3 次下载限制
- 前端展示限额与“订阅即将上线”提示

当前阶段不会实现：

- 真实 Stripe Checkout
- Customer Portal
- Webhook
- 自动切换 `free -> pro`
