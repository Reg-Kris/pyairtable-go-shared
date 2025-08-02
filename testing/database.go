// Package testing provides testing utilities for Go microservices
package testing

import (
	"fmt"
	"testing"

	"github.com/Reg-Kris/pyairtable-go-shared/config"
	"github.com/Reg-Kris/pyairtable-go-shared/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB represents a test database instance
type TestDB struct {
	*database.DB
	cleanup func()
}

// NewTestDB creates a new test database instance
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()
	
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	testDB := &TestDB{
		DB: &database.DB{DB: db},
		cleanup: func() {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		},
	}
	
	return testDB
}

// NewTestDBWithConfig creates a test database with custom configuration
func NewTestDBWithConfig(t *testing.T, cfg *config.DatabaseConfig) *TestDB {
	t.Helper()
	
	if cfg == nil {
		return NewTestDB(t)
	}
	
	// For testing, always use SQLite in memory
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	testDB := &TestDB{
		DB: &database.DB{DB: db},
		cleanup: func() {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		},
	}
	
	return testDB
}

// Cleanup cleans up the test database
func (tdb *TestDB) Cleanup() {
	if tdb.cleanup != nil {
		tdb.cleanup()
	}
}

// Reset resets the database by dropping and recreating all tables
func (tdb *TestDB) Reset(models ...interface{}) error {
	// Drop all tables
	for _, model := range models {
		if err := tdb.Migrator().DropTable(model); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}
	
	// Recreate tables
	return tdb.AutoMigrate(models...)
}

// Seed seeds the database with test data
func (tdb *TestDB) Seed(data ...interface{}) error {
	for _, item := range data {
		if err := tdb.Create(item).Error; err != nil {
			return fmt.Errorf("failed to seed data: %w", err)
		}
	}
	return nil
}

// BeginTransaction starts a transaction for testing
func (tdb *TestDB) BeginTransaction() *gorm.DB {
	return tdb.DB.Begin()
}

// WithTransaction executes a function within a transaction and rolls it back
func (tdb *TestDB) WithTransaction(fn func(*gorm.DB) error) error {
	tx := tdb.BeginTransaction()
	defer tx.Rollback()
	
	return fn(tx)
}

// AssertCount asserts that a table has the expected number of records
func (tdb *TestDB) AssertCount(t *testing.T, model interface{}, expected int64) {
	t.Helper()
	
	var count int64
	if err := tdb.Model(model).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}
	
	if count != expected {
		t.Errorf("Expected %d records, got %d", expected, count)
	}
}

// AssertExists asserts that a record exists
func (tdb *TestDB) AssertExists(t *testing.T, model interface{}, condition string, args ...interface{}) {
	t.Helper()
	
	var count int64
	query := tdb.Model(model)
	if condition != "" {
		query = query.Where(condition, args...)
	}
	
	if err := query.Count(&count).Error; err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	
	if count == 0 {
		t.Errorf("Expected record to exist with condition: %s", condition)
	}
}

// AssertNotExists asserts that a record does not exist
func (tdb *TestDB) AssertNotExists(t *testing.T, model interface{}, condition string, args ...interface{}) {
	t.Helper()
	
	var count int64
	query := tdb.Model(model)
	if condition != "" {
		query = query.Where(condition, args...)
	}
	
	if err := query.Count(&count).Error; err != nil {
		t.Fatalf("Failed to check non-existence: %v", err)
	}
	
	if count > 0 {
		t.Errorf("Expected record not to exist with condition: %s", condition)
	}
}

// CreateTestRecord creates a test record and returns its ID
func (tdb *TestDB) CreateTestRecord(t *testing.T, model interface{}) uint {
	t.Helper()
	
	if err := tdb.Create(model).Error; err != nil {
		t.Fatalf("Failed to create test record: %v", err)
	}
	
	// Use reflection to get ID field
	if idGetter, ok := model.(interface{ GetID() uint }); ok {
		return idGetter.GetID()
	}
	
	return 0
}

// TruncateTable truncates a table
func (tdb *TestDB) TruncateTable(t *testing.T, model interface{}) {
	t.Helper()
	
	if err := tdb.Unscoped().Delete(model, "1 = 1").Error; err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}
}

