package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// JSON represents a JSON field type for GORM
type JSON map[string]interface{}

// Value implements the driver.Valuer interface for JSON
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSON
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan non-string value into JSON")
	}
	
	return json.Unmarshal(bytes, j)
}

// PaginationRequest represents a pagination request
type PaginationRequest struct {
	Page     int    `json:"page" form:"page" binding:"min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Sort     string `json:"sort" form:"sort"`
	Order    string `json:"order" form:"order" binding:"oneof=asc desc"`
	Search   string `json:"search" form:"search"`
	Filters  JSON   `json:"filters" form:"filters"`
}

// GetOffset returns the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.GetPageSize()
}

// GetPageSize returns the page size with default
func (p *PaginationRequest) GetPageSize() int {
	if p.PageSize <= 0 {
		return 20 // Default page size
	}
	if p.PageSize > 100 {
		return 100 // Max page size
	}
	return p.PageSize
}

// GetSort returns the sort field with default
func (p *PaginationRequest) GetSort() string {
	if p.Sort == "" {
		return "created_at" // Default sort field
	}
	return p.Sort
}

// GetOrder returns the sort order with default
func (p *PaginationRequest) GetOrder() string {
	if p.Order == "" {
		return "desc" // Default sort order
	}
	return p.Order
}

// PaginationResponse represents a paginated response
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewPaginationResponse creates a new paginated response
func NewPaginationResponse(data interface{}, req *PaginationRequest, total int64) *PaginationResponse {
	pageSize := req.GetPageSize()
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	page := req.Page
	if page <= 0 {
		page = 1
	}
	
	return &PaginationResponse{
		Data: data,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}
}

// SortRequest represents a sort request
type SortRequest struct {
	Field string `json:"field" binding:"required"`
	Order string `json:"order" binding:"oneof=asc desc"`
}

// FilterRequest represents a filter request
type FilterRequest struct {
	Field    string      `json:"field" binding:"required"`
	Operator string      `json:"operator" binding:"required"`
	Value    interface{} `json:"value"`
}

// ValidOperators returns valid filter operators
func ValidOperators() []string {
	return []string{
		"eq", "ne", "gt", "gte", "lt", "lte",
		"in", "not_in", "like", "not_like",
		"is_null", "is_not_null", "between",
		"starts_with", "ends_with", "contains",
		"not_contains",
	}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *APIMeta    `json:"meta,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// APIError represents an API error response
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// APIMeta represents API response metadata
type APIMeta struct {
	RequestID   string      `json:"request_id,omitempty"`
	Version     string      `json:"version,omitempty"`
	Pagination  *Pagination `json:"pagination,omitempty"`
	RateLimit   *RateLimit  `json:"rate_limit,omitempty"`
	Performance *Performance `json:"performance,omitempty"`
}

// RateLimit represents rate limit information
type RateLimit struct {
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
	Reset     int `json:"reset"`
}

// Performance represents performance metrics
type Performance struct {
	Duration    string `json:"duration"`
	QueryCount  int    `json:"query_count,omitempty"`
	CacheHits   int    `json:"cache_hits,omitempty"`
	CacheMisses int    `json:"cache_misses,omitempty"`
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: fmt.Sprintf("%d", UnixNow()),
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse(code, message string, details map[string]interface{}) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: fmt.Sprintf("%d", UnixNow()),
	}
}

// NewPaginatedResponse creates a paginated API response
func NewPaginatedResponse(data interface{}, pagination *Pagination) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		Meta:      &APIMeta{Pagination: pagination},
		Timestamp: fmt.Sprintf("%d", UnixNow()),
	}
}

// BulkRequest represents a bulk operation request
type BulkRequest struct {
	Operation string        `json:"operation" binding:"required,oneof=create update delete"`
	Items     []interface{} `json:"items" binding:"required"`
}

// BulkResponse represents a bulk operation response
type BulkResponse struct {
	Success     []BulkResult `json:"success"`
	Failed      []BulkResult `json:"failed"`
	TotalCount  int          `json:"total_count"`
	SuccessCount int          `json:"success_count"`
	FailedCount int          `json:"failed_count"`
}

// BulkResult represents a single result in a bulk operation
type BulkResult struct {
	Index   int         `json:"index"`
	ID      interface{} `json:"id,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query   string            `json:"query" binding:"required"`
	Filters map[string]string `json:"filters"`
	Sort    []SortRequest     `json:"sort"`
	PaginationRequest
}

// ExportRequest represents an export request
type ExportRequest struct {
	Format  string   `json:"format" binding:"required,oneof=csv json xlsx"`
	Fields  []string `json:"fields"`
	Filters JSON     `json:"filters"`
	Sort    string   `json:"sort"`
	Order   string   `json:"order"`
}

// ImportRequest represents an import request
type ImportRequest struct {
	Format     string            `json:"format" binding:"required,oneof=csv json xlsx"`
	Data       string            `json:"data" binding:"required"`
	Mapping    map[string]string `json:"mapping"`
	Options    ImportOptions     `json:"options"`
}

// ImportOptions represents import options
type ImportOptions struct {
	SkipHeader    bool `json:"skip_header"`
	UpdateExisting bool `json:"update_existing"`
	BulkSize      int  `json:"bulk_size"`
}

// ImportResult represents an import result
type ImportResult struct {
	TotalRows     int            `json:"total_rows"`
	SuccessCount  int            `json:"success_count"`
	FailedCount   int            `json:"failed_count"`
	Errors        []ImportError  `json:"errors"`
	Summary       ImportSummary  `json:"summary"`
}

// ImportError represents an import error
type ImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// ImportSummary represents import summary statistics
type ImportSummary struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Skipped int `json:"skipped"`
}

// NullString represents a nullable string
type NullString struct {
	sql.NullString
}

// MarshalJSON implements json.Marshaler
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON implements json.Unmarshaler
func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		ns.Valid = true
		ns.String = *s
	} else {
		ns.Valid = false
	}
	return nil
}

// NullInt64 represents a nullable int64
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON implements json.Marshaler
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ni.Int64)
}

// UnmarshalJSON implements json.Unmarshaler
func (ni *NullInt64) UnmarshalJSON(data []byte) error {
	var i *int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	if i != nil {
		ni.Valid = true
		ni.Int64 = *i
	} else {
		ni.Valid = false
	}
	return nil
}

// UnixNow returns current Unix timestamp (placeholder implementation)
func UnixNow() int64 {
	// This would typically use time.Now().Unix()
	// Using placeholder for consistency with errors package
	return 1640995200 // 2022-01-01 00:00:00 UTC
}