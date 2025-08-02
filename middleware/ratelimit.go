package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Reg-Kris/pyairtable-go-shared/cache"
	"github.com/Reg-Kris/pyairtable-go-shared/errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int           // Requests per second allowed
	BurstSize         int           // Burst size for token bucket
	WindowSize        time.Duration // Time window for sliding window
	MaxRequests       int           // Max requests per window
	KeyFunc           func(*gin.Context) string // Function to generate rate limit key
	SkipPaths         []string      // Paths to skip rate limiting
	UseRedis          bool          // Whether to use Redis for distributed rate limiting
	RedisClient       *cache.Client // Redis client for distributed rate limiting
}

// DefaultKeyFunc generates rate limit key based on client IP
func DefaultKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// UserKeyFunc generates rate limit key based on user ID
func UserKeyFunc(c *gin.Context) string {
	userID := GetUserIDFromContext(c)
	if userID != "" {
		return "user:" + userID
	}
	return c.ClientIP()
}

// TenantKeyFunc generates rate limit key based on tenant ID
func TenantKeyFunc(c *gin.Context) string {
	tenantID := GetTenantIDFromContext(c)
	if tenantID != "" {
		return "tenant:" + tenantID
	}
	return c.ClientIP()
}

// TokenBucketRateLimit implements token bucket rate limiting
func TokenBucketRateLimit(config RateLimitConfig) gin.HandlerFunc {
	limiters := make(map[string]*rate.Limiter)
	
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}
	
	return func(c *gin.Context) {
		// Skip rate limiting for specified paths
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}
		
		key := config.KeyFunc(c)
		
		// Get or create limiter for this key
		limiter, exists := limiters[key]
		if !exists {
			limiter = rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.BurstSize)
			limiters[key] = limiter
		}
		
		// Check if request is allowed
		if !limiter.Allow() {
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerSecond))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Second).Unix(), 10))
			
			retryAfter := int(limiter.Reserve().Delay().Seconds()) + 1
			c.JSON(http.StatusTooManyRequests, errors.NewRateLimitedError(retryAfter))
			c.Abort()
			return
		}
		
		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerSecond))
		
		c.Next()
	}
}

// SlidingWindowRateLimit implements sliding window rate limiting using Redis
func SlidingWindowRateLimit(config RateLimitConfig) gin.HandlerFunc {
	if config.RedisClient == nil {
		panic("Redis client is required for sliding window rate limiting")
	}
	
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}
	
	if config.WindowSize == 0 {
		config.WindowSize = time.Minute
	}
	
	return func(c *gin.Context) {
		// Skip rate limiting for specified paths
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}
		
		key := "ratelimit:" + config.KeyFunc(c)
		ctx := context.Background()
		now := time.Now()
		window := now.Truncate(config.WindowSize)
		
		// Get current count for this window
		countKey := fmt.Sprintf("%s:%d", key, window.Unix())
		
		// Increment counter
		count, err := config.RedisClient.IncrementBy(ctx, countKey, 1)
		if err != nil {
			// If Redis is unavailable, allow the request but log the error
			c.Next()
			return
		}
		
		// Set expiration for the key
		if count == 1 {
			config.RedisClient.Expire(ctx, countKey, config.WindowSize)
		}
		
		// Check if limit exceeded
		if int(count) > config.MaxRequests {
			remaining := 0
			reset := window.Add(config.WindowSize).Unix()
			
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.MaxRequests))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
			c.Header("X-RateLimit-Window", config.WindowSize.String())
			
			retryAfter := int(window.Add(config.WindowSize).Sub(now).Seconds()) + 1
			c.JSON(http.StatusTooManyRequests, errors.NewRateLimitedError(retryAfter))
			c.Abort()
			return
		}
		
		// Add rate limit headers
		remaining := config.MaxRequests - int(count)
		reset := window.Add(config.WindowSize).Unix()
		
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.MaxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
		c.Header("X-RateLimit-Window", config.WindowSize.String())
		
		c.Next()
	}
}

// FixedWindowRateLimit implements fixed window rate limiting using Redis
func FixedWindowRateLimit(config RateLimitConfig) gin.HandlerFunc {
	if config.RedisClient == nil {
		panic("Redis client is required for fixed window rate limiting")
	}
	
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}
	
	if config.WindowSize == 0 {
		config.WindowSize = time.Minute
	}
	
	return func(c *gin.Context) {
		// Skip rate limiting for specified paths
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}
		
		key := "ratelimit:" + config.KeyFunc(c)
		ctx := context.Background()
		now := time.Now()
		window := now.Truncate(config.WindowSize)
		
		// Create window-specific key
		windowKey := fmt.Sprintf("%s:%d", key, window.Unix())
		
		// Try to increment counter
		count, err := config.RedisClient.IncrementBy(ctx, windowKey, 1)
		if err != nil {
			// If Redis is unavailable, allow the request but log the error
			c.Next()
			return
		}
		
		// Set expiration for new windows
		if count == 1 {
			config.RedisClient.Expire(ctx, windowKey, config.WindowSize)
		}
		
		// Check if limit exceeded
		if int(count) > config.MaxRequests {
			remaining := 0
			reset := window.Add(config.WindowSize).Unix()
			
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.MaxRequests))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
			c.Header("X-RateLimit-Window", config.WindowSize.String())
			
			retryAfter := int(window.Add(config.WindowSize).Sub(now).Seconds()) + 1
			c.JSON(http.StatusTooManyRequests, errors.NewRateLimitedError(retryAfter))
			c.Abort()
			return
		}
		
		// Add rate limit headers
		remaining := config.MaxRequests - int(count)
		reset := window.Add(config.WindowSize).Unix()
		
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.MaxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
		c.Header("X-RateLimit-Window", config.WindowSize.String())
		
		c.Next()
	}
}

// PerUserRateLimit applies different rate limits based on user type
func PerUserRateLimit(config map[string]RateLimitConfig, defaultConfig RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaimsFromContext(c)
		var selectedConfig RateLimitConfig
		
		if claims != nil {
			// Check user roles for specific rate limits
			for _, role := range claims.Roles {
				if cfg, exists := config[role]; exists {
					selectedConfig = cfg
					break
				}
			}
		}
		
		// Use default config if no specific config found
		if selectedConfig.MaxRequests == 0 {
			selectedConfig = defaultConfig
		}
		
		// Apply the selected rate limit
		if selectedConfig.UseRedis {
			FixedWindowRateLimit(selectedConfig)(c)
		} else {
			TokenBucketRateLimit(selectedConfig)(c)
		}
	}
}