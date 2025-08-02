// Package health provides health check utilities for Fiber framework
package health

import (
    "time"
    "github.com/gofiber/fiber/v3"
)

// HealthHandler is a simple health check handler for Fiber
func HealthHandler(c fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status": "healthy",
        "timestamp": time.Now().Format(time.RFC3339),
        "service": c.App().Config().AppName,
    })
}

// ReadyHandler is a readiness check handler for Fiber
func ReadyHandler(c fiber.Ctx) error {
    // TODO: Add actual readiness checks (DB connection, etc.)
    return c.JSON(fiber.Map{
        "status": "ready",
        "timestamp": time.Now().Format(time.RFC3339),
        "service": c.App().Config().AppName,
    })
}