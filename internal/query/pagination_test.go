package query

import (
	"testing"
)

func TestPaginationService_New(t *testing.T) {
	t.Run("creates service with defaults", func(t *testing.T) {
		ps := NewPaginationService()

		if ps.defaultPageSize != 100 {
			t.Errorf("Expected defaultPageSize 100, got %d", ps.defaultPageSize)
		}
		if ps.maxPageSize != 10000 {
			t.Errorf("Expected maxPageSize 10000, got %d", ps.maxPageSize)
		}
		if ps.minPageSize != 10 {
			t.Errorf("Expected minPageSize 10, got %d", ps.minPageSize)
		}
	})
}

func TestPaginationService_ValidatePageSize(t *testing.T) {
	ps := NewPaginationService()

	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"negative becomes min", -1, 10},
		{"zero becomes min", 0, 10},
		{"below min clamped", 5, 10},
		{"at min", 10, 10},
		{"within range", 100, 100},
		{"at max", 10000, 10000},
		{"above max clamped", 50000, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ps.ValidatePageSize(tt.input)
			if result != tt.expected {
				t.Errorf("ValidatePageSize(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPaginationService_ApplySQLOffset(t *testing.T) {
	ps := NewPaginationService()

	tests := []struct {
		name     string
		query    string
		page     int
		pageSize int
		expected string
	}{
		{
			name:     "first page",
			query:    "SELECT * FROM users",
			page:     1,
			pageSize: 100,
			expected: "SELECT * FROM users LIMIT 100 OFFSET 0",
		},
		{
			name:     "second page",
			query:    "SELECT * FROM users",
			page:     2,
			pageSize: 100,
			expected: "SELECT * FROM users LIMIT 100 OFFSET 100",
		},
		{
			name:     "page 3 with size 50",
			query:    "SELECT * FROM users",
			page:     3,
			pageSize: 50,
			expected: "SELECT * FROM users LIMIT 50 OFFSET 100",
		},
		{
			name:     "query with semicolon",
			query:    "SELECT * FROM users;",
			page:     1,
			pageSize: 100,
			expected: "SELECT * FROM users LIMIT 100 OFFSET 0",
		},
		{
			name:     "query with WHERE clause",
			query:    "SELECT * FROM users WHERE active = true",
			page:     1,
			pageSize: 100,
			expected: "SELECT * FROM users WHERE active = true LIMIT 100 OFFSET 0",
		},
		{
			name:     "query with ORDER BY",
			query:    "SELECT * FROM users ORDER BY created_at DESC",
			page:     2,
			pageSize: 50,
			expected: "SELECT * FROM users ORDER BY created_at DESC LIMIT 50 OFFSET 50",
		},
		{
			name:     "page 0 treated as page 1",
			query:    "SELECT * FROM users",
			page:     0,
			pageSize: 100,
			expected: "SELECT * FROM users LIMIT 100 OFFSET 0",
		},
		{
			name:     "negative page treated as page 1",
			query:    "SELECT * FROM users",
			page:     -5,
			pageSize: 100,
			expected: "SELECT * FROM users LIMIT 100 OFFSET 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ps.ApplySQLOffset(tt.query, tt.page, tt.pageSize)
			if result != tt.expected {
				t.Errorf("ApplySQLOffset(%q, %d, %d) = %q, want %q",
					tt.query, tt.page, tt.pageSize, result, tt.expected)
			}
		})
	}
}

func TestPaginationService_CalculateOffset(t *testing.T) {
	ps := NewPaginationService()

	tests := []struct {
		name     string
		page     int
		pageSize int
		expected int
	}{
		{"page 1", 1, 100, 0},
		{"page 2", 2, 100, 100},
		{"page 3 size 50", 3, 50, 100},
		{"page 0", 0, 100, 0},
		{"negative page", -1, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ps.CalculateOffset(tt.page, tt.pageSize)
			if result != tt.expected {
				t.Errorf("CalculateOffset(%d, %d) = %d, want %d",
					tt.page, tt.pageSize, result, tt.expected)
			}
		})
	}
}

