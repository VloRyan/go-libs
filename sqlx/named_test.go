package sqlx

import (
	"reflect"
	"strings"
	"testing"
)

func Test_compileNamedQuery(t *testing.T) {
	type args struct {
		query    string
		bindType rune
	}
	tests := []struct {
		name      string
		args      args
		wantNames []string
	}{{
		name: "Question mark",
		args: args{
			query:    "SELECT * FROM my_table WHERE col_a = :name1 AND col_b=:name2 AND col_c = :name_3 AND col_d = :object.field",
			bindType: '?',
		},
		wantNames: []string{"name1", "name2", "name_3", "object.field"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantQuery := tt.args.query
			for _, name := range tt.wantNames {
				wantQuery = strings.Replace(wantQuery, ":"+name, string(tt.args.bindType), 1)
			}
			gotQuery, gotNames, err := compileNamedQuery([]byte(tt.args.query), tt.args.bindType)
			if err != nil {
				t.Fatalf("compileNamedQuery() error = %v", err)
			}
			if gotQuery != wantQuery {
				t.Fatalf("compileNamedQuery() gotQuery = %v, want %v", gotQuery, wantQuery)
			}
			if !reflect.DeepEqual(gotNames, tt.wantNames) {
				t.Fatalf("compileNamedQuery() gotNames = %v, want %v", gotNames, tt.wantNames)
			}
		})
	}
}
