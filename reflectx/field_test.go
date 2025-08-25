package reflectx

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	BaseStruct
	Int       int
	Uint      uint
	Float     float64
	String    string
	Bool      bool
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
		expected  any
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
		{fieldName: "Int", value: uint(4), expected: 4},
		{fieldName: "Int", value: 5.6, expected: 5},
		{fieldName: "Uint", value: 3, expected: uint(3)},
		{fieldName: "Uint", value: 5.6, expected: uint(5)},
		{fieldName: "Float", value: 3, expected: 3.0},
		{fieldName: "Float", value: uint(4), expected: 4.0},
		{fieldName: "String", value: 4, expected: "\x04"},
		{fieldName: "StringPtr", value: "Hallo world", expected: ptr("Hallo world")},
		{fieldName: "String", value: ptr("Hallo world"), expected: "Hallo world"},
		{fieldName: "Int", value: ptr(uint(4)), expected: 4},
		{fieldName: "UintPtr", value: uint(4), expected: ptr(uint(4))},
		{fieldName: "Int", value: "80", expected: 80},
		{fieldName: "Uint", value: "80", expected: uint(80)},
		{fieldName: "Bool", value: 1, expected: true},
		{fieldName: "Bool", value: uint(1), expected: true},
		{fieldName: "Bool", value: float64(1), expected: true},
		{fieldName: "Bool", value: "true", expected: true},
	}
	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.value).String()+"->"+tt.fieldName, func(t *testing.T) {
			s := &testStruct{}
			sv := reflect.ValueOf(s)
			field := sv.Elem().FieldByName(tt.fieldName)
			if err := SetFieldValue(field, tt.value); (err != nil) != tt.wantErr {
				t.Errorf("SetFieldValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			expected := reflect.ValueOf(tt.value)
			ve := reflect.ValueOf(tt.expected)
			if (ve.Kind() != reflect.Ptr && ve.IsValid() && !ve.IsZero()) ||
				(ve.Kind() == reflect.Ptr && !ve.IsNil()) {
				expected = ve
			}
			if diff := cmp.Diff(expected.Interface(), field.Interface()); diff != "" {
				t.Errorf("SetFieldValue() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

type timeObject struct {
	Time     time.Time
	Date     time.Time `time_format:"DateOnly"`
	Custom   time.Time `time_format:"02.01.2006 15:04:05"`
	TimeOnly time.Time `time_format:"TimeOnly"`
	TimePtr  *time.Time
}

func TestSetTimeField(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		valueFormat string
		field       string
		wantErr     assert.ErrorAssertionFunc
	}{{
		name:    "RFC3339 Time",
		value:   "2025-08-11T12:01:02Z",
		field:   "Time",
		wantErr: assert.NoError,
	}, {
		name:        "Date",
		value:       "2025-08-11",
		valueFormat: time.DateOnly,
		field:       "Date",
		wantErr:     assert.NoError,
	}, {
		name:    "With timezone",
		value:   "2025-08-11T12:01:02+02:00",
		field:   "Time",
		wantErr: assert.NoError,
	}, {
		name:    "RFC3339 TimePtr",
		value:   "2025-08-11T12:01:02Z",
		field:   "TimePtr",
		wantErr: assert.NoError,
	}, {
		name:        "Custom",
		value:       "11.08.2025 12:01:02",
		valueFormat: "02.01.2006 15:04:05",
		field:       "Custom",
		wantErr:     assert.NoError,
	}, {
		name:        "TimeOnly",
		value:       "12:01:02",
		valueFormat: "15:04:05",
		field:       "TimeOnly",
		wantErr:     assert.NoError,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &timeObject{}
			structField, _ := TypeOf(obj, true).FieldByName(tt.field)
			fieldValue := ValueOf(obj, true).FieldByName(tt.field)
			tt.wantErr(t, SetTimeField(tt.value, structField, fieldValue), "SetTimeField()")
			timeFormat := tt.valueFormat
			if timeFormat == "" {
				timeFormat = time.RFC3339
			}
			parsedValue, err := time.Parse(timeFormat, tt.value)
			if !tt.wantErr(t, err) {
				t.Fatal(err)
			}
			if DeRefValue(fieldValue).Interface().(time.Time) != parsedValue {
				t.Errorf("SetTimeField() want: %s, got: %s", parsedValue, fieldValue.Interface().(time.Time))
			}
		})
	}
}
