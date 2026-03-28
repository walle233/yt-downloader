package queue

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const downloadsQueue = "downloads:queue"

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(addr string) *RedisQueue {
	return &RedisQueue{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}

func (q *RedisQueue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

func (q *RedisQueue) EnqueueDownload(ctx context.Context, id int64) error {
	return q.client.LPush(ctx, downloadsQueue, id).Err()
}

func (q *RedisQueue) BlockingPopDownload(ctx context.Context, timeout time.Duration) (int64, error) {
	values, err := q.client.BRPop(ctx, timeout, downloadsQueue).Result()
	if err != nil {
		return 0, err
	}

	if len(values) != 2 {
		return 0, fmt.Errorf("unexpected BRPOP response length: %d", len(values))
	}

	id, err := strconv.ParseInt(values[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse queue payload: %w", err)
	}

	return id, nil
}
