package sqlx

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-cmp/cmp"
)

type dummyObject struct {
	ID     uint
	Nested struct {
		Name string
	}
}

func TestExec(t *testing.T) {
	type args struct {
		query string
		args  any
	}
	tests := []struct {
		name        string
		args        args
		setupExpect func(m sqlmock.Sqlmock)
		wantErr     assert.ErrorAssertionFunc
	}{{
		name: "struct arg",
		args: args{
			query: "SELECT id FROM tab WHERE id = :id AND name = :nested.name",
			args:  &dummyObject{ID: 7, Nested: struct{ Name string }{Name: "test"}},
		},
		setupExpect: func(m sqlmock.Sqlmock) {
			m.ExpectExec("SELECT id FROM tab WHERE id = ? AND name = ?").
				WithArgs(uint(7), "test").
				WillReturnResult(sqlmock.NewResult(7, 1))
		},
		wantErr: assert.NoError,
	}, {
		name: "map arg",
		args: args{
			query: "SELECT id FROM tab WHERE id = :id",
			args:  map[string]any{"id": uint(7)},
		},
		setupExpect: func(m sqlmock.Sqlmock) {
			m.ExpectExec("SELECT id FROM tab WHERE id = ?").
				WithArgs(uint(7)).
				WillReturnResult(sqlmock.NewResult(7, 1))
		},
		wantErr: assert.NoError,
	}, {
		name: "map arg in",
		args: args{
			query: "SELECT id FROM tab WHERE id IN (:p1, :p2)",
			args:  map[string]any{"p1": uint(7), "p2": uint(8)},
		},
		setupExpect: func(m sqlmock.Sqlmock) {
			query := "SELECT id FROM tab WHERE id IN (?, ?)"
			m.ExpectExec(query).
				WithArgs(uint(7), uint(8)).
				WillReturnResult(sqlmock.NewResult(8, 2))
		},
		wantErr: assert.NoError,
	}, {
		name: "slice args",
		args: args{
			query: "INSERT INTO tab(f1, f2) VALUES (:p1[0], :p2[0]), (:p1[1], :p2[1]), (:p1[2], :p2[2])",
			args: []map[string]any{
				{"p1": uint(7), "p2": uint(8)},
				{"p1": uint(7), "p2": uint(8)},
				{"p1": uint(7), "p2": uint(8)},
			},
		},
		setupExpect: func(m sqlmock.Sqlmock) {
			query := "INSERT INTO tab(f1, f2) VALUES (?, ?), (?, ?), (?, ?)"
			m.ExpectExec(query).
				WithArgs(
					uint(7), uint(8),
					uint(7), uint(8),
					uint(7), uint(8),
				).
				WillReturnResult(sqlmock.NewResult(8, 2))
		},
		wantErr: assert.NoError,
	}, {
		name: "slice args with one element",
		args: args{
			query: "INSERT INTO tab(f1, f2) VALUES (:p1, :p2)",
			args:  []map[string]any{{"p1": uint(7), "p2": uint(8)}},
		},
		setupExpect: func(m sqlmock.Sqlmock) {
			query := "INSERT INTO tab(f1, f2) VALUES (?, ?)"
			m.ExpectExec(query).
				WithArgs(uint(7), uint(8)).
				WillReturnResult(sqlmock.NewResult(8, 2))
		},
		wantErr: assert.NoError,
	}, {
		name: "fail on unresolvable struct path",
		args: args{
			query: "SELECT id FROM tab WHERE id = :id AND name = :nested.type",
			args:  &dummyObject{ID: 7, Nested: struct{ Name string }{Name: "test"}},
		},
		setupExpect: func(m sqlmock.Sqlmock) {
			m.ExpectRollback()
		},
		wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
			return assert.ErrorContains(t, err, "'nested.type'", i...)
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatal(err)
			}
			tt.setupExpect(mock)
			_, err = Exec(db, tt.args.query, tt.args.args)
			tt.wantErr(t, err, "Exec()")
		})
	}
}

func TestSelect(t *testing.T) {
	type args struct {
		query string
		args  []any
	}
	tests := []struct {
		name        string
		args        args
		setupExpect func(m sqlmock.Sqlmock)
		wantObject  any
	}{{
		name: "object arg",
		args: args{
			query: "SELECT id FROM tab WHERE id = :id",
			args:  []any{&dummyObject{ID: 7}},
		},
		setupExpect: func(m sqlmock.Sqlmock) {
			m.ExpectQuery("SELECT id FROM tab WHERE id = ?").
				WithArgs(uint(7)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
		},
		wantObject: &dummyObject{ID: 7},
	}, {
		name: "map arg",
		args: args{
			query: "SELECT id FROM tab WHERE id = :id",
			args:  []any{map[string]any{"id": uint(7)}},
		},
		setupExpect: func(m sqlmock.Sqlmock) {
			m.ExpectQuery("SELECT id FROM tab WHERE id = ?").
				WithArgs(uint(7)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
		},
		wantObject: &dummyObject{ID: 7},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatal(err)
			}
			tt.setupExpect(mock)
			object := &dummyObject{}
			err = Select(db, object, tt.args.query, tt.args.args...)
			if err != nil {
				t.Fatalf("Select() error = %v", err)
			}
			if diff := cmp.Diff(tt.wantObject, object); diff != "" {
				t.Fatalf("Select() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
