// Package middleware provides common middleware for Fiber framework
package middleware

import (
    "time"
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/limiter"
)

// RateLimiter returns a rate limiting middleware for Fiber
func RateLimiter() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        100,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "error": "Too many requests",
            })
        },
    })
}

// AuthMiddleware provides JWT authentication for Fiber
func AuthMiddleware(secret string) fiber.Handler {
    return func(c fiber.Ctx) error {
        // TODO: Implement JWT validation
        token := c.Get("Authorization")
        if token == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Missing authorization header",
            })
        }
        
        // For now, just check if token exists
        if token != "Bearer valid-token" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Invalid token",
            })
        }
        
        return c.Next()
    }
}