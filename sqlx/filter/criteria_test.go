package filter

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEmptyCriteria_ToWhere(t *testing.T) {
	tests := []struct {
		name     string
		criteria Criteria
		want     Where
		wantErr  bool
	}{{
		name:     "Empty",
		criteria: &EmptyCriteria{},
		want:     Where{},
	}, {
		name:     "Empty and empty",
		criteria: New().And(&EmptyCriteria{}),
		want:     Where{},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.criteria.ToWhere("")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToWhere() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnaryCriteria_ToWhere(t *testing.T) {
	tests := []struct {
		name      string
		criteria  *UnaryCriteria
		tableName string
		want      Where
		wantErr   bool
	}{{
		name: "eq",
		criteria: &UnaryCriteria{
			OpType:     EqOp,
			ColumnExpr: "field",
			ValueExpr:  ":field",
		},
		tableName: "MyTable",
		want: Where{
			Clause: "MyTable.field = :field",
		},
	}, {
		name: "Between",
		criteria: &UnaryCriteria{
			OpType:     BetweenOp,
			ColumnExpr: "field",
			ValueExpr:  ":field",
			Parameter:  map[string]any{"field_0": 0, "field_1": 1},
		},
		tableName: "table",
		want: Where{
			Clause:    "table.field BETWEEN :field_0 AND :field_1",
			Parameter: map[string]any{"field_0": 0, "field_1": 1},
		},
	}, {
		name: "In",
		criteria: &UnaryCriteria{
			OpType:     InOp,
			ColumnExpr: "field",
			Parameter: map[string]any{
				"field_0": 0,
				"field_1": 1,
				"field_2": 3,
			},
		},
		tableName: "table",
		want: Where{
			Clause: "table.field IN (:field_0, :field_1, :field_2)",
			Parameter: map[string]any{
				"field_0": 0,
				"field_1": 1,
				"field_2": 3,
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.criteria.ToWhere(tt.tableName)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToWhere() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNotCriteria_ToWhere(t *testing.T) {
	tests := []struct {
		name     string
		criteria *NotCriteria
		want     Where
		wantErr  bool
	}{{
		name: "not empty",
		criteria: &NotCriteria{
			C: &EmptyCriteria{},
		},
		want: Where{
			Clause: "NOT ()",
		},
	}, {
		name: "not not",
		criteria: &NotCriteria{
			C: &NotCriteria{
				C: &EmptyCriteria{},
			},
		},
		want: Where{
			Clause: "NOT (NOT ())",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.criteria.ToWhere("")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToWhere() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBinaryCriteria_ToWhere(t *testing.T) {
	tests := []struct {
		name     string
		criteria *BinaryCriteria
		want     Where
		wantErr  bool
	}{{
		name: "and",
		criteria: &BinaryCriteria{
			First:  &EmptyCriteria{},
			Second: &EmptyCriteria{},
			Conn:   ConnOpAnd,
		},
		want: Where{
			Clause: "() AND ()",
		},
	}, {
		name: "or",
		criteria: &BinaryCriteria{
			First:  &EmptyCriteria{},
			Second: &EmptyCriteria{},
			Conn:   ConnOpOr,
		},
		want: Where{
			Clause: "() OR ()",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.criteria.ToWhere("")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToWhere() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
