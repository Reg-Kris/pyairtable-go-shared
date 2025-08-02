package models

import (
	"time"
)

// Tenant represents a tenant/organization in the multi-tenant system
type Tenant struct {
	BaseModel
	Name        string    `json:"name" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null"`
	Domain      string    `json:"domain" gorm:"uniqueIndex"`
	Email       string    `json:"email" gorm:"not null"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Country     string    `json:"country" gorm:"not null"`
	PostalCode  string    `json:"postal_code"`
	Website     string    `json:"website"`
	Logo        string    `json:"logo"`
	Description string    `json:"description"`
	Status      Status    `json:"status" gorm:"default:'active'"`
	PlanType    PlanType  `json:"plan_type" gorm:"default:'free'"`
	TrialEndsAt *time.Time `json:"trial_ends_at"`
	Settings    JSON      `json:"settings" gorm:"type:jsonb"`
	Metadata    JSON      `json:"metadata" gorm:"type:jsonb"`
	
	// Relationships
	Users      []User      `json:"users,omitempty" gorm:"foreignKey:TenantID"`
	Workspaces []Workspace `json:"workspaces,omitempty" gorm:"foreignKey:TenantID"`
	Quotas     []TenantQuota `json:"quotas,omitempty" gorm:"foreignKey:TenantID"`
}

// PlanType represents the subscription plan type
type PlanType string

const (
	PlanTypeFree       PlanType = "free"
	PlanTypeBasic      PlanType = "basic"
	PlanTypeProfessional PlanType = "professional"
	PlanTypeEnterprise PlanType = "enterprise"
)

// IsActive checks if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive
}

// IsTrialExpired checks if the trial period has expired
func (t *Tenant) IsTrialExpired() bool {
	return t.TrialEndsAt != nil && time.Now().After(*t.TrialEndsAt)
}

// GetUserCount returns the number of users in the tenant
func (t *Tenant) GetUserCount() int {
	return len(t.Users)
}

// GetWorkspaceCount returns the number of workspaces in the tenant
func (t *Tenant) GetWorkspaceCount() int {
	return len(t.Workspaces)
}

// CanCreateWorkspace checks if the tenant can create more workspaces
func (t *Tenant) CanCreateWorkspace() bool {
	limits := t.GetPlanLimits()
	return t.GetWorkspaceCount() < limits.MaxWorkspaces
}

// CanInviteUser checks if the tenant can invite more users
func (t *Tenant) CanInviteUser() bool {
	limits := t.GetPlanLimits()
	return t.GetUserCount() < limits.MaxUsers
}

// GetPlanLimits returns the limits for the current plan
func (t *Tenant) GetPlanLimits() PlanLimits {
	switch t.PlanType {
	case PlanTypeFree:
		return PlanLimits{
			MaxUsers:      5,
			MaxWorkspaces: 3,
			MaxRecords:    1000,
			MaxStorage:    1024 * 1024 * 100, // 100MB
			Features:      []string{"basic_features"},
		}
	case PlanTypeBasic:
		return PlanLimits{
			MaxUsers:      25,
			MaxWorkspaces: 10,
			MaxRecords:    10000,
			MaxStorage:    1024 * 1024 * 1024, // 1GB
			Features:      []string{"basic_features", "advanced_features"},
		}
	case PlanTypeProfessional:
		return PlanLimits{
			MaxUsers:      100,
			MaxWorkspaces: 50,
			MaxRecords:    100000,
			MaxStorage:    1024 * 1024 * 1024 * 10, // 10GB
			Features:      []string{"basic_features", "advanced_features", "professional_features"},
		}
	case PlanTypeEnterprise:
		return PlanLimits{
			MaxUsers:      -1, // Unlimited
			MaxWorkspaces: -1, // Unlimited
			MaxRecords:    -1, // Unlimited
			MaxStorage:    -1, // Unlimited
			Features:      []string{"all_features"},
		}
	default:
		return PlanLimits{}
	}
}

// PlanLimits represents the limits for a subscription plan
type PlanLimits struct {
	MaxUsers      int      `json:"max_users"`
	MaxWorkspaces int      `json:"max_workspaces"`
	MaxRecords    int      `json:"max_records"`
	MaxStorage    int64    `json:"max_storage"`
	Features      []string `json:"features"`
}

// TenantQuota represents resource quotas for a tenant
type TenantQuota struct {
	BaseModel
	TenantID     uint   `json:"tenant_id" gorm:"index;not null"`
	ResourceType string `json:"resource_type" gorm:"not null"`
	Used         int64  `json:"used" gorm:"default:0"`
	Limit        int64  `json:"limit" gorm:"not null"`
	
	// Relationships
	Tenant Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// IsExceeded checks if the quota is exceeded
func (q *TenantQuota) IsExceeded() bool {
	return q.Used >= q.Limit
}

// GetUsagePercentage returns the usage percentage
func (q *TenantQuota) GetUsagePercentage() float64 {
	if q.Limit == 0 {
		return 0
	}
	return float64(q.Used) / float64(q.Limit) * 100
}

// CanAllocate checks if the specified amount can be allocated
func (q *TenantQuota) CanAllocate(amount int64) bool {
	return q.Used+amount <= q.Limit
}

// Allocate allocates the specified amount
func (q *TenantQuota) Allocate(amount int64) bool {
	if !q.CanAllocate(amount) {
		return false
	}
	q.Used += amount
	return true
}

// Deallocate deallocates the specified amount
func (q *TenantQuota) Deallocate(amount int64) {
	q.Used -= amount
	if q.Used < 0 {
		q.Used = 0
	}
}

// TenantInvitation represents an invitation to join a tenant
type TenantInvitation struct {
	BaseModel
	TenantID   uint      `json:"tenant_id" gorm:"index;not null"`
	Email      string    `json:"email" gorm:"not null"`
	Token      string    `json:"token" gorm:"uniqueIndex;not null"`
	Role       string    `json:"role" gorm:"not null"`
	InvitedBy  uint      `json:"invited_by" gorm:"not null"`
	ExpiresAt  time.Time `json:"expires_at" gorm:"not null"`
	AcceptedAt *time.Time `json:"accepted_at"`
	
	// Relationships
	Tenant    Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	InvitedByUser User `json:"invited_by_user,omitempty" gorm:"foreignKey:InvitedBy"`
}

// IsExpired checks if the invitation is expired
func (i *TenantInvitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsAccepted checks if the invitation has been accepted
func (i *TenantInvitation) IsAccepted() bool {
	return i.AcceptedAt != nil
}

// IsValid checks if the invitation is valid
func (i *TenantInvitation) IsValid() bool {
	return !i.IsExpired() && !i.IsAccepted()
}

// Accept accepts the invitation
func (i *TenantInvitation) Accept() {
	now := time.Now()
	i.AcceptedAt = &now
}

// TenantSettings represents tenant-specific settings
type TenantSettings struct {
	Branding     BrandingSettings     `json:"branding"`
	Security     SecuritySettings     `json:"security"`
	Integrations IntegrationSettings  `json:"integrations"`
	Notifications NotificationSettings `json:"notifications"`
	Features     FeatureSettings      `json:"features"`
}

// BrandingSettings represents branding customization settings
type BrandingSettings struct {
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	Logo           string `json:"logo"`
	Favicon        string `json:"favicon"`
	CustomCSS      string `json:"custom_css"`
}

// SecuritySettings represents security-related settings
type SecuritySettings struct {
	PasswordPolicy       PasswordPolicy `json:"password_policy"`
	TwoFactorRequired    bool          `json:"two_factor_required"`
	SessionTimeout       int           `json:"session_timeout"`
	MaxLoginAttempts     int           `json:"max_login_attempts"`
	IPWhitelistEnabled   bool          `json:"ip_whitelist_enabled"`
	IPWhitelist          []string      `json:"ip_whitelist"`
}

// PasswordPolicy represents password policy settings
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	ExpirationDays   int  `json:"expiration_days"`
}

// IntegrationSettings represents integration settings
type IntegrationSettings struct {
	EmailProvider string                 `json:"email_provider"`
	SMSProvider   string                 `json:"sms_provider"`
	Webhooks      []WebhookConfig        `json:"webhooks"`
	APIKeys       map[string]string      `json:"api_keys"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	URL     string   `json:"url"`
	Events  []string `json:"events"`
	Secret  string   `json:"secret"`
	Enabled bool     `json:"enabled"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	EmailEnabled  bool     `json:"email_enabled"`
	SMSEnabled    bool     `json:"sms_enabled"`
	PushEnabled   bool     `json:"push_enabled"`
	Events        []string `json:"events"`
	Templates     map[string]NotificationTemplate `json:"templates"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	Subject  string            `json:"subject"`
	Body     string            `json:"body"`
	Variables map[string]string `json:"variables"`
}

// FeatureSettings represents feature toggle settings
type FeatureSettings struct {
	EnabledFeatures  []string          `json:"enabled_features"`
	DisabledFeatures []string          `json:"disabled_features"`
	FeatureFlags     map[string]bool   `json:"feature_flags"`
	CustomSettings   map[string]interface{} `json:"custom_settings"`
}