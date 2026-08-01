package types

import (
	"reflect"
	"testing"
)

func TestParseChannelExpression(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{name: "blank means all channels", input: "", want: nil},
		{name: "whitespace only means all channels", input: "   ", want: nil},
		{name: "single channel", input: "1", want: []int{1}},
		{name: "list", input: "1,2,4", want: []int{1, 2, 4}},
		{name: "range", input: "1-3", want: []int{1, 2, 3}},
		{name: "mixed list and range", input: "1,3-5", want: []int{1, 3, 4, 5}},
		{name: "dedups overlapping entries", input: "1,1-2", want: []int{1, 2}},
		{name: "sorts unordered input", input: "5,1,3", want: []int{1, 3, 5}},
		{name: "tolerates surrounding whitespace", input: " 1 , 3-4 ", want: []int{1, 3, 4}},
		{name: "invalid token", input: "abc", wantErr: true},
		{name: "reversed range", input: "5-1", wantErr: true},
		{name: "trailing comma", input: "1,", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChannelExpression(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseChannelExpression(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseChannelExpression(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
