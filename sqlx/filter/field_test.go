package filter

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vloryan/go-libs/testhelper"
)

func TestColumnFilter(t *testing.T) {
	tests := []struct {
		name    string
		f       func() Criteria
		want    Criteria
		wantErr bool
	}{{
		name: "eq",
		f: func() Criteria {
			return Field("field").Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "field",
			ValueExpr:  ":field",
			Parameter:  map[string]any{"field": "test"},
		},
	}, {
		name: "lower",
		f: func() Criteria {
			return Field("field").ToLower().Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "LOWER(field)",
			ValueExpr:  ":field",
			Parameter:  map[string]any{"field": "test"},
		},
	}, {
		name: "asDate",
		f: func() Criteria {
			return Field("field").AsDate().Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "DATE(field)",
			ValueExpr:  ":field",
			Parameter:  map[string]any{"field": "test"},
		},
	}, {
		name: "jsonb_extract",
		f: func() Criteria {
			return Field("field").JSONBExtract("$.A").Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "jsonb_extract(field, '$.A')",
			ValueExpr:  ":field",
			Parameter:  map[string]any{"field": "test"},
		},
	}, {
		name: "custom func",
		f: func() Criteria {
			return Field("field").WithColumnFunc(func(columnExpr string, args ...any) string {
				var strArgs []string
				for _, arg := range args {
					strArgs = append(strArgs, arg.(string))
				}
				return "CUSTOM(" + columnExpr + ", '" + strings.Join(strArgs, "', '") + "')"
			}, "a", "b", "c").Eq("test")
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "CUSTOM(field, 'a', 'b', 'c')",
			ValueExpr:  ":field",
			Parameter:  map[string]any{"field": "test"},
		},
	}, {
		name: "data compare",
		f: func() Criteria {
			return Field("field").AsDate().Eq(testhelper.FixedNow, AsDate)
		},
		want: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "DATE(field)",
			ValueExpr:  "DATE(:field)",
			Parameter:  map[string]any{"field": testhelper.FixedNow},
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
