package search

import (
	"strings"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestParseOperator(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedOp  string
		expectedVal string
	}{
		{
			name:        "Equals operator",
			input:       "5",
			expectedOp:  "=",
			expectedVal: "5",
		},
		{
			name:        "Greater than or equals operator",
			input:       ">=4",
			expectedOp:  ">=",
			expectedVal: "4",
		},
		{
			name:        "Less than operator",
			input:       "<3",
			expectedOp:  "<",
			expectedVal: "3",
		},
		{
			name:        "No operator, just value",
			input:       "hello",
			expectedOp:  "=",
			expectedVal: "hello",
		},
		{
			name:        "Empty string",
			input:       "",
			expectedOp:  "=",
			expectedVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOp, gotVal := parseOperator(tt.input)
			if gotOp != tt.expectedOp {
				t.Errorf("parseOperator() got operator = %v, want %v", gotOp, tt.expectedOp)
			}
			if gotVal != tt.expectedVal {
				t.Errorf("parseOperator() got value = %v, want %v", gotVal, tt.expectedVal)
			}
		})
	}
}

// Helper to check if a substring exists in the generated SQL of the WHERE clause
func hasWhereClause(db *gorm.DB, substr string) bool {
	if db.Statement == nil {
		return false
	}
	c, ok := db.Statement.Clauses["WHERE"]
	if !ok {
		return false
	}

	// Reconstruct the SQL from the clause expression (basic approximation for testing)
	// We can iterate over c.Expression.(clause.Where).Exprs
	where, ok := c.Expression.(clause.Where)
	if !ok {
		return false
	}

	for _, expr := range where.Exprs {
		if sqlExpr, ok := expr.(clause.Expr); ok {
			if strings.Contains(sqlExpr.SQL, substr) {
				return true
			}
		}
	}
	return false
}

func TestEngineApply(t *testing.T) {
	engine := NewEngine()

	// Setup a dummy DB with a Statement so GORM methods don't panic
	db := &gorm.DB{
		Statement: &gorm.Statement{
			Clauses: make(map[string]clause.Clause),
			Table:   "images",
			Vars:    make([]interface{}, 0),
		},
		Config: &gorm.Config{
			DryRun: true,
		},
	}

	tests := []struct {
		name             string
		criteria         SearchCriteria
		wantWhereContain []string
	}{
		{
			name: "Basic Filters",
			criteria: SearchCriteria{
				Filters: map[string]string{
					"rating": ">=4",
				},
			},
			wantWhereContain: []string{"(image_metadata->>'rating')::numeric >= ?"},
		},
		{
			name: "Favourited True",
			criteria: SearchCriteria{
				Filters: map[string]string{
					"favourited": "true",
				},
			},
			wantWhereContain: []string{"favourited = ?"},
		},
		{
			name: "Favourited False",
			criteria: SearchCriteria{
				Filters: map[string]string{
					"favourited": "false",
				},
			},
			wantWhereContain: []string{"favourited = ? OR favourited IS NULL"},
		},
		{
			name: "Favorite Alias",
			criteria: SearchCriteria{
				Filters: map[string]string{
					"favorite": "true",
				},
			},
			wantWhereContain: []string{"favourited = ?"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset clauses for each run
			db.Statement.Clauses = make(map[string]clause.Clause)
			resultDB := engine.Apply(db, tt.criteria)

			for _, want := range tt.wantWhereContain {
				if !hasWhereClause(resultDB, want) {
					t.Errorf("Apply() expected WHERE clause containing %q, but not found", want)
				}
			}
		})
	}
}

func TestEngineApplyCollections(t *testing.T) {
	engine := NewEngine()

	// Setup a dummy DB
	db := &gorm.DB{
		Statement: &gorm.Statement{
			Clauses: make(map[string]clause.Clause),
			Table:   "collections",
			Vars:    make([]interface{}, 0),
		},
		Config: &gorm.Config{
			DryRun: true,
		},
	}

	tests := []struct {
		name             string
		criteria         SearchCriteria
		wantWhereContain []string
	}{
		{
			name: "Basic Filters",
			criteria: SearchCriteria{
				Filters: map[string]string{
					"owner": "jane",
				},
			},
			wantWhereContain: []string{"users.username = ?"},
		},
		{
			name: "Favourited True",
			criteria: SearchCriteria{
				Filters: map[string]string{
					"favourited": "true",
				},
			},
			wantWhereContain: []string{"favourited = ?"},
		},
		{
			name: "Favourited False",
			criteria: SearchCriteria{
				Filters: map[string]string{
					"favourited": "false",
				},
			},
			wantWhereContain: []string{"favourited = ? OR favourited IS NULL"},
		},
		{
			name: "Favorite Alias",
			criteria: SearchCriteria{
				Filters: map[string]string{
					"favorite": "true",
				},
			},
			wantWhereContain: []string{"favourited = ?"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset clauses for each run
			db.Statement.Clauses = make(map[string]clause.Clause)
			resultDB := engine.ApplyCollections(db, tt.criteria)

			for _, want := range tt.wantWhereContain {
				if !hasWhereClause(resultDB, want) {
					t.Errorf("ApplyCollections() expected WHERE clause containing %q, but not found", want)
				}
			}
		})
	}
}