func TestPaginationService_CalculateTotalPages(t *testing.T) {
	ps := NewPaginationService()

	tests := []struct {
		name     string
		total    int64
		pageSize int
		expected int
	}{
		{"exact division", 100, 10, 10},
		{"with remainder", 105, 10, 11},
		{"zero total", 0, 10, 0},
		{"less than page size", 5, 10, 1},
		{"exactly one page", 10, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ps.CalculateTotalPages(tt.total, tt.pageSize)
			if result != tt.expected {
				t.Errorf("CalculateTotalPages(%d, %d) = %d, want %d",
					tt.total, tt.pageSize, result, tt.expected)
			}
		})
	}
}

func TestPaginationService_GetCountQuery(t *testing.T) {
	ps := NewPaginationService()

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "simple select",
			query:    "SELECT * FROM users",
			expected: "SELECT COUNT(*) FROM users",
		},
		{
			name:     "select with WHERE",
			query:    "SELECT * FROM users WHERE active = true",
			expected: "SELECT COUNT(*) FROM users WHERE active = true",
		},
		{
			name:     "select with ORDER BY removed",
			query:    "SELECT * FROM users ORDER BY created_at",
			expected: "SELECT COUNT(*) FROM users",
		},
		{
			name:     "select with LIMIT removed",
			query:    "SELECT * FROM users LIMIT 100",
			expected: "SELECT COUNT(*) FROM users",
		},
		{
			name:     "select with OFFSET removed",
			query:    "SELECT * FROM users OFFSET 50",
			expected: "SELECT COUNT(*) FROM users",
		},
		{
			name:     "select with semicolon",
			query:    "SELECT * FROM users;",
			expected: "SELECT COUNT(*) FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ps.GetCountQuery(tt.query)
			if result != tt.expected {
				t.Errorf("GetCountQuery(%q) = %q, want %q",
					tt.query, result, tt.expected)
			}
		})
	}
}

func TestPaginationService_EstimateCount(t *testing.T) {
	ps := NewPaginationService()

	t.Run("small table exact count", func(t *testing.T) {
		result := ps.EstimateCount(5000)
		if !result.IsExact {
			t.Error("Expected IsExact=true for small table")
		}
		if result.Count != 5000 {
			t.Errorf("Expected count 5000, got %d", result.Count)
		}
	})

	t.Run("large table estimate", func(t *testing.T) {
		result := ps.EstimateCount(50000)
		if result.IsExact {
			t.Error("Expected IsExact=false for large table")
		}
		if result.Count != 50000 {
			t.Errorf("Expected count 50000, got %d", result.Count)
		}
	})

	t.Run("exactly threshold", func(t *testing.T) {
		result := ps.EstimateCount(10000)
		if result.IsExact {
			t.Error("Expected IsExact=false for exactly threshold")
		}
	})
}

func TestPaginationContext(t *testing.T) {
	t.Run("creates context with correct values", func(t *testing.T) {
		ctx := NewPaginationContext(1, 100, 5000)

		if ctx.Page != 1 {
			t.Errorf("Expected Page 1, got %d", ctx.Page)
		}
		if ctx.PageSize != 100 {
			t.Errorf("Expected PageSize 100, got %d", ctx.PageSize)
		}
		if ctx.TotalCount != 5000 {
			t.Errorf("Expected TotalCount 5000, got %d", ctx.TotalCount)
		}
		if ctx.TotalPages != 50 {
			t.Errorf("Expected TotalPages 50, got %d", ctx.TotalPages)
		}
		if ctx.HasNext != true {
			t.Error("Expected HasNext=true")
		}
		if ctx.HasPrev != false {
			t.Error("Expected HasPrev=false")
		}
	})

	t.Run("last page detection", func(t *testing.T) {
		ctx := NewPaginationContext(50, 100, 5000)
		if ctx.HasNext != false {
			t.Error("Expected HasNext=false on last page")
		}
		if ctx.HasPrev != true {
			t.Error("Expected HasPrev=true on last page")
		}
	})

	t.Run("empty result", func(t *testing.T) {
		ctx := NewPaginationContext(1, 100, 0)
		if ctx.TotalPages != 0 {
			t.Errorf("Expected TotalPages 0, got %d", ctx.TotalPages)
		}
		if ctx.HasNext != false {
			t.Error("Expected HasNext=false with no results")
		}
	})
}
