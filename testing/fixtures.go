package testing

import (
	"time"

	"github.com/Reg-Kris/pyairtable-go-shared/models"
)

// TestFixtures provides test data fixtures
type TestFixtures struct{}

// NewTestFixtures creates a new test fixtures instance
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{}
}

// CreateTestUser creates a test user
func (f *TestFixtures) CreateTestUser(overrides ...func(*models.User)) *models.User {
	user := &models.User{
		TenantModel: models.TenantModel{
			AuditableModel: models.AuditableModel{
				BaseModel: models.BaseModel{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CreatedBy: 1,
				UpdatedBy: 1,
				Version:   1,
			},
			TenantID: 1,
		},
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz",
		Status:       models.StatusActive,
		EmailVerified: true,
		Timezone:     "UTC",
		Language:     "en",
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(user)
	}
	
	return user
}

// CreateTestTenant creates a test tenant
func (f *TestFixtures) CreateTestTenant(overrides ...func(*models.Tenant)) *models.Tenant {
	tenant := &models.Tenant{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Tenant",
		Slug:        "test-tenant",
		Email:       "admin@testtenant.com",
		Country:     "US",
		Status:      models.StatusActive,
		PlanType:    models.PlanTypeFree,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(tenant)
	}
	
	return tenant
}

// CreateTestWorkspace creates a test workspace
func (f *TestFixtures) CreateTestWorkspace(overrides ...func(*models.Workspace)) *models.Workspace {
	workspace := &models.Workspace{
		TenantModel: models.TenantModel{
			AuditableModel: models.AuditableModel{
				BaseModel: models.BaseModel{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CreatedBy: 1,
				UpdatedBy: 1,
				Version:   1,
			},
			TenantID: 1,
		},
		Name:        "Test Workspace",
		Slug:        "test-workspace",
		Description: "A test workspace",
		Status:      models.StatusActive,
		IsPublic:    false,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(workspace)
	}
	
	return workspace
}

// CreateTestTable creates a test table
func (f *TestFixtures) CreateTestTable(overrides ...func(*models.Table)) *models.Table {
	table := &models.Table{
		TenantModel: models.TenantModel{
			AuditableModel: models.AuditableModel{
				BaseModel: models.BaseModel{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				CreatedBy: 1,
				UpdatedBy: 1,
				Version:   1,
			},
			TenantID: 1,
		},
		WorkspaceID: 1,
		Name:        "Test Table",
		Slug:        "test-table",
		Description: "A test table",
		Status:      models.StatusActive,
		IsPublic:    false,
		RecordCount: 0,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(table)
	}
	
	return table
}

// CreateTestField creates a test field
func (f *TestFixtures) CreateTestField(overrides ...func(*models.Field)) *models.Field {
	field := &models.Field{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TableID:     1,
		Name:        "Test Field",
		Type:        models.FieldTypeText,
		Description: "A test field",
		Required:    false,
		Unique:      false,
		Position:    0,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(field)
	}
	
	return field
}

// CreateTestRecord creates a test record
func (f *TestFixtures) CreateTestRecord(overrides ...func(*models.Record)) *models.Record {
	record := &models.Record{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TableID: 1,
		Data: map[string]interface{}{
			"name": "Test Record",
			"description": "A test record",
		},
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(record)
	}
	
	return record
}

// CreateTestRole creates a test role
func (f *TestFixtures) CreateTestRole(overrides ...func(*models.Role)) *models.Role {
	role := &models.Role{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "test_role",
		Description: "A test role",
		IsSystem:    false,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(role)
	}
	
	return role
}

// CreateTestPermission creates a test permission
func (f *TestFixtures) CreateTestPermission(overrides ...func(*models.Permission)) *models.Permission {
	permission := &models.Permission{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "test_permission",
		Description: "A test permission",
		Resource:    "test_resource",
		Action:      "read",
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(permission)
	}
	
	return permission
}

// CreateTestSession creates a test session
func (f *TestFixtures) CreateTestSession(overrides ...func(*models.Session)) *models.Session {
	session := &models.Session{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:    1,
		Token:     "test_token_123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IPAddress: "127.0.0.1",
		UserAgent: "Test Agent",
		IsActive:  true,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(session)
	}
	
	return session
}

// CreateTestAPIKey creates a test API key
func (f *TestFixtures) CreateTestAPIKey(overrides ...func(*models.APIKey)) *models.APIKey {
	apiKey := &models.APIKey{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:     1,
		Name:       "Test API Key",
		KeyHash:    "hashed_api_key",
		Prefix:     "pk_test",
		IsActive:   true,
		Scopes:     []string{"read", "write"},
		UsageCount: 0,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(apiKey)
	}
	
	return apiKey
}

// CreateTestWorkspaceMember creates a test workspace member
func (f *TestFixtures) CreateTestWorkspaceMember(overrides ...func(*models.WorkspaceMember)) *models.WorkspaceMember {
	member := &models.WorkspaceMember{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		WorkspaceID: 1,
		UserID:      1,
		Role:        models.WorkspaceRoleEditor,
		JoinedAt:    time.Now(),
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(member)
	}
	
	return member
}

// CreateTestView creates a test view
func (f *TestFixtures) CreateTestView(overrides ...func(*models.View)) *models.View {
	view := &models.View{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TableID:     1,
		Name:        "Test View",
		Type:        models.ViewTypeGrid,
		Description: "A test view",
		IsPublic:    false,
		Position:    0,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(view)
	}
	
	return view
}

// CreateTestTenantQuota creates a test tenant quota
func (f *TestFixtures) CreateTestTenantQuota(overrides ...func(*models.TenantQuota)) *models.TenantQuota {
	quota := &models.TenantQuota{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		TenantID:     1,
		ResourceType: "users",
		Used:         0,
		Limit:        100,
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(quota)
	}
	
	return quota
}

// CreateTestPasswordResetToken creates a test password reset token
func (f *TestFixtures) CreateTestPasswordResetToken(overrides ...func(*models.PasswordResetToken)) *models.PasswordResetToken {
	token := &models.PasswordResetToken{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:    1,
		Token:     "reset_token_123",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	
	// Apply overrides
	for _, override := range overrides {
		override(token)
	}
	
	return token
}

// CreateExpiredSession creates an expired session for testing
func (f *TestFixtures) CreateExpiredSession() *models.Session {
	return f.CreateTestSession(func(s *models.Session) {
		s.ExpiresAt = time.Now().Add(-1 * time.Hour)
	})
}

// CreateInactiveUser creates an inactive user for testing
func (f *TestFixtures) CreateInactiveUser() *models.User {
	return f.CreateTestUser(func(u *models.User) {
		u.Status = models.StatusInactive
		u.EmailVerified = false
	})
}

// CreateAdminUser creates an admin user for testing
func (f *TestFixtures) CreateAdminUser() *models.User {
	return f.CreateTestUser(func(u *models.User) {
		u.Email = "admin@example.com"
		u.FirstName = "Admin"
		u.LastName = "User"
	})
}

// CreateMultipleUsers creates multiple test users
func (f *TestFixtures) CreateMultipleUsers(count int) []*models.User {
	users := make([]*models.User, count)
	for i := 0; i < count; i++ {
		users[i] = f.CreateTestUser(func(u *models.User) {
			u.ID = uint(i + 1)
			u.Email = fmt.Sprintf("user%d@example.com", i+1)
			u.FirstName = fmt.Sprintf("User%d", i+1)
		})
	}
	return users
}

// CreateMultipleTenants creates multiple test tenants
func (f *TestFixtures) CreateMultipleTenants(count int) []*models.Tenant {
	tenants := make([]*models.Tenant, count)
	for i := 0; i < count; i++ {
		tenants[i] = f.CreateTestTenant(func(t *models.Tenant) {
			t.ID = uint(i + 1)
			t.Name = fmt.Sprintf("Test Tenant %d", i+1)
			t.Slug = fmt.Sprintf("test-tenant-%d", i+1)
			t.Email = fmt.Sprintf("admin%d@testtenant.com", i+1)
		})
	}
	return tenants
}

// SeedDatabase seeds the database with test data
func (f *TestFixtures) SeedDatabase(db *TestDB) error {
	// Create test tenant
	tenant := f.CreateTestTenant()
	if err := db.Create(tenant).Error; err != nil {
		return err
	}
	
	// Create test user
	user := f.CreateTestUser()
	if err := db.Create(user).Error; err != nil {
		return err
	}
	
	// Create test workspace
	workspace := f.CreateTestWorkspace()
	if err := db.Create(workspace).Error; err != nil {
		return err
	}
	
	// Create test table
	table := f.CreateTestTable()
	if err := db.Create(table).Error; err != nil {
		return err
	}
	
	// Create test fields
	nameField := f.CreateTestField(func(field *models.Field) {
		field.Name = "Name"
		field.Type = models.FieldTypeText
		field.Required = true
	})
	if err := db.Create(nameField).Error; err != nil {
		return err
	}
	
	emailField := f.CreateTestField(func(field *models.Field) {
		field.ID = 2
		field.Name = "Email"
		field.Type = models.FieldTypeEmail
		field.Position = 1
	})
	if err := db.Create(emailField).Error; err != nil {
		return err
	}
	
	return nil
}