package advanced_search

import (
	"fmt"
	"strings"
)

// ColumnType represents the type of a column in the database
type ColumnType int

const (
	StringType ColumnType = iota
	IntType
	FloatType
	BoolType
	DateType
	DateTimeType
)

// Column represents a column in the database with a name, type, and possible aliases
type Column struct {
	Name    string
	Type    ColumnType
	Aliases []string
}

// AdvancedSqlSearch extends the AdvancedSearch with column definitions
type AdvancedSqlSearch struct {
	base         *AdvancedSearch
	columns      []Column
	whereClauses []Clause
	sortClauses  []SortClause
}

// NewAdvancedSqlSearch creates a new instance of AdvancedSqlSearch
func NewAdvancedSqlSearch(query string, columns ...Column) (*AdvancedSqlSearch, error) {
	as := &AdvancedSqlSearch{
		base:    NewAdvancedSearch(query),
		columns: columns,
	}

	whereClauses, sortClauses, err := as.base.Parse()
	if err != nil {
		return nil, err
	}

	as.whereClauses = whereClauses
	as.sortClauses = sortClauses

	return as, nil
}

// getColumnByFieldName returns the column that matches the field name or alias
func (as *AdvancedSqlSearch) getColumnByFieldName(fieldName string) *Column {
	for _, col := range as.columns {
		if strings.EqualFold(col.Name, fieldName) {
			return &col
		}
		for _, alias := range col.Aliases {
			if strings.EqualFold(alias, fieldName) {
				return &col
			}
		}
	}
	return nil
}

// WhereStatement generates the SQL WHERE clause
func (as *AdvancedSqlSearch) WhereStatement() (string, error) {
	var whereClauses []string

	for _, clause := range as.whereClauses {
		column := as.getColumnByFieldName(clause.Field)
		if column != nil {
			whereClauses = append(whereClauses, clause.Sql(column.Name))
		}
	}

	if len(whereClauses) == 0 {
		return "", nil
	}

	return strings.Join(whereClauses, " AND "), nil
}

// SortStatement generates the SQL ORDER BY clause
func (as *AdvancedSqlSearch) SortStatement() (string, error) {
	var sortClauses []string

	for _, sortClause := range as.sortClauses {
		column := as.getColumnByFieldName(sortClause.Field)
		if column != nil {
			direction := "ASC"
			if sortClause.Direction == Desc {
				direction = "DESC"
			}
			sortClauses = append(sortClauses, fmt.Sprintf("%s %s", column.Name, direction))
		}
	}

	if len(sortClauses) == 0 {
		return "", nil
	}

	return strings.Join(sortClauses, ", "), nil
}

// Sql generates both the WHERE and ORDER BY clauses
func (as *AdvancedSqlSearch) Sql() (string, string, error) {
	whereStmt, err := as.WhereStatement()
	if err != nil {
		return "", "", err
	}

	sortStmt, err := as.SortStatement()
	if err != nil {
		return "", "", err
	}

	return whereStmt, sortStmt, nil
}
