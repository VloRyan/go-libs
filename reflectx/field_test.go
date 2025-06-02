package reflectx

import (
	"reflect"
	"testing"
)

type testStruct struct {
	BaseStruct
	Int       int
	Uint      uint
	Float     float64
	String    string
	IntPtr    *int
	UintPtr   *uint
	FloatPtr  *float64
	StringPtr *string

	Slice   []int
	Map     map[string]string
	Another AnotherStruct
	Sub     *BaseStruct
}
type AnotherStruct struct {
	BaseStruct
	FieldB string
}
type BaseStruct struct {
	FieldA string
}

func TestFindField(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "direct field", path: "Int"},
		{name: "embedded field", path: "FieldA"},
		{name: "nested field", path: "Another.FieldB"},
		{name: "nested embedded field", path: "Another.FieldA"},
		{name: "nested ptr field", path: "Sub.FieldA"},
		{name: "indexed field", path: "Slice[1]"},
		{name: "map field", path: "Map[Key]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &testStruct{
				Sub:   &BaseStruct{},
				Slice: []int{1, 2, 3},
				Map:   map[string]string{"Key": "Value"},
			}
			got := FindField(s, tt.path)
			if !got.IsValid() {
				t.Errorf("FindField() field not found")
			}
		})
	}
}

func TestSetFieldValue(t *testing.T) {
	tests := []struct {
		fieldName string
		value     any
		wantErr   bool
	}{
		// Direct
		{fieldName: "Int", value: 3},
		{fieldName: "Uint", value: uint(4)},
		{fieldName: "Float", value: 5.6},
		{fieldName: "String", value: "Hallo world"},
		{fieldName: "IntPtr", value: ptr(3)},
		{fieldName: "UintPtr", value: ptr(uint(4))},
		{fieldName: "FloatPtr", value: ptr(5.6)},
		{fieldName: "StringPtr", value: ptr("Hallo world")},
		{fieldName: "Slice", value: []int{1, 2, 3}},
		{fieldName: "Map", value: map[string]string{"Hallo": "Welt"}},

		// Conversions
		{fieldName: "Int", value: uint(4)},
		{fieldName: "Int", value: 5.6},
		{fieldName: "Uint", value: 3},
		{fieldName: "Uint", value: 5.6},
		{fieldName: "Float", value: 3},
		{fieldName: "Float", value: uint(4)},
		{fieldName: "String", value: 4},
		{fieldName: "StringPtr", value: "Hallo world"},
		{fieldName: "String", value: ptr("Hallo world")},
		{fieldName: "Int", value: ptr(uint(4))},
		{fieldName: "UintPtr", value: uint(4)},
		{fieldName: "Int", value: "80"},
	}
	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.value).String()+"->"+tt.fieldName, func(t *testing.T) {
			s := &testStruct{}
			sv := reflect.ValueOf(s)
			field := sv.Elem().FieldByName(tt.fieldName)
			if err := SetFieldValue(field, tt.value); (err != nil) != tt.wantErr {
				t.Errorf("SetFieldValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
