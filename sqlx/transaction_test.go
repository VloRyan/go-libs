package sqlx

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Transaction_Select(t *testing.T) {
	tests := []struct {
		name string
		want []*testStruct
	}{{
		name: "Select one",
		want: []*testStruct{{
			ID:                  1,
			Name:                "Hans",
			AttribWithOtherName: "",
		}},
	}, {
		name: "Select multiple",
		want: []*testStruct{{
			ID:                  1,
			Name:                "Maxima",
			AttribWithOtherName: "",
		}, {
			ID:                  2,
			Name:                "Ludger",
			AttribWithOtherName: "",
		}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := prepareDB(t).Begin()
			if err != nil {
				t.Fatalf("DB.Begin() error = %v", err)
			}
			defer func(tx *Transaction) {
				_ = tx.Rollback()
			}(tx)
			for _, w := range tt.want {
				_, err := tx.Exec("INSERT INTO test_table(name) VALUES(:name);", w)
				if err != nil {
					t.Fatalf("Exec() error = %v", err)
				}
			}
			var actualAsSlice []*testStruct
			if len(tt.want) == 1 {
				actualAsSlice = []*testStruct{{}}
				err = tx.Select(actualAsSlice[0], "SELECT * FROM test_table;")
			} else {
				var actual []*testStruct
				err = tx.Select(&actual, "SELECT * FROM test_table;")
				actualAsSlice = actual
			}
			if err != nil {
				t.Fatalf("Exec() error = %v", err)
			}

			if diff := cmp.Diff(tt.want, actualAsSlice); diff != "" {
				t.Fatalf("Select() mismatch (-want +got): %v", diff)
			}
		})
	}
}
