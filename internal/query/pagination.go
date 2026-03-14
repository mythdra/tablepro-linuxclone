// Package query provides query execution services with timeout, cancellation, and result streaming.
package query

import (
	"regexp"
	"strings"
)

const (
	// DefaultPageSize is the default number of rows per page (100)
	DefaultPageSize = 100

	// MinPageSize is the minimum allowed page size (10)
	MinPageSize = 10

	// MaxPageSize is the maximum allowed page size (10000)
	MaxPageSize = 10000

	// SmallTableThreshold is the row count threshold for exact vs estimated count
	SmallTableThreshold = 10000
)

// PaginationService manages query pagination for SQL and NoSQL databases.
// Provides LIMIT/OFFSET calculation, count query generation, and page size validation.
type PaginationService struct {
	defaultPageSize int
	maxPageSize     int
	minPageSize     int
}

// PaginationContext holds pagination state and navigation info.
// Used by frontend to render pagination controls and calculate offsets.
type PaginationContext struct {
	// Page is the current page number (1-indexed)
	Page int `json:"page"`

	// PageSize is the number of rows per page
	PageSize int `json:"pageSize"`

	// TotalCount is the total number of rows
	TotalCount int64 `json:"totalCount"`

	// TotalPages is the total number of pages
	TotalPages int `json:"totalPages"`

	// HasNext indicates if there's a next page
	HasNext bool `json:"hasNext"`

	// HasPrev indicates if there's a previous page
	HasPrev bool `json:"hasPrev"`

	// Offset is the calculated OFFSET value for SQL queries
	Offset int `json:"offset"`

	// IsExact indicates if the count is exact or estimated
	IsExact bool `json:"isExact"`
}

// CountResult holds the result of a count operation.
// Includes flag indicating whether count is exact or estimated.
type CountResult struct {
	// Count is the total row count
	Count int64 `json:"count"`

	// IsExact indicates if the count is exact or estimated
	IsExact bool `json:"isExact"`
}

// NewPaginationService creates a new PaginationService with default configuration.
// Uses DefaultPageSize, MinPageSize, and MaxPageSize constants.
func NewPaginationService() *PaginationService {
	return &PaginationService{
		defaultPageSize: DefaultPageSize,
		maxPageSize:     MaxPageSize,
		minPageSize:     MinPageSize,
	}
}

// ValidatePageSize validates and clamps page size to allowed range.
// Returns MinPageSize if input is too small, MaxPageSize if too large.
func (p *PaginationService) ValidatePageSize(pageSize int) int {
	if pageSize < p.minPageSize {
		return p.minPageSize
	}
	if pageSize > p.maxPageSize {
		return p.maxPageSize
	}
	return pageSize
}

// ApplySQLOffset adds LIMIT and OFFSET clauses to a SQL query.
// Preserves ORDER BY clauses, removes existing LIMIT/OFFSET.
// Returns query with pagination applied for server-side fetching.
func (p *PaginationService) ApplySQLOffset(query string, page int, pageSize int) string {
	// Validate inputs
	if page < 1 {
		page = 1
	}
	if pageSize < p.minPageSize {
		pageSize = p.minPageSize
	}
	if pageSize > p.maxPageSize {
		pageSize = p.maxPageSize
	}

	offset := p.CalculateOffset(page, pageSize)

	// Clean up query: remove trailing semicolon and existing LIMIT/OFFSET
	cleanedQuery := p.cleanQuery(query, true) // preserve ORDER BY

	// Append LIMIT and OFFSET
	return cleanedQuery + " LIMIT " + itoa(pageSize) + " OFFSET " + itoa(offset)
}

// cleanQuery removes trailing semicolons and existing LIMIT/OFFSET clauses.
// Optionally removes ORDER BY if preserveOrderBy is false.
func (p *PaginationService) cleanQuery(query string, preserveOrderBy bool) string {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Remove trailing semicolon
	query = strings.TrimSuffix(query, ";")

	// Remove ORDER BY clause if not preserving
	if !preserveOrderBy {
		orderByRegex := regexp.MustCompile(`(?i)\s+ORDER\s+BY\s+[^;]+$`)
		query = orderByRegex.ReplaceAllString(query, "")
	}

	// Remove existing LIMIT clause (case-insensitive)
	limitRegex := regexp.MustCompile(`(?i)\s+LIMIT\s+\d+(\s+OFFSET\s+\d+)?$`)
	query = limitRegex.ReplaceAllString(query, "")

	// Remove existing OFFSET clause (case-insensitive)
	offsetRegex := regexp.MustCompile(`(?i)\s+OFFSET\s+\d+$`)
	query = offsetRegex.ReplaceAllString(query, "")

	return strings.TrimSpace(query)
}

