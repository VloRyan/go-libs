package statement

import (
	"reflect"
	"testing"

	"github.com/vloryan/go-libs/sqlx/pagination"
)

func TestToOrderBy(t *testing.T) {
	type args struct {
		fields    []string
		tableName string
	}
	tests := []struct {
		name string
		args args
		want []*OrderBy
	}{{
		name: "Ascending",
		args: args{
			fields:    []string{"FieldA"},
			tableName: "TabA",
		},
		want: []*OrderBy{
			{
				FieldName: "tab_a.field_a",
			},
		},
	}, {
		name: "Descending",
		args: args{
			fields:    []string{"-FieldA"},
			tableName: "TabA",
		},
		want: []*OrderBy{
			{
				FieldName: "tab_a.field_a",
				Direction: OrderDescending,
			},
		},
	}, {
		name: "Mix",
		args: args{
			fields:    []string{"-FieldA", "FieldB", "Sub.FieldA", "-Sub.Base.FieldB"},
			tableName: "TabA",
		},
		want: []*OrderBy{
			{
				FieldName: "tab_a.field_a",
				Direction: OrderDescending,
			},
			{
				FieldName: "tab_a.field_b",
			},
			{
				FieldName: "sub.field_a",
			},
			{
				FieldName: "sub_entity.field_b",
				Direction: OrderDescending,
			},
		},
	}, {
		name: "Coalesce",
		args: args{
			fields:    []string{"FieldA:FieldB"},
			tableName: "TabA",
		},
		want: []*OrderBy{
			{
				FieldName: "tab_a.field_a",
				Coalesce:  "FieldB",
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &pagination.Page{
				Sort: tt.args.fields,
			}
			got := ToOrderBy(p, tt.args.tableName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToOrderBy() got = %v, want %v", got, tt.want)
			}
		})
	}
}
