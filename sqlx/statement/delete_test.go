package statement

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDelete(t *testing.T) {
	testcases := []struct {
		name           string
		tableName      string
		whereCondition string
		want           string
	}{
		{
			name:           "Delete Where",
			tableName:      "t1",
			whereCondition: "f1 > 100",
			want:           "DELETE FROM t1 WHERE f1 > 100",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			stmt := NewDelete(tt.tableName).
				Where(tt.whereCondition)
			actual := stmt.SQL()
			if diff := cmp.Diff(actual, tt.want); diff != "" {
				t.Errorf("actual(-), want(+): %v", diff)
			}
		})
	}
}
