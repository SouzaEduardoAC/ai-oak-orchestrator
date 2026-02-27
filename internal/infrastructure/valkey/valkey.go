package valkey

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/valkey-io/valkey-go"
)

type Client struct {
	client valkey.Client
}

func NewClient(url string) (*Client, error) {
	opts, err := valkey.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client, err := valkey.NewClient(opts)
	if err != nil {
		return nil, err
	}

	if err := client.Do(context.Background(), client.B().Ping().Build()).Error(); err != nil {
		return nil, err
	}

	return &Client{client: client}, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		val = string(data)
	}

	b := c.client.B().Set().Key(key).Value(val)
	var cmd valkey.Completed
	if expiration > 0 {
		cmd = b.Ex(expiration).Build()
	} else {
		cmd = b.Build()
	}
	return c.client.Do(ctx, cmd).Error()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Do(ctx, c.client.B().Get().Key(key).Build()).ToString()
}

func (c *Client) Del(ctx context.Context, key string) error {
	return c.client.Do(ctx, c.client.B().Del().Key(key).Build()).Error()
}

func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Do(ctx, c.client.B().Keys().Pattern(pattern).Build()).AsStrSlice()
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
