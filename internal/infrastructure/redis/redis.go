package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.rdb.Keys(ctx, pattern).Result()
}

func (c *Client) SaveSession(ctx context.Context, session *domain.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return c.Set(ctx, "session:"+session.ID, data, 24*time.Hour)
}

func (c *Client) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	data, err := c.Get(ctx, "session:"+id)
	if err != nil {
		return nil, err
	}
	var session domain.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}
	return &session, nil
}
