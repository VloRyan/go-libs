package filter

import "testing"

func TestWhere_Clause(t *testing.T) {
	tests := []struct {
		name  string
		where Where
		want  string
	}{{
		name: "empty",
		want: "",
	}, {
		name:  "where",
		where: Where{Statement: "fieldA = :a"},
		want:  "WHERE fieldA = :a",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.where.Clause(); got != tt.want {
				t.Errorf("Clause() = %v, want %v", got, tt.want)
			}
		})
	}
}