// CalculateOffset calculates the OFFSET value for a given page and page size.
// Formula: (page - 1) * pageSize. Page 1 returns offset 0.
func (p *PaginationService) CalculateOffset(page int, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}

// CalculateTotalPages calculates total pages from total count and page size.
// Returns 0 for empty results, rounds up for partial pages.
func (p *PaginationService) CalculateTotalPages(total int64, pageSize int) int {
	if total == 0 {
		return 0
	}
	pages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		pages++
	}
	return pages
}

// GetCountQuery generates a COUNT(*) query from a SELECT query.
// Extracts table name and WHERE clause, strips other clauses.
// Falls back to wrapping original query for complex SELECTs.
func (p *PaginationService) GetCountQuery(query string) string {
	cleaned := p.cleanQuery(query, false)

	// Find the FROM clause and extract table name + WHERE clause
	fromRegex := regexp.MustCompile(`(?i)\s+FROM\s+([^\s]+)(.*)$`)
	matches := fromRegex.FindStringSubmatch(cleaned)

	if len(matches) >= 2 {
		tableName := matches[1]
		rest := ""
		if len(matches) >= 3 {
			rest = matches[2]
		}

		// Extract WHERE clause if present (stop at GROUP BY, HAVING, ORDER BY, LIMIT, OFFSET)
		whereRegex := regexp.MustCompile(`(?i)\s+WHERE\s+(.+?)(?:\s+(?:GROUP\s+BY|HAVING|ORDER\s+BY|LIMIT|OFFSET)|$)`)
		whereMatches := whereRegex.FindStringSubmatch(rest)

		if len(whereMatches) >= 2 {
			return "SELECT COUNT(*) FROM " + tableName + " WHERE " + strings.TrimSpace(whereMatches[1])
		}

		return "SELECT COUNT(*) FROM " + tableName
	}

	// Fallback: wrap original query
	return "SELECT COUNT(*) FROM (" + cleaned + ") AS subquery"
}

// EstimateCount returns count information, using exact count for small tables.
// Tables with rowCount < SmallTableThreshold are considered exact.
func (p *PaginationService) EstimateCount(count int64) CountResult {
	return CountResult{
		Count:   count,
		IsExact: count < SmallTableThreshold,
	}
}

// NewPaginationContext creates a new PaginationContext with calculated values.
// Computes totalPages, offset, hasNext, hasPrev, and isExact flags.
func NewPaginationContext(page int, pageSize int, totalCount int64) *PaginationContext {
	ps := NewPaginationService()
	totalPages := ps.CalculateTotalPages(totalCount, pageSize)
	offset := ps.CalculateOffset(page, pageSize)

	return &PaginationContext{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		Offset:     offset,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
		IsExact:    totalCount < SmallTableThreshold,
	}
}

// ApplyMongoCursor applies cursor-based pagination for MongoDB.
// For MongoDB, we use skip/limit pattern.
// Returns skip and limit values for MongoDB query.
func (p *PaginationService) ApplyMongoCursor(page int, pageSize int) (skip int, limit int) {
	if page < 1 {
		page = 1
	}
	pageSize = p.ValidatePageSize(pageSize)
	skip = (page - 1) * pageSize
	limit = pageSize
	return skip, limit
}

// ScanRedisCount returns the COUNT parameter for Redis SCAN command.
// Validates page size within allowed range.
func (p *PaginationService) ScanRedisCount(pageSize int) int {
	return p.ValidatePageSize(pageSize)
}

// itoa converts int to string without importing strconv.
// Used to avoid import cycle in pagination package.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := false
	if n < 0 {
		negative = true
		n = -n
	}

	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	if negative {
		digits = append(digits, '-')
	}

	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	return string(digits)
}
