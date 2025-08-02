// Package cache provides Redis caching utilities with circuit breaker pattern
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Reg-Kris/pyairtable-go-shared/config"
	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
)

// Client represents a Redis client with circuit breaker
type Client struct {
	redis   *redis.Client
	breaker *gobreaker.CircuitBreaker
}

// New creates a new Redis client with circuit breaker
func New(cfg *config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Configure circuit breaker
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "redis-cache",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Log state changes
			fmt.Printf("Circuit breaker %s changed from %v to %v\n", name, from, to)
		},
	})

	return &Client{
		redis:   rdb,
		breaker: breaker,
	}, nil
}

// Set stores a value in cache with expiration
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	_, err = c.breaker.Execute(func() (interface{}, error) {
		return nil, c.redis.Set(ctx, key, data, expiration).Err()
	})

	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.redis.Get(ctx, key).Result()
	})

	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	data, ok := result.(string)
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		return nil, c.redis.Del(ctx, key).Err()
	})

	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.redis.Exists(ctx, key).Result()
	})

	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	count, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type: %T", result)
	}

	return count > 0, nil
}

// SetNX sets a key only if it doesn't exist
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.redis.SetNX(ctx, key, data, expiration).Result()
	})

	if err != nil {
		return false, fmt.Errorf("failed to set cache: %w", err)
	}

	success, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected result type: %T", result)
	}

	return success, nil
}

// Expire sets expiration for a key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		return nil, c.redis.Expire(ctx, key, expiration).Err()
	})

	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}

	return nil
}

// Increment increments a numeric value
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.redis.Incr(ctx, key).Result()
	})

	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}

	count, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	return count, nil
}

// IncrementBy increments a numeric value by delta
func (c *Client) IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.redis.IncrBy(ctx, key, delta).Result()
	})

	if err != nil {
		return 0, fmt.Errorf("failed to increment by: %w", err)
	}

	count, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	return count, nil
}

// GetMultiple retrieves multiple values from cache
func (c *Client) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.redis.MGet(ctx, keys...).Result()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get multiple: %w", err)
	}

	values, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	results := make(map[string]interface{})
	for i, key := range keys {
		if i < len(values) && values[i] != nil {
			results[key] = values[i]
		}
	}

	return results, nil
}

// Health checks the Redis connection health
func (c *Client) Health(ctx context.Context) error {
	_, err := c.breaker.Execute(func() (interface{}, error) {
		return nil, c.redis.Ping(ctx).Err()
	})

	if err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.redis.Close()
}

// GetStats returns Redis connection statistics
func (c *Client) GetStats() map[string]interface{} {
	stats := c.redis.PoolStats()
	
	return map[string]interface{}{
		"hits":         stats.Hits,
		"misses":       stats.Misses,
		"timeouts":     stats.Timeouts,
		"total_conns":  stats.TotalConns,
		"idle_conns":   stats.IdleConns,
		"stale_conns":  stats.StaleConns,
	}
}

// GetBreakerStats returns circuit breaker statistics
func (c *Client) GetBreakerStats() map[string]interface{} {
	counts := c.breaker.Counts()
	
	return map[string]interface{}{
		"requests":              counts.Requests,
		"total_successes":       counts.TotalSuccesses,
		"total_failures":        counts.TotalFailures,
		"consecutive_successes": counts.ConsecutiveSuccesses,
		"consecutive_failures":  counts.ConsecutiveFailures,
		"state":                 c.breaker.State().String(),
	}
}

// Custom errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)