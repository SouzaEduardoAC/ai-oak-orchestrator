package valkey

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ecoza/ai-oak-orchestrator/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func NewClient(rawURL string, password string) (*Client, error) {
	// Standardize password: trim whitespace and any accidental quotes
	password = strings.TrimSpace(password)
	password = strings.Trim(password, "'\"")

	// We parse the URL to get the host/port correctly, but we'll manually
	// set the password to avoid any encoding/parsing issues.
	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		// Fallback for cases where the URL might be just a host:port
		opts = &redis.Options{
			Addr: rawURL,
		}
	}

	// Always use the explicit password if provided
	if password != "" {
		opts.Password = password
	}
	
	// Force RESP2 for compatibility with older Valkey/Redis handshakes
	opts.Protocol = 2

	client := redis.NewClient(opts)

	// Verify connection with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
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

	return c.client.Set(ctx, key, val, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
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
