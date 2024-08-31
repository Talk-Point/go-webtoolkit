package advanced_search

import (
	"testing"
)

func TestNewAdvancedSqlSearch(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		columns   []Column
		wantWhere string
		wantSort  string
		wantErr   bool
	}{
		{
			name:      "single field with sort",
			query:     "field1:=:value1 sort:field1,field2",
			columns:   []Column{{Name: "field1", Type: StringType, Aliases: []string{"f1"}}},
			wantWhere: "field1 = 'value1'",
			wantSort:  "field1 ASC",
			wantErr:   false,
		},
		{
			name:      "multiple fields with sort",
			query:     "field1:=:value1 field2:>:10 sort:field1,field2",
			columns:   []Column{{Name: "field1", Type: StringType, Aliases: []string{"f1"}}, {Name: "field2", Type: IntType, Aliases: []string{"f2"}}},
			wantWhere: "field1 = 'value1' AND field2 > 10",
			wantSort:  "field1 ASC, field2 ASC",
			wantErr:   false,
		},
		{
			name:      "field with alias",
			query:     "f1:=:value1 sort:f1",
			columns:   []Column{{Name: "field1", Type: StringType, Aliases: []string{"f1"}}},
			wantWhere: "field1 = 'value1'",
			wantSort:  "field1 ASC",
			wantErr:   false,
		},
		{
			name:      "unsupported field",
			query:     "field3:=:value1 sort:field3",
			columns:   []Column{{Name: "field1", Type: StringType, Aliases: []string{"f1"}}},
			wantWhere: "",
			wantSort:  "",
			wantErr:   false,
		},
		{
			name:      "no sort",
			query:     "field1:=:value1",
			columns:   []Column{{Name: "field1", Type: StringType, Aliases: []string{"f1"}}},
			wantWhere: "field1 = 'value1'",
			wantSort:  "",
			wantErr:   false,
		},
		{
			name:      "invalid query format",
			query:     "field1:=value1",
			columns:   []Column{{Name: "field1", Type: StringType, Aliases: []string{"f1"}}},
			wantWhere: "",
			wantSort:  "",
			wantErr:   true,
		},
		{
			name:      "invalid type value",
			query:     "field1:=a",
			columns:   []Column{{Name: "field1", Type: IntType, Aliases: []string{"f1"}}},
			wantWhere: "",
			wantSort:  "",
			wantErr:   true,
		},
		{
			name:      "empty check",
			query:     "",
			columns:   []Column{{Name: "field1", Type: IntType, Aliases: []string{"f1"}}},
			wantWhere: "",
			wantSort:  "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as, err := NewAdvancedSqlSearch(tt.query, tt.columns...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			if err == nil {
				whereStatement, sortStatement, err := as.Sql()
				if (err != nil) != tt.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}

				if whereStatement != tt.wantWhere {
					t.Errorf("unexpected where statement: %s, want: %s", whereStatement, tt.wantWhere)
				}
				if sortStatement != tt.wantSort {
					t.Errorf("unexpected sort statement: %s, want: %s", sortStatement, tt.wantSort)
				}
			}
		})
	}
}
