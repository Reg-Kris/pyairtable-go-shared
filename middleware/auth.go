// Package middleware provides common HTTP middleware for authentication, logging, metrics, and rate limiting
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Reg-Kris/pyairtable-go-shared/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	Email     string `json:"email"`
	Roles     []string `json:"roles"`
	Scopes    []string `json:"scopes"`
	jwt.RegisteredClaims
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string
	Issuer        string
	SkipPaths     []string
	RequiredRoles []string
	RequiredScopes []string
}

// JWT returns a JWT authentication middleware
func JWT(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for specified paths
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, errors.NewUnauthorizedError("Missing authorization header"))
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, errors.NewUnauthorizedError("Invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.NewTokenInvalidError()
			}
			return []byte(config.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, errors.NewTokenInvalidError().WithCause(err))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, errors.NewTokenInvalidError())
			c.Abort()
			return
		}

		// Check token expiration
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, errors.NewTokenExpiredError())
			c.Abort()
			return
		}

		// Check issuer
		if config.Issuer != "" && claims.Issuer != config.Issuer {
			c.JSON(http.StatusUnauthorized, errors.NewTokenInvalidError())
			c.Abort()
			return
		}

		// Check required roles
		if len(config.RequiredRoles) > 0 {
			if !hasAnyRole(claims.Roles, config.RequiredRoles) {
				c.JSON(http.StatusForbidden, errors.NewForbiddenError("Insufficient role permissions"))
				c.Abort()
				return
			}
		}

		// Check required scopes
		if len(config.RequiredScopes) > 0 {
			if !hasAnyScope(claims.Scopes, config.RequiredScopes) {
				c.JSON(http.StatusForbidden, errors.NewInsufficientScopeError(strings.Join(config.RequiredScopes, ", ")))
				c.Abort()
				return
			}
		}

		// Set claims in context
		ctx := context.WithValue(c.Request.Context(), "claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireRole returns middleware that requires specific roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaimsFromContext(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, errors.NewUnauthorizedError("Missing authentication"))
			c.Abort()
			return
		}

		if !hasAnyRole(claims.Roles, roles) {
			c.JSON(http.StatusForbidden, errors.NewForbiddenError("Insufficient role permissions"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireScope returns middleware that requires specific scopes
func RequireScope(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaimsFromContext(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, errors.NewUnauthorizedError("Missing authentication"))
			c.Abort()
			return
		}

		if !hasAnyScope(claims.Scopes, scopes) {
			c.JSON(http.StatusForbidden, errors.NewInsufficientScopeError(strings.Join(scopes, ", ")))
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetClaimsFromContext extracts JWT claims from Gin context
func GetClaimsFromContext(c *gin.Context) *JWTClaims {
	if claims, ok := c.Request.Context().Value("claims").(*JWTClaims); ok {
		return claims
	}
	return nil
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c *gin.Context) string {
	if userID, ok := c.Request.Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(c *gin.Context) string {
	if tenantID, ok := c.Request.Context().Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

// hasAnyRole checks if user has any of the required roles
func hasAnyRole(userRoles, requiredRoles []string) bool {
	roleMap := make(map[string]bool)
	for _, role := range userRoles {
		roleMap[role] = true
	}
	
	for _, requiredRole := range requiredRoles {
		if roleMap[requiredRole] {
			return true
		}
	}
	
	return false
}

// hasAnyScope checks if user has any of the required scopes
func hasAnyScope(userScopes, requiredScopes []string) bool {
	scopeMap := make(map[string]bool)
	for _, scope := range userScopes {
		scopeMap[scope] = true
	}
	
	for _, requiredScope := range requiredScopes {
		if scopeMap[requiredScope] {
			return true
		}
	}
	
	return false
}

// CreateToken creates a new JWT token
func CreateToken(claims *JWTClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}