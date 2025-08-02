package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	TenantModel
	Email           string    `json:"email" gorm:"uniqueIndex;not null"`
	FirstName       string    `json:"first_name" gorm:"not null"`
	LastName        string    `json:"last_name" gorm:"not null"`
	PasswordHash    string    `json:"-" gorm:"not null"`
	Status          Status    `json:"status" gorm:"default:'pending'"`
	EmailVerified   bool      `json:"email_verified" gorm:"default:false"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	LoginCount      int       `json:"login_count" gorm:"default:0"`
	Timezone        string    `json:"timezone" gorm:"default:'UTC'"`
	Language        string    `json:"language" gorm:"default:'en'"`
	Avatar          string    `json:"avatar"`
	Phone           string    `json:"phone"`
	Bio             string    `json:"bio"`
	Preferences     JSON      `json:"preferences" gorm:"type:jsonb"`
	Metadata        JSON      `json:"metadata" gorm:"type:jsonb"`
	
	// Relationships
	Roles       []Role       `json:"roles" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:user_permissions;"`
	Sessions    []Session    `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	APIKeys     []APIKey     `json:"api_keys,omitempty" gorm:"foreignKey:UserID"`
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsEmailVerified checks if the user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerified
}

// CanLogin checks if the user can login
func (u *User) CanLogin() bool {
	return u.IsActive() && u.IsEmailVerified()
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func (u *User) HasPermission(permissionName string) bool {
	// Check direct permissions
	for _, permission := range u.Permissions {
		if permission.Name == permissionName {
			return true
		}
	}
	
	// Check role permissions
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			if permission.Name == permissionName {
				return true
			}
		}
	}
	
	return false
}

// UpdateLastLogin updates the last login timestamp and count
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.LoginCount++
}

// Role represents a role in the system
type Role struct {
	BaseModel
	Name        string       `json:"name" gorm:"uniqueIndex;not null"`
	Description string       `json:"description"`
	IsSystem    bool         `json:"is_system" gorm:"default:false"`
	
	// Relationships
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	Users       []User       `json:"users,omitempty" gorm:"many2many:user_roles;"`
}

// Permission represents a permission in the system
type Permission struct {
	BaseModel
	Name        string `json:"name" gorm:"uniqueIndex;not null"`
	Description string `json:"description"`
	Resource    string `json:"resource" gorm:"not null"`
	Action      string `json:"action" gorm:"not null"`
	
	// Relationships
	Roles []Role `json:"roles,omitempty" gorm:"many2many:role_permissions;"`
	Users []User `json:"users,omitempty" gorm:"many2many:user_permissions;"`
}

// Session represents a user session
type Session struct {
	BaseModel
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	
	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// IsExpired checks if the session is expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	BaseModel
	UserID      uint      `json:"user_id" gorm:"index;not null"`
	Name        string    `json:"name" gorm:"not null"`
	KeyHash     string    `json:"key_hash" gorm:"uniqueIndex;not null"`
	Prefix      string    `json:"prefix" gorm:"not null"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	UsageCount  int       `json:"usage_count" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Scopes      []string  `json:"scopes" gorm:"type:text[]"`
	IPWhitelist []string  `json:"ip_whitelist" gorm:"type:text[]"`
	
	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// IsExpired checks if the API key is expired
func (k *APIKey) IsExpired() bool {
	return k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt)
}

// IsValid checks if the API key is valid
func (k *APIKey) IsValid() bool {
	return k.IsActive && !k.IsExpired()
}

// HasScope checks if the API key has a specific scope
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// IsIPAllowed checks if an IP address is allowed
func (k *APIKey) IsIPAllowed(ip string) bool {
	if len(k.IPWhitelist) == 0 {
		return true // No restrictions
	}
	
	for _, allowedIP := range k.IPWhitelist {
		if allowedIP == ip {
			return true
		}
	}
	
	return false
}

// UpdateUsage updates usage statistics
func (k *APIKey) UpdateUsage() {
	now := time.Now()
	k.LastUsedAt = &now
	k.UsageCount++
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Theme               string            `json:"theme"`
	Language            string            `json:"language"`
	Timezone            string            `json:"timezone"`
	DateFormat          string            `json:"date_format"`
	TimeFormat          string            `json:"time_format"`
	EmailNotifications  bool              `json:"email_notifications"`
	PushNotifications   bool              `json:"push_notifications"`
	TwoFactorEnabled    bool              `json:"two_factor_enabled"`
	CustomSettings      map[string]interface{} `json:"custom_settings"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	BaseModel
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Token     string    `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time `json:"used_at"`
	
	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// IsExpired checks if the token is expired
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *PasswordResetToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the token is valid
func (t *PasswordResetToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// MarkAsUsed marks the token as used
func (t *PasswordResetToken) MarkAsUsed() {
	now := time.Now()
	t.UsedAt = &now
}