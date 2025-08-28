package statement

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUpdate(t *testing.T) {
	testcases := []struct {
		name           string
		tableName      string
		columns        []ColumnExpression
		whereCondition string
		want           string
	}{
		{
			name:           "Update Where",
			tableName:      "t1",
			columns:        FieldColumnsFromNames("f1", "f2"),
			whereCondition: "f1 > 100",
			want:           "UPDATE t1 SET f1 = :f1, f2 = :f2 WHERE f1 > 100",
		}, {
			name:      "Update Param Alias",
			tableName: "t1",
			columns: []ColumnExpression{{
				Name:  "f1",
				Alias: "field_1",
			}, {
				Name:  "f2",
				Alias: "field_2",
			}},
			whereCondition: "f1 > 100",
			want:           "UPDATE t1 SET f1 = :field_1, f2 = :field_2 WHERE f1 > 100",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmt := NewUpdate(tt.tableName).
				WithFields(tt.columns...).
				Where(tt.whereCondition)
			actual := stmt.SQL()
			if diff := cmp.Diff(actual, tt.want); diff != "" {
				t.Errorf("actual(-), want(+): %v", diff)
			}
		})
	}
}
