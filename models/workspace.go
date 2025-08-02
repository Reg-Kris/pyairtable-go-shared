package models

import (
	"time"
)

// Workspace represents a workspace in the system
type Workspace struct {
	TenantModel
	Name        string    `json:"name" gorm:"not null"`
	Slug        string    `json:"slug" gorm:"not null"`
	Description string    `json:"description"`
	Color       string    `json:"color" gorm:"default:'#3B82F6'"`
	Icon        string    `json:"icon"`
	Status      Status    `json:"status" gorm:"default:'active'"`
	IsPublic    bool      `json:"is_public" gorm:"default:false"`
	Settings    JSON      `json:"settings" gorm:"type:jsonb"`
	Metadata    JSON      `json:"metadata" gorm:"type:jsonb"`
	
	// Relationships
	Tenant Tenant              `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	Tables []Table             `json:"tables,omitempty" gorm:"foreignKey:WorkspaceID"`
	Members []WorkspaceMember  `json:"members,omitempty" gorm:"foreignKey:WorkspaceID"`
	Invitations []WorkspaceInvitation `json:"invitations,omitempty" gorm:"foreignKey:WorkspaceID"`
}

// IsActive checks if the workspace is active
func (w *Workspace) IsActive() bool {
	return w.Status == StatusActive
}

// GetMemberCount returns the number of members in the workspace
func (w *Workspace) GetMemberCount() int {
	return len(w.Members)
}

// GetTableCount returns the number of tables in the workspace
func (w *Workspace) GetTableCount() int {
	return len(w.Tables)
}

// HasMember checks if a user is a member of the workspace
func (w *Workspace) HasMember(userID uint) bool {
	for _, member := range w.Members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}

// GetMemberRole returns the role of a user in the workspace
func (w *Workspace) GetMemberRole(userID uint) WorkspaceRole {
	for _, member := range w.Members {
		if member.UserID == userID {
			return member.Role
		}
	}
	return ""
}

// WorkspaceRole represents roles within a workspace
type WorkspaceRole string

const (
	WorkspaceRoleOwner       WorkspaceRole = "owner"
	WorkspaceRoleAdmin       WorkspaceRole = "admin"
	WorkspaceRoleEditor      WorkspaceRole = "editor"
	WorkspaceRoleCommenter   WorkspaceRole = "commenter"
	WorkspaceRoleViewer      WorkspaceRole = "viewer"
)

// CanManageWorkspace checks if the role can manage workspace settings
func (r WorkspaceRole) CanManageWorkspace() bool {
	return r == WorkspaceRoleOwner || r == WorkspaceRoleAdmin
}

// CanCreateTables checks if the role can create tables
func (r WorkspaceRole) CanCreateTables() bool {
	return r == WorkspaceRoleOwner || r == WorkspaceRoleAdmin || r == WorkspaceRoleEditor
}

// CanEditTables checks if the role can edit tables
func (r WorkspaceRole) CanEditTables() bool {
	return r == WorkspaceRoleOwner || r == WorkspaceRoleAdmin || r == WorkspaceRoleEditor
}

// CanDeleteTables checks if the role can delete tables
func (r WorkspaceRole) CanDeleteTables() bool {
	return r == WorkspaceRoleOwner || r == WorkspaceRoleAdmin
}

// CanInviteMembers checks if the role can invite members
func (r WorkspaceRole) CanInviteMembers() bool {
	return r == WorkspaceRoleOwner || r == WorkspaceRoleAdmin
}

// CanManageMembers checks if the role can manage members
func (r WorkspaceRole) CanManageMembers() bool {
	return r == WorkspaceRoleOwner || r == WorkspaceRoleAdmin
}

// WorkspaceMember represents a member of a workspace
type WorkspaceMember struct {
	BaseModel
	WorkspaceID uint          `json:"workspace_id" gorm:"index;not null"`
	UserID      uint          `json:"user_id" gorm:"index;not null"`
	Role        WorkspaceRole `json:"role" gorm:"not null"`
	JoinedAt    time.Time     `json:"joined_at" gorm:"default:CURRENT_TIMESTAMP"`
	
	// Relationships
	Workspace Workspace `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// WorkspaceInvitation represents an invitation to join a workspace
type WorkspaceInvitation struct {
	BaseModel
	WorkspaceID uint          `json:"workspace_id" gorm:"index;not null"`
	Email       string        `json:"email" gorm:"not null"`
	Role        WorkspaceRole `json:"role" gorm:"not null"`
	Token       string        `json:"token" gorm:"uniqueIndex;not null"`
	InvitedBy   uint          `json:"invited_by" gorm:"not null"`
	ExpiresAt   time.Time     `json:"expires_at" gorm:"not null"`
	AcceptedAt  *time.Time    `json:"accepted_at"`
	
	// Relationships
	Workspace     Workspace `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
	InvitedByUser User      `json:"invited_by_user,omitempty" gorm:"foreignKey:InvitedBy"`
}

// IsExpired checks if the invitation is expired
func (i *WorkspaceInvitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsAccepted checks if the invitation has been accepted
func (i *WorkspaceInvitation) IsAccepted() bool {
	return i.AcceptedAt != nil
}

// IsValid checks if the invitation is valid
func (i *WorkspaceInvitation) IsValid() bool {
	return !i.IsExpired() && !i.IsAccepted()
}

// Accept accepts the invitation
func (i *WorkspaceInvitation) Accept() {
	now := time.Now()
	i.AcceptedAt = &now
}

// Table represents a table within a workspace
type Table struct {
	TenantModel
	WorkspaceID uint   `json:"workspace_id" gorm:"index;not null"`
	Name        string `json:"name" gorm:"not null"`
	Slug        string `json:"slug" gorm:"not null"`
	Description string `json:"description"`
	Color       string `json:"color" gorm:"default:'#10B981'"`
	Icon        string `json:"icon"`
	Status      Status `json:"status" gorm:"default:'active'"`
	IsPublic    bool   `json:"is_public" gorm:"default:false"`
	RecordCount int64  `json:"record_count" gorm:"default:0"`
	Schema      JSON   `json:"schema" gorm:"type:jsonb"`
	Settings    JSON   `json:"settings" gorm:"type:jsonb"`
	Metadata    JSON   `json:"metadata" gorm:"type:jsonb"`
	
	// Relationships
	Workspace Workspace `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
	Fields    []Field   `json:"fields,omitempty" gorm:"foreignKey:TableID"`
	Records   []Record  `json:"records,omitempty" gorm:"foreignKey:TableID"`
	Views     []View    `json:"views,omitempty" gorm:"foreignKey:TableID"`
}

// IsActive checks if the table is active
func (t *Table) IsActive() bool {
	return t.Status == StatusActive
}

// GetFieldCount returns the number of fields in the table
func (t *Table) GetFieldCount() int {
	return len(t.Fields)
}

// GetViewCount returns the number of views in the table
func (t *Table) GetViewCount() int {
	return len(t.Views)
}

// Field represents a field/column in a table
type Field struct {
	BaseModel
	TableID     uint      `json:"table_id" gorm:"index;not null"`
	Name        string    `json:"name" gorm:"not null"`
	Type        FieldType `json:"type" gorm:"not null"`
	Description string    `json:"description"`
	Required    bool      `json:"required" gorm:"default:false"`
	Unique      bool      `json:"unique" gorm:"default:false"`
	Position    int       `json:"position" gorm:"default:0"`
	Options     JSON      `json:"options" gorm:"type:jsonb"`
	Validation  JSON      `json:"validation" gorm:"type:jsonb"`
	
	// Relationships
	Table Table `json:"table,omitempty" gorm:"foreignKey:TableID"`
}

// FieldType represents the type of a field
type FieldType string

const (
	FieldTypeText         FieldType = "text"
	FieldTypeNumber       FieldType = "number"
	FieldTypeEmail        FieldType = "email"
	FieldTypeURL          FieldType = "url"
	FieldTypePhone        FieldType = "phone"
	FieldTypeDate         FieldType = "date"
	FieldTypeDateTime     FieldType = "datetime"
	FieldTypeBoolean      FieldType = "boolean"
	FieldTypeSelect       FieldType = "select"
	FieldTypeMultiSelect  FieldType = "multiselect"
	FieldTypeAttachment   FieldType = "attachment"
	FieldTypeFormula      FieldType = "formula"
	FieldTypeLookup       FieldType = "lookup"
	FieldTypeRelation     FieldType = "relation"
	FieldTypeRollup       FieldType = "rollup"
	FieldTypeAutoNumber   FieldType = "autonumber"
	FieldTypeBarcode      FieldType = "barcode"
	FieldTypeRating       FieldType = "rating"
	FieldTypeDuration     FieldType = "duration"
	FieldTypeCurrency     FieldType = "currency"
	FieldTypePercent      FieldType = "percent"
)

// View represents a view of a table
type View struct {
	BaseModel
	TableID     uint     `json:"table_id" gorm:"index;not null"`
	Name        string   `json:"name" gorm:"not null"`
	Type        ViewType `json:"type" gorm:"not null"`
	Description string   `json:"description"`
	IsPublic    bool     `json:"is_public" gorm:"default:false"`
	Position    int      `json:"position" gorm:"default:0"`
	Config      JSON     `json:"config" gorm:"type:jsonb"`
	
	// Relationships
	Table Table `json:"table,omitempty" gorm:"foreignKey:TableID"`
}

// ViewType represents the type of view
type ViewType string

const (
	ViewTypeGrid     ViewType = "grid"
	ViewTypeForm     ViewType = "form"
	ViewTypeCalendar ViewType = "calendar"
	ViewTypeGallery  ViewType = "gallery"
	ViewTypeKanban   ViewType = "kanban"
	ViewTypeGantt    ViewType = "gantt"
	ViewTypeChart    ViewType = "chart"
)

// Record represents a record/row in a table
type Record struct {
	BaseModel
	TableID uint                   `json:"table_id" gorm:"index;not null"`
	Data    map[string]interface{} `json:"data" gorm:"type:jsonb"`
	
	// Relationships
	Table Table `json:"table,omitempty" gorm:"foreignKey:TableID"`
}

// GetFieldValue returns the value of a specific field
func (r *Record) GetFieldValue(fieldName string) interface{} {
	if r.Data == nil {
		return nil
	}
	return r.Data[fieldName]
}

// SetFieldValue sets the value of a specific field
func (r *Record) SetFieldValue(fieldName string, value interface{}) {
	if r.Data == nil {
		r.Data = make(map[string]interface{})
	}
	r.Data[fieldName] = value
}

// WorkspaceSettings represents workspace-specific settings
type WorkspaceSettings struct {
	Theme           string                 `json:"theme"`
	DefaultView     string                 `json:"default_view"`
	AllowComments   bool                   `json:"allow_comments"`
	AllowAttachments bool                  `json:"allow_attachments"`
	MaxFileSize     int64                  `json:"max_file_size"`
	TimeZone        string                 `json:"time_zone"`
	DateFormat      string                 `json:"date_format"`
	TimeFormat      string                 `json:"time_format"`
	Currency        string                 `json:"currency"`
	Language        string                 `json:"language"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
	Integrations    map[string]interface{} `json:"integrations"`
	Permissions     WorkspacePermissions   `json:"permissions"`
}

// WorkspacePermissions represents permission settings for a workspace
type WorkspacePermissions struct {
	DefaultRole        WorkspaceRole `json:"default_role"`
	AllowPublicSharing bool          `json:"allow_public_sharing"`
	RequireInvitation  bool          `json:"require_invitation"`
	AllowGuestAccess   bool          `json:"allow_guest_access"`
	MaxMembers         int           `json:"max_members"`
}