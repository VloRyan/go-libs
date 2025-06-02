package reflectx

import (
	"fmt"
	"reflect"
	"testing"
)

func TestElemTypeOf(t *testing.T) {
	tests := []struct {
		arg      any
		wantElem bool
	}{
		{arg: 1},
		{arg: uint(2)},
		{arg: 3.4},
		{arg: "hallo"},
		{arg: []string{"hallo"}, wantElem: true},
		{arg: map[string]string{"hallo": "welt"}, wantElem: true},
		{arg: ptr("hallo"), wantElem: true},
	}
	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.arg).String(), func(t *testing.T) {
			want := reflect.TypeOf(tt.arg)
			if tt.wantElem {
				want = want.Elem()
			}
			if got := ElemTypeOf(tt.arg, true); !reflect.DeepEqual(got, want) {
				t.Errorf("ElemTypeOf() = %v, want %v", got, want)
			}
		})
	}
}

func TestTypeOf(t *testing.T) {
	type args struct {
		v                   any
		resolvePointedValue bool
	}
	tests := []struct {
		args     args
		wantElem bool
	}{
		{args: args{v: 1, resolvePointedValue: false}},
		{args: args{v: 1, resolvePointedValue: true}},
		{args: args{v: ptr("hallo"), resolvePointedValue: false}},
		{args: args{v: ptr("hallo"), resolvePointedValue: true}, wantElem: true},
	}
	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.args.v).String()+","+fmt.Sprintf("%v", tt.args.resolvePointedValue), func(t *testing.T) {
			want := reflect.TypeOf(tt.args.v)
			if tt.wantElem {
				want = want.Elem()
			}
			if got := TypeOf(tt.args.v, tt.args.resolvePointedValue); !reflect.DeepEqual(got, want) {
				t.Errorf("ElemTypeOf() = %v, want %v", got, want)
			}
		})
	}
}

func TestElemValueOf(t *testing.T) {
	tests := []struct {
		arg      any
		wantElem bool
	}{
		{arg: 1},
		{arg: uint(2)},
		{arg: 3.4},
		{arg: "hallo"},
		{arg: []string{"hallo"}},
		{arg: map[string]string{"hallo": "welt"}},
		{arg: ptr("hallo"), wantElem: true},
	}
	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.arg).String(), func(t *testing.T) {
			want := reflect.ValueOf(tt.arg)
			if tt.wantElem {
				want = want.Elem()
			}
			if got := ElemValueOf(tt.arg); !reflect.DeepEqual(got, want) {
				t.Errorf("ElemValueOf() = %v, want %v", got, want)
			}
		})
	}
}

func TestValueOf(t *testing.T) {
	type args struct {
		v                   any
		resolvePointedValue bool
	}
	tests := []struct {
		args     args
		wantElem bool
	}{
		{args: args{v: 1, resolvePointedValue: false}},
		{args: args{v: 1, resolvePointedValue: true}},
		{args: args{v: ptr("hallo"), resolvePointedValue: false}},
		{args: args{v: ptr("hallo"), resolvePointedValue: true}, wantElem: true},
	}
	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt.args.v).String()+","+fmt.Sprintf("%v", tt.args.resolvePointedValue), func(t *testing.T) {
			want := reflect.ValueOf(tt.args.v)
			if tt.wantElem {
				want = want.Elem()
			}
			if got := ValueOf(tt.args.v, tt.args.resolvePointedValue); !reflect.DeepEqual(got, want) {
				t.Errorf("ElemTypeOf() = %v, want %v", got, want)
			}
		})
	}
}
