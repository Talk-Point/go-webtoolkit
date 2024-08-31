package advanced_search

import (
	"fmt"
	"strconv"
	"strings"
)

// Operator represents the type for SQL operators
type Operator int

const (
	Equal Operator = iota
	NotEqual
	GreaterThan
	GreaterThanOrEqual
	LessThan
	LessThanOrEqual
	Null
	NotNull
	Like
	UnknownOperator
)

// String returns the string representation of the Operator
func (o Operator) String() string {
	switch o {
	case Equal:
		return "="
	case NotEqual:
		return "!="
	case GreaterThan:
		return ">"
	case GreaterThanOrEqual:
		return ">="
	case LessThan:
		return "<"
	case LessThanOrEqual:
		return "<="
	case Null:
		return "IS NULL"
	case NotNull:
		return "IS NOT NULL"
	case Like:
		return "LIKE"
	default:
		return "unknown"
	}
}

// Map string representations to enum values
var operatorMap = map[string]Operator{
	"=":       Equal,
	"!=":      NotEqual,
	">":       GreaterThan,
	">=":      GreaterThanOrEqual,
	"<":       LessThan,
	"<=":      LessThanOrEqual,
	"null":    Null,
	"notnull": NotNull,
	"like":    Like,
}

// Clause represents a single SQL WHERE clause
type Clause struct {
	Field    string
	Operator Operator
	Value    string
}

func escapeString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func (c Clause) Sql(name string) string {
	field := c.Field
	if name != "" {
		field = name
	}

	switch c.Operator {
	case Null, NotNull:
		return fmt.Sprintf("%s %s", field, c.Operator)
	case Like:
		return fmt.Sprintf("%s %s '%%%s%%'", field, c.Operator, c.Value)
	default:
		// Check if the value is an integer or float
		if _, err := strconv.Atoi(c.Value); err == nil {
			return fmt.Sprintf("%s %s %s", field, c.Operator, c.Value)
		}
		if _, err := strconv.ParseFloat(c.Value, 64); err == nil {
			return fmt.Sprintf("%s %s %s", field, c.Operator, c.Value)
		}
		return fmt.Sprintf("%s %s '%s'", field, c.Operator, escapeString(c.Value))
	}
}

func NewClause(field string, operator Operator, value string) Clause {
	return Clause{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

type Direction int

const (
	Asc Direction = iota
	Desc
)

type SortClause struct {
	Field     string
	Direction Direction
}

// ParseError represents a parsing error with context
type ParseError struct {
	Part  string
	Cause string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("error parsing part '%s': %s", e.Part, e.Cause)
}

// AdvancedSearch represents the advanced search query
type AdvancedSearch struct {
	query string
}

// NewAdvancedSearch creates a new instance of AdvancedSearch
func NewAdvancedSearch(query string) *AdvancedSearch {
	return &AdvancedSearch{query: query}
}

func parseSortClause(sort string) *SortClause {
	if strings.HasSuffix(sort, "__asc") {
		return &SortClause{
			Field:     sort[:len(sort)-5],
			Direction: Asc,
		}
	}
	if strings.HasSuffix(sort, "__desc") {
		return &SortClause{
			Field:     sort[:len(sort)-6],
			Direction: Desc,
		}
	}

	return &SortClause{
		Field:     sort,
		Direction: Asc,
	}
}

// Parse parses the query string and returns a slice of Clause objects
func (as *AdvancedSearch) Parse() ([]Clause, []SortClause, error) {
	queryParts := strings.Fields(as.query)
	if len(queryParts) == 0 {
		return nil, nil, nil
	}

	clauses := make([]Clause, 0, len(queryParts))
	sortClauses := make([]SortClause, 0)

	for _, part := range queryParts {
		if strings.HasPrefix(part, "sort:") {
			sortParts := strings.SplitN(part, ":", 2)
			if len(sortParts) != 2 {
				return nil, nil, &ParseError{Part: part, Cause: "invalid sort format"}
			}
			for _, sort := range strings.Split(sortParts[1], ",") {
				sortClause := parseSortClause(sort)
				if sortClause != nil {
					sortClauses = append(sortClauses, *sortClause)
				}
			}
			continue
		}

		clauseParts := strings.SplitN(part, ":", 3)
		if len(clauseParts) < 2 {
			return nil, nil, &ParseError{Part: part, Cause: "invalid query format"}
		}

		field := clauseParts[0]
		operatorKey := clauseParts[1]

		operator, ok := operatorMap[strings.ToLower(operatorKey)]
		if !ok {
			return nil, nil, &ParseError{Part: part, Cause: "unsupported operator"}
		}

		var value string
		if operator != Null && operator != NotNull {
			if len(clauseParts) != 3 {
				return nil, nil, &ParseError{Part: part, Cause: "missing value"}
			}
			value = clauseParts[2]
			if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
				value = value[1 : len(value)-1]
			}
		}

		clause := NewClause(field, operator, value)
		clauses = append(clauses, clause)
	}

	return clauses, sortClauses, nil
}
