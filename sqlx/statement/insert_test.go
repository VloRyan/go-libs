package statement

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInsert(t *testing.T) {
	testcases := []struct {
		name      string
		tableName string
		columns   []ColumnExpression
		valuesLen uint
		want      string
	}{
		{
			name:      "Simple",
			tableName: "t1",
			columns:   FieldColumnsFromNames("f1", "f2"),
			want:      "INSERT INTO t1(f1, f2) VALUES(:f1, :f2)",
		},
		{
			name:      "With function",
			tableName: "t1",
			columns: []ColumnExpression{
				{Name: "f1", ValueExpression: "CURRENT_TIMESTAMP", Source: Expression},
				{Name: "f2", ValueExpression: "jsonb(:f2)", Source: Expression},
			},
			want: "INSERT INTO t1(f1, f2) VALUES(CURRENT_TIMESTAMP, jsonb(:f2))",
		},
		{
			name:      "With multi values",
			tableName: "t1",
			columns:   FieldColumnsFromNames("f1", "f2"),
			valuesLen: 3,
			want:      "INSERT INTO t1(f1, f2) VALUES(:f1[0], :f2[0]), (:f1[1], :f2[1]), (:f1[2], :f2[2])",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmt := NewInsert(tt.tableName).
				WithValuesLen(tt.valuesLen).
				WithFields(tt.columns...)
			actual := stmt.SQL()
			if diff := cmp.Diff(actual, tt.want); diff != "" {
				t.Errorf("actual(-), want(+): %v", diff)
			}
		})
	}
}
