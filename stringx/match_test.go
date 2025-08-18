package stringx

import "testing"

func TestMatchPartial(t *testing.T) {
	tests := []struct {
		name string
		s    string
		by   string
		want int
	}{
		{
			name: "partial match",
			s:    "HalloWorld",
			by:   "Word",
			want: 3,
		}, {
			name: "no match",
			s:    "HalloWorld",
			by:   "Sister",
			want: 0,
		}, {
			name: "full match",
			s:    "HalloWorld",
			by:   "Hallo",
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchPartial(tt.s, tt.by); got != tt.want {
				t.Errorf("MatchPartial() = %v, want %v", got, tt.want)
			}
		})
	}
}