// GetTableName returns the table name for a model
func (tdb *TestDB) GetTableName(model interface{}) string {
	stmt := &gorm.Statement{DB: tdb.DB.DB}
	stmt.Parse(model)
	return stmt.Schema.Table
}

// ExecuteRawSQL executes raw SQL for testing
func (tdb *TestDB) ExecuteRawSQL(t *testing.T, sql string, args ...interface{}) {
	t.Helper()
	
	if err := tdb.Exec(sql, args...).Error; err != nil {
		t.Fatalf("Failed to execute raw SQL: %v", err)
	}
}

// MockRepository creates a mock repository for testing
type MockRepository[T any] struct {
	items   map[uint]*T
	counter uint
}

// NewMockRepository creates a new mock repository
func NewMockRepository[T any]() *MockRepository[T] {
	return &MockRepository[T]{
		items:   make(map[uint]*T),
		counter: 0,
	}
}

// Create mocks creating a record
func (r *MockRepository[T]) Create(entity *T) error {
	r.counter++
	if idSetter, ok := interface{}(entity).(interface{ SetID(uint) }); ok {
		idSetter.SetID(r.counter)
	}
	r.items[r.counter] = entity
	return nil
}

// GetByID mocks retrieving a record by ID
func (r *MockRepository[T]) GetByID(id uint) (*T, error) {
	if item, exists := r.items[id]; exists {
		return item, nil
	}
	return nil, gorm.ErrRecordNotFound
}

// Update mocks updating a record
func (r *MockRepository[T]) Update(entity *T) error {
	if idGetter, ok := interface{}(entity).(interface{ GetID() uint }); ok {
		id := idGetter.GetID()
		if _, exists := r.items[id]; exists {
			r.items[id] = entity
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

// Delete mocks deleting a record by ID
func (r *MockRepository[T]) Delete(id uint) error {
	if _, exists := r.items[id]; exists {
		delete(r.items, id)
		return nil
	}
	return gorm.ErrRecordNotFound
}

// List mocks listing records with pagination
func (r *MockRepository[T]) List(offset, limit int) ([]T, error) {
	var result []T
	count := 0
	
	for _, item := range r.items {
		if count >= offset {
			result = append(result, *item)
			if len(result) >= limit {
				break
			}
		}
		count++
	}
	
	return result, nil
}

// Count mocks counting records
func (r *MockRepository[T]) Count() (int64, error) {
	return int64(len(r.items)), nil
}

// FindWhere mocks finding records with a condition
func (r *MockRepository[T]) FindWhere(condition string, args ...interface{}) ([]T, error) {
	// Simple mock implementation - returns all items
	var result []T
	for _, item := range r.items {
		result = append(result, *item)
	}
	return result, nil
}

// FirstWhere mocks finding the first record with a condition
func (r *MockRepository[T]) FirstWhere(condition string, args ...interface{}) (*T, error) {
	for _, item := range r.items {
		return item, nil // Return first item for simplicity
	}
	return nil, gorm.ErrRecordNotFound
}

// Clear clears all data from the mock repository
func (r *MockRepository[T]) Clear() {
	r.items = make(map[uint]*T)
	r.counter = 0
}

// DatabaseTestSuite provides a test suite for database operations
type DatabaseTestSuite struct {
	DB *TestDB
	T  *testing.T
}

// NewDatabaseTestSuite creates a new database test suite
func NewDatabaseTestSuite(t *testing.T) *DatabaseTestSuite {
	return &DatabaseTestSuite{
		DB: NewTestDB(t),
		T:  t,
	}
}

// SetUp sets up the test suite
func (suite *DatabaseTestSuite) SetUp() {
	// Override this method in specific test suites
}

// TearDown tears down the test suite
func (suite *DatabaseTestSuite) TearDown() {
	suite.DB.Cleanup()
}

// BeforeEach runs before each test
func (suite *DatabaseTestSuite) BeforeEach() {
	// Override this method in specific test suites
}

// AfterEach runs after each test
func (suite *DatabaseTestSuite) AfterEach() {
	// Override this method in specific test suites
}