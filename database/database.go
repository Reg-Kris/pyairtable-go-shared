// Package database provides database connection and management utilities using GORM
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Reg-Kris/pyairtable-go-shared/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB holds the database connection
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	return &DB{DB: db}, nil
}

// Health checks the database connection health
func (db *DB) Health() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	return sqlDB.Ping()
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	return sqlDB.Close()
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(fn func(*gorm.DB) error) error {
	return db.DB.Transaction(fn)
}

// GetStats returns database connection statistics
func (db *DB) GetStats() (map[string]interface{}, error) {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	stats := sqlDB.Stats()
	
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}, nil
}

// Migrate runs database migrations for the given models
func (db *DB) Migrate(models ...interface{}) error {
	return db.AutoMigrate(models...)
}

// Repository provides generic repository operations
type Repository[T any] struct {
	db *DB
}

// NewRepository creates a new repository for the given type
func NewRepository[T any](db *DB) *Repository[T] {
	return &Repository[T]{db: db}
}

// Create creates a new record
func (r *Repository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

// GetByID retrieves a record by ID
func (r *Repository[T]) GetByID(id uint) (*T, error) {
	var entity T
	err := r.db.First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Update updates a record
func (r *Repository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

// Delete deletes a record by ID
func (r *Repository[T]) Delete(id uint) error {
	var entity T
	return r.db.Delete(&entity, id).Error
}

// List retrieves records with pagination
func (r *Repository[T]) List(offset, limit int) ([]T, error) {
	var entities []T
	err := r.db.Offset(offset).Limit(limit).Find(&entities).Error
	return entities, err
}

// Count returns the total count of records
func (r *Repository[T]) Count() (int64, error) {
	var count int64
	var entity T
	err := r.db.Model(&entity).Count(&count).Error
	return count, err
}

// FindWhere finds records matching the given condition
func (r *Repository[T]) FindWhere(condition string, args ...interface{}) ([]T, error) {
	var entities []T
	err := r.db.Where(condition, args...).Find(&entities).Error
	return entities, err
}

// FirstWhere finds the first record matching the given condition
func (r *Repository[T]) FirstWhere(condition string, args ...interface{}) (*T, error) {
	var entity T
	err := r.db.Where(condition, args...).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}