package statement

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSelect(t *testing.T) {
	testcases := []struct {
		name           string
		tableName      string
		columns        []ColumnExpression
		whereCondition string
		joins          []TableJoinDefinition
		order          []*OrderBy
		limit          int
		offset         int
		want           string
	}{
		{
			name:      "Simple",
			tableName: "t1",
			columns:   FieldColumnsFromNames("f1", "f2"),
			want:      "SELECT t1.f1, t1.f2 FROM t1",
		},
		{
			name:           "Where",
			tableName:      "t1",
			columns:        FieldColumnsFromNames("f1", "f2"),
			whereCondition: "t1.f1 > 100",
			want:           "SELECT t1.f1, t1.f2 FROM t1 WHERE t1.f1 > 100",
		},
		{
			name:      "Single join",
			tableName: "t1",
			columns:   FieldColumnsFromNames("f1", "f2"),
			joins: []TableJoinDefinition{
				{
					Table: ObjectName{
						Name: "t2",
					},
					SelectFields: FieldColumnsFromNames("f1", "f2"),
					OnConditions: []string{"t1.f1 = t2.f1"},
				},
			},
			want: "SELECT t1.f1, t1.f2, t2.f1, t2.f2 FROM t1 JOIN t2 ON t1.f1 = t2.f1",
		},
		{
			name:      "Multi join",
			tableName: "t1",
			columns:   FieldColumnsFromNames("f1", "f2"),
			joins: []TableJoinDefinition{
				{
					Table: ObjectName{
						Name: "t2",
					},
					SelectFields: FieldColumnsFromNames("f1", "f2"),
					OnConditions: []string{"t2.f1 = t1.f1", "t2.f2 = 'TYPE'"},
				},
				{
					Table: ObjectName{
						Name: "t3",
					},
					OnConditions: []string{"t3.f1 = t2.f1"},
				},
				{
					Table: ObjectName{
						Name: "t4",
					},
					SelectFields: FieldColumnsFromNames("f1", "f2"),
					OnConditions: []string{"t4.f1 = t3.f1"},
				},
			},
			want: "SELECT t1.f1, t1.f2, t2.f1, t2.f2, t4.f1, t4.f2 FROM t1 JOIN t2 ON t2.f1 = t1.f1 AND t2.f2 = 'TYPE' JOIN t3 ON t3.f1 = t2.f1 JOIN t4 ON t4.f1 = t3.f1",
		},
		{
			name:      "Order",
			tableName: "t1",
			columns:   FieldColumnsFromNames("f1", "f2"),
			order:     []*OrderBy{{FieldName: "t1.f1"}},
			want:      "SELECT t1.f1, t1.f2 FROM t1 ORDER BY t1.f1",
		},
		{
			name:           "Offset with limit",
			tableName:      "t1",
			columns:        FieldColumnsFromNames("f1", "f2"),
			whereCondition: "t1.f1 > 100",
			offset:         100,
			limit:          3,
			want:           "SELECT t1.f1, t1.f2 FROM t1 WHERE t1.f1 > 100 LIMIT 3 OFFSET 100",
		},
		{
			name:      "Count",
			tableName: "t1",
			columns: []ColumnExpression{
				{Name: "Total count", SelectExpression: "COUNT(*)", Alias: "total_count", Source: Expression},
			},
			want: "SELECT COUNT(*) AS total_count FROM t1",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmt := NewSelect(tt.columns...).
				From(tt.tableName).
				Joins(tt.joins...).
				Where(tt.whereCondition).
				Offset(tt.offset).
				Order(tt.order).
				Limit(tt.limit)

			actual := stmt.SQL()
			if diff := cmp.Diff(actual, tt.want); diff != "" {
				t.Errorf("actual(-), want(+): %v", diff)
			}
		})
	}
}
