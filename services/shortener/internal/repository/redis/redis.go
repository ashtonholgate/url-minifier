package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Client wraps Redis client
type Client struct {
	client *redis.Client
}

// NewClient creates a new Redis client
func NewClient(ctx context.Context, uri string, password string, db int) (*Client, error) {
	// Parse options from URI
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	// Override with provided options
	opt.Password = password
	opt.DB = db

	// Create client
	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

// Close disconnects from Redis
func (c *Client) Close() error {
	return c.client.Close()
}

// Get retrieves a value from Redis
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set stores a value in Redis with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Del deletes a key from Redis
func (c *Client) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
