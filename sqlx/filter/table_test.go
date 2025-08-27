package filter

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vloryan/go-libs/testhelper"
)

func TestTableFilter(t *testing.T) {
	tests := []struct {
		name    string
		f       func() Criteria
		want    Criteria
		wantErr bool
	}{{
		name: "eq",
		f: func() Criteria {
			return NewTable("table").Column("field").Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "table.field",
			ValueExpr:  ":table_field",
			Parameter:  map[string]any{"table_field": "test"},
			TableName:  "table",
		},
	}, {
		name: "lower",
		f: func() Criteria {
			return NewTable("table").Column("field").ToLower().Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "LOWER(table.field)",
			ValueExpr:  ":table_field",
			Parameter:  map[string]any{"table_field": "test"},
			TableName:  "table",
		},
	}, {
		name: "asDate",
		f: func() Criteria {
			return NewTable("table").Column("field").AsDate().Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "DATE(table.field)",
			ValueExpr:  ":table_field",
			Parameter:  map[string]any{"table_field": "test"},
			TableName:  "table",
		},
	}, {
		name: "jsonb_extract",
		f: func() Criteria {
			return NewTable("table").Column("field").JSONBExtract("$.A").Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "jsonb_extract(table.field, '$.A')",
			ValueExpr:  ":table_field",
			Parameter:  map[string]any{"table_field": "test"},
			TableName:  "table",
		},
	}, {
		name: "custom func",
		f: func() Criteria {
			return NewTable("table").Column("field").WithColumnFunc(func(columnExpr string, args ...any) string {
				var strArgs []string
				for _, arg := range args {
					strArgs = append(strArgs, arg.(string))
				}
				return "CUSTOM(" + columnExpr + ", '" + strings.Join(strArgs, "', '") + "')"
			}, "a", "b", "c").Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "CUSTOM(table.field, 'a', 'b', 'c')",
			ValueExpr:  ":table_field",
			Parameter:  map[string]any{"table_field": "test"},
			TableName:  "table",
		},
	}, {
		name: "data compare",
		f: func() Criteria {
			return NewTable("table").Column("field").AsDate().Eq(testhelper.FixedNow, AsDate)
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "DATE(table.field)",
			ValueExpr:  "DATE(:table_field)",
			Parameter:  map[string]any{"table_field": testhelper.FixedNow},
			TableName:  "table",
		},
	}, {
		name: "between",
		f: func() Criteria {
			return NewTable("table").Column("field").Between(0, 1)
		},
		want: &UnaryCriteria{
			OpType:     BetweenOp,
			ColumnExpr: "table.field",
			ValueExpr:  ":table_field_0 AND :table_field_1",
			Parameter:  map[string]any{"table_field_0": 0, "table_field_1": 1},
			TableName:  "table",
		},
	}, {
		name: "in",
		f: func() Criteria {
			return NewTable("table").Column("field").In([]any{1, 2, 4})
		},
		want: &UnaryCriteria{
			OpType:     InOp,
			ColumnExpr: "table.field",
			ValueExpr:  "(:table_field_0, :table_field_1, :table_field_2)",
			Parameter:  map[string]any{"table_field_0": 1, "table_field_1": 2, "table_field_2": 4},
			TableName:  "table",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.f()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToWhere() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
