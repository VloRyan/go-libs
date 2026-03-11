package reflectx

import (
	"reflect"
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	type args struct {
		v reflect.Value
		t reflect.Type
	}
	tests := []struct {
		name        string
		args        args
		want        reflect.Value
		wantSuccess bool
	}{
		{
			name: "interface{}->int",
			args: args{
				v: reflect.ValueOf(any(3.0)),
				t: reflect.TypeOf(0),
			},
			want:        reflect.ValueOf(3),
			wantSuccess: true,
		},
		{
			name: "int64{}->time.Time",
			args: args{
				v: reflect.ValueOf(int64(1774915200)),
				t: reflect.TypeOf(time.Time{}),
			},
			want:        reflect.ValueOf(time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)),
			wantSuccess: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotSuccess := Convert(tt.args.v, tt.args.t)
			if gotSuccess != tt.wantSuccess {
				t.Fatalf("Convert() gotSuccess = %v, wantSuccess %v", gotSuccess, tt.wantSuccess)
			}
			if !reflect.DeepEqual(got.Interface(), tt.want.Interface()) {
				t.Errorf("Convert() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToNumber(t *testing.T) {
	type args struct {
		s    string
		kind reflect.Kind
	}
	tests := []struct {
		name  string
		args  args
		want  reflect.Value
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := convertToNumber(tt.args.s, tt.args.kind)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToNumber() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("convertToNumber() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_isNumeric(t *testing.T) {
	type args struct {
		k reflect.Kind
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNumeric(tt.args.k); got != tt.want {
				t.Errorf("isNumeric() = %v, want %v", got, tt.want)
			}
		})
	}
}
