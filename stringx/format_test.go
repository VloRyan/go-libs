package stringx

import "testing"

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{arg: "Hallo", want: "hallo"},
		{arg: "Hallo World", want: "halloWorld"},
		{arg: "hallo World", want: "halloWorld"},
		{arg: "Hallo_world", want: "halloWorld"},
		{arg: "Hallo-world", want: "halloWorld"},
		{arg: "hallo world", want: "halloWorld"},
		{arg: "URL", want: "url"},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			if got := ToCamelCase(tt.arg); got != tt.want {
				t.Errorf("ToCamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{arg: "Hallo", want: "hallo"},
		{arg: "Hallo World", want: "hallo_world"},
		{arg: "hallo World", want: "hallo_world"},
		{arg: "Hallo_world", want: "hallo_world"},
		{arg: "Hallo-world", want: "hallo_world"},
		{arg: "hallo world", want: "hallo_world"},
		{arg: "URL", want: "url"},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			if got := ToSnakeCase(tt.arg); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{arg: "Hallo", want: "hallo"},
		{arg: "Hallo World", want: "hallo-world"},
		{arg: "hallo World", want: "hallo-world"},
		{arg: "Hallo_world", want: "hallo-world"},
		{arg: "Hallo-world", want: "hallo-world"},
		{arg: "hallo world", want: "hallo-world"},
		{arg: "URL", want: "url"},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			if got := ToKebabCase(tt.arg); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
