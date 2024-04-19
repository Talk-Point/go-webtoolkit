package query

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/Talk-Point/go-toolkit/pkg/v2/shared"
)

type Operator string

func (o Operator) String() string {
	return string(o)
}

func (o Operator) ToFireStoreOperator() string {
	switch o {
	case Eq:
		return "=="
	case Eqe:
		return "=="
	case Gt:
		return ">"
	case Gte:
		return ">="
	case Lt:
		return "<"
	case Lte:
		return "<="
	case Contains:
		return "in"
	case ArrayContains:
		return "array-contains"
	default:
		return "=="
	}
}

const (
	Eq            Operator = "eq"
	Eqe           Operator = "eqe"
	Gt            Operator = "gt"
	Gte           Operator = "gte"
	Lt            Operator = "lt"
	Lte           Operator = "lte"
	Contains      Operator = "contains"
	ArrayContains Operator = "array-contains"
)

type Filter struct {
	Field    string
	Operator Operator
	Value    interface{}
}

func parseOperator(value string) Operator {
	switch value {
	case "eq":
		return Eq
	case "eqe":
		return Eqe
	case "gt":
		return Gt
	case "gte":
		return Gte
	case "lt":
		return Lt
	case "lte":
		return Lte
	case "contains":
		return Contains
	case "array-contains":
		return ArrayContains
	default:
		return Eq
	}
}

func NewFiltersFromUrl(value *url.URL) ([]Filter, error) {
	var filters []Filter
	for key, values := range value.Query() {
		if len(values) > 0 {
			if shared.Contains(key, []string{"limit", "sort", "next", "prev"}) {
				continue
			}
			for _, v := range values {
				if strings.Contains(key, "__") {
					parts := strings.Split(key, "__")
					if len(parts) != 2 {
						filters = append(filters, Filter{
							Field:    parts[0],
							Operator: Eq,
							Value:    v,
						})
						continue
					}
					operator := parseOperator(parts[1])
					filters = append(filters, Filter{
						Field:    parts[0],
						Operator: operator,
						Value:    v,
					})
				} else {
					filters = append(filters, Filter{
						Field:    key,
						Operator: Eq,
						Value:    v,
					})
				}
			}
		}
	}
	return filters, nil
}

func NewFilterFromUrlString(value string) ([]Filter, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	return NewFiltersFromUrl(u)
}

type Direction string

const (
	Asc  Direction = "asc"
	Desc Direction = "desc"
)

type QueryOptions struct {
	Limit            int
	Next             string
	Previous         string
	OrderBy          string
	OrderByDirection Direction
	Filters          []Filter
}

func parseLimit(value string, maxLimit int, defaultLimit int) int {
	l := defaultLimit
	if value != "" {
		limit, err := strconv.Atoi(value)
		if err == nil {
			l = limit
		}
	}
	if l < 1 {
		return 1
	}
	if l > maxLimit {
		return maxLimit
	}
	return l
}

func parseOrderBy(value string) (string, Direction) {
	if value == "" {
		return "id", Desc
	}

	direction := Asc
	if strings.HasPrefix(value, "-") {
		direction = Desc
		value = strings.TrimPrefix(value, "-")
	}

	return value, direction
}

func NewQueryOptionsFromUrl(value *url.URL) (QueryOptions, error) {
	q := value.Query()

	limit := parseLimit(q.Get("limit"), 100, 30)
	orderBy, orderDirection := parseOrderBy(q.Get("sort"))
	next := q.Get("next")
	previous := q.Get("prev")
	filters, err := NewFiltersFromUrl(value)
	if err != nil {
		return QueryOptions{}, err
	}

	return QueryOptions{
		Limit:            limit,
		Next:             next,
		Previous:         previous,
		OrderBy:          orderBy,
		OrderByDirection: orderDirection,
		Filters:          filters,
	}, nil
}

func NewQueryOptionsFromUrlString(value string) (QueryOptions, error) {
	u, err := url.Parse(value)
	if err != nil {
		return QueryOptions{}, err
	}
	return NewQueryOptionsFromUrl(u)
}

func buildQueryString(params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		if value == "" {
			continue
		}
		if key == "" {
			continue
		}
		values.Add(key, value)
	}
	return values.Encode()
}

func NewQueryOptionsFromForm(values url.Values, fieldMappings ...map[string]string) (QueryOptions, error) {
	params := make(map[string]string)
	var fieldMapping map[string]string
	if len(fieldMappings) > 0 {
		fieldMapping = fieldMappings[0]
	}
	for key, values := range values {
		if len(values) > 0 {
			// Check if there's a mapping for the key
			if fieldMapping != nil {
				if newKey, ok := fieldMapping[key]; ok {
					params[newKey] = values[0]
					continue
				}
			}
			params[key] = values[0]
		}
	}
	queryString := buildQueryString(params)
	return NewQueryOptionsFromUrlString("/?" + queryString)
}
