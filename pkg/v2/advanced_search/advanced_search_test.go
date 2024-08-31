package advanced_search

import (
	"reflect"
	"testing"
)

func TestParseWhereClause(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    []Clause
		wantErr bool
	}{
		{
			name:  "single clause with equal operator",
			query: "field1:=:value1",
			want: []Clause{
				{Field: "field1", Operator: Equal, Value: "value1"},
			},
			wantErr: false,
		},
		{
			name:  "single clause with not equal operator",
			query: "field1:!=:value1",
			want: []Clause{
				{Field: "field1", Operator: NotEqual, Value: "value1"},
			},
			wantErr: false,
		},
		{
			name:  "multiple clauses",
			query: "field1:=:value1 field2:>:value2",
			want: []Clause{
				{Field: "field1", Operator: Equal, Value: "value1"},
				{Field: "field2", Operator: GreaterThan, Value: "value2"},
			},
			wantErr: false,
		},
		{
			name:    "invalid format missing parts",
			query:   "field1:value1",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unsupported operator",
			query:   "field1:unsupported:value1",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing value for non-null operator",
			query:   "field1:=",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "value with special characters",
			query: "field1:=:'value1!@#$%^&*()'",
			want: []Clause{
				{Field: "field1", Operator: Equal, Value: "value1!@#$%^&*()"},
			},
			wantErr: false,
		},
		{
			name:  "value with single quotes",
			query: "field1:=:'value1'",
			want: []Clause{
				{Field: "field1", Operator: Equal, Value: "value1"},
			},
			wantErr: false,
		},
		{
			name:    "empty check",
			query:   "",
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := NewAdvancedSearch(tt.query)
			got, _, err := as.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSortClause(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    []SortClause
		wantErr bool
	}{
		{
			name:  "single clause with equal operator",
			query: "sort:field1__desc,field2__asc,field3",
			want: []SortClause{
				{Field: "field1", Direction: Desc},
				{Field: "field2", Direction: Asc},
				{Field: "field3", Direction: Asc},
			},
			wantErr: false,
		},
		{
			name:  "single clause with equal operator",
			query: "sort:field3",
			want: []SortClause{
				{Field: "field3", Direction: Asc},
			},
			wantErr: false,
		},
		{
			name:  "single clause with equal operator",
			query: "sort:field3,field__2,field__1__desc",
			want: []SortClause{
				{Field: "field3", Direction: Asc},
				{Field: "field__2", Direction: Asc},
				{Field: "field__1", Direction: Desc},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := NewAdvancedSearch(tt.query)
			_, got, err := as.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClause_Sql(t *testing.T) {
	tests := []struct {
		name   string
		clause Clause
		want   string
	}{
		{
			name: "simple SQL injection",
			clause: Clause{
				Field:    "username",
				Operator: Equal,
				Value:    "admin' OR '1'='1",
			},
			want: "username = 'admin'' OR ''1''=''1'",
		},
		{
			name: "escaped characters",
			clause: Clause{
				Field:    "username",
				Operator: Equal,
				Value:    "O'Reilly",
			},
			want: "username = 'O''Reilly'",
		},
		{
			name: "numeric value",
			clause: Clause{
				Field:    "age",
				Operator: GreaterThan,
				Value:    "25",
			},
			want: "age > 25",
		},
		{
			name: "float value",
			clause: Clause{
				Field:    "price",
				Operator: LessThan,
				Value:    "19.99",
			},
			want: "price < 19.99",
		},
		{
			name: "special characters",
			clause: Clause{
				Field:    "description",
				Operator: Like,
				Value:    "20% off!",
			},
			want: "description LIKE '%20% off!%'",
		},
		{
			name: "SQL comment injection",
			clause: Clause{
				Field:    "username",
				Operator: Equal,
				Value:    "admin' --",
			},
			want: "username = 'admin'' --'",
		},
		{
			name: "union select injection",
			clause: Clause{
				Field:    "username",
				Operator: Equal,
				Value:    "admin' UNION SELECT * FROM users --",
			},
			want: "username = 'admin'' UNION SELECT * FROM users --'",
		},
		{
			name: "boolean-based injection",
			clause: Clause{
				Field:    "username",
				Operator: Equal,
				Value:    "admin' AND '1'='1",
			},
			want: "username = 'admin'' AND ''1''=''1'",
		},
		{
			name: "time-based injection",
			clause: Clause{
				Field:    "username",
				Operator: Equal,
				Value:    "admin' AND SLEEP(5) --",
			},
			want: "username = 'admin'' AND SLEEP(5) --'",
		},
		{
			name: "hexadecimal injection",
			clause: Clause{
				Field:    "username",
				Operator: Equal,
				Value:    "admin' AND 0x50=0x50 --",
			},
			want: "username = 'admin'' AND 0x50=0x50 --'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.clause.Sql(""); got != tt.want {
				t.Errorf("Clause.Sql() = %v, want %v", got, tt.want)
			}
		})
	}
}
