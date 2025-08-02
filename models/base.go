// Package models provides shared data models for all microservices
package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// SoftDeletable interface for models that support soft deletion
type SoftDeletable interface {
	IsDeleted() bool
}

// IsDeleted checks if the model is soft deleted
func (b *BaseModel) IsDeleted() bool {
	return b.DeletedAt.Valid
}

// Timestampable interface for models with timestamp fields
type Timestampable interface {
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// SetCreatedAt sets the created at timestamp
func (b *BaseModel) SetCreatedAt(t time.Time) {
	b.CreatedAt = t
}

// SetUpdatedAt sets the updated at timestamp
func (b *BaseModel) SetUpdatedAt(t time.Time) {
	b.UpdatedAt = t
}

// GetCreatedAt returns the created at timestamp
func (b *BaseModel) GetCreatedAt() time.Time {
	return b.CreatedAt
}

// GetUpdatedAt returns the updated at timestamp
func (b *BaseModel) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

// AuditableModel contains fields for audit logging
type AuditableModel struct {
	BaseModel
	CreatedBy uint   `json:"created_by" gorm:"index"`
	UpdatedBy uint   `json:"updated_by" gorm:"index"`
	Version   int64  `json:"version" gorm:"default:1"`
}

// BeforeUpdate increments version for optimistic locking
func (a *AuditableModel) BeforeUpdate(tx *gorm.DB) error {
	a.Version++
	return nil
}

// TenantModel contains fields for multi-tenant models
type TenantModel struct {
	AuditableModel
	TenantID uint `json:"tenant_id" gorm:"index;not null"`
}

// Status represents the status of an entity
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusPending  Status = "pending"
	StatusArchived Status = "archived"
	StatusDeleted  Status = "deleted"
)

// IsActive checks if the status is active
func (s Status) IsActive() bool {
	return s == StatusActive
}

// IsInactive checks if the status is inactive
func (s Status) IsInactive() bool {
	return s == StatusInactive
}

// IsPending checks if the status is pending
func (s Status) IsPending() bool {
	return s == StatusPending
}

// IsArchived checks if the status is archived
func (s Status) IsArchived() bool {
	return s == StatusArchived
}

// IsDeleted checks if the status is deleted
func (s Status) IsDeleted() bool {
	return s == StatusDeleted
}