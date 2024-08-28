package advanced_search

import (
	"testing"
)

func TestAdvancedSearch_Sql(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid query",
			query:   "username:=:'admin'",
			want:    "username = 'admin'",
			wantErr: false,
		},
		{
			name:    "valid query",
			query:   "username:!=:'tag1' username:!=:'tag2'",
			want:    "username != 'tag1' AND username != 'tag2'",
			wantErr: false,
		},
		{
			name:    "valid query",
			query:   "lagerbestand:>=:1.0",
			want:    "lagerbestand >= 1.0",
			wantErr: false,
		},
		{
			name:    "valid query",
			query:   "lagerbestand:>=:1",
			want:    "lagerbestand >= 1",
			wantErr: false,
		},
		{
			name:    "valid query",
			query:   "lagerbestand:>=:1a",
			want:    "lagerbestand >= '1a'",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAdvancedSqlSearch(tt.query).Sql()
			if (err != nil) != tt.wantErr {
				t.Errorf("AdvancedSearch.Sql() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AdvancedSearch.Sql() = %v, want %v", got, tt.want)
			}
		})
	}
}
