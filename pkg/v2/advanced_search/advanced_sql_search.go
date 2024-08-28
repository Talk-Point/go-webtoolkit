package advanced_search

import (
	"fmt"
	"strings"
)

type AdvancedSqlSearch struct {
	base    *AdvancedSearch
	columns []string
}

func NewAdvancedSqlSearch(query string, columns ...string) *AdvancedSqlSearch {
	return &AdvancedSqlSearch{
		base:    NewAdvancedSearch(query),
		columns: columns,
	}
}

func (as *AdvancedSqlSearch) Sql() (string, error) {
	clauses, err := as.base.Parse()
	if err != nil {
		return "", err
	}

	sqlParts := make([]string, 0, len(clauses))
	for _, clause := range clauses {
		if len(as.columns) > 0 {
			// Check if the clause field is in the allowed columns
			if !contains(as.columns, clause.Field) {
				return "", fmt.Errorf("field %s is not allowed", clause.Field)
			}
		}
		sqlParts = append(sqlParts, clause.Sql())
	}

	return strings.Join(sqlParts, " AND "), nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
