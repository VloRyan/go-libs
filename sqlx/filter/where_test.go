package filter

import "testing"

func TestWhere_SQL(t *testing.T) {
	tests := []struct {
		name  string
		where Where
		want  string
	}{{
		name: "empty",
		want: "",
	}, {
		name:  "where",
		where: Where{Clause: "fieldA = :a"},
		want:  "WHERE fieldA = :a",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.where.SQL(); got != tt.want {
				t.Errorf("SQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
