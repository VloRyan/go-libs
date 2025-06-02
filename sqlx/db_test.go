package sqlx

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	_ "github.com/mattn/go-sqlite3"
)

func TestOpenPingClose(t *testing.T) {
	type args struct {
		driverName     string
		dataSourceName string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "sqlite in_mem",
			args: args{
				driverName:     "sqlite3",
				dataSourceName: ":memory:",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Open(tt.args.driverName, tt.args.dataSourceName)
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			if err := db.Ping(); err != nil {
				t.Fatalf("Ping() error = %v", err)
			}
			if err := db.Close(); err != nil {
				t.Fatalf("Close() error = %v", err)
			}
		})
	}
}

type testStruct struct {
	ID                  uint
	Name                string
	AttribWithOtherName string `DB:"attrib"`
}

func Test_DB_Select(t *testing.T) {
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
			db := prepareDB(t)
			for _, w := range tt.want {
				_, err := db.Exec("INSERT INTO test_table(name) VALUES(:name);", w)
				if err != nil {
					t.Fatalf("Exec() error = %v", err)
				}
			}
			var actualAsSlice []*testStruct
			var err error
			if len(tt.want) == 1 {
				actualAsSlice = []*testStruct{{}}
				err = db.Select(actualAsSlice[0], "SELECT * FROM test_table;")
			} else {
				var actual []*testStruct
				err = db.Select(&actual, "SELECT * FROM test_table;")
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

func prepareDB(t *testing.T) *DB {
	db, err := Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	createTestTable(t, db)
	return db
}

func createTestTable(t *testing.T, db *DB) {
	_, err := db.Exec(`CREATE TABLE test_table (
  				id INTEGER PRIMARY KEY AUTOINCREMENT, 
  				name TEXT NOT NULL
  			)`,
	)
	if err != nil {
		t.Fatal(err)
	}
}
