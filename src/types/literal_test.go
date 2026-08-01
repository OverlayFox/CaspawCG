package types

import "testing"

func TestBuildDataMap(t *testing.T) {
	data := BuildDataMap([]LiteralField{
		{CasparKey: "a", Type: DataTypeInt, RawValue: "42"},
		{CasparKey: "b", Type: DataTypeInt, RawValue: "not-a-number"},
		{CasparKey: "c", Type: DataTypeFloat, RawValue: "3.14"},
		{CasparKey: "d", Type: DataTypeFloat, RawValue: "not-a-number"},
		{CasparKey: "e", Type: DataTypeString, RawValue: "hello"},
		{CasparKey: "f", Type: DataTypePath, RawValue: "/some/path"},
	})

	cases := map[string]any{
		"a": 42,
		"b": 0,
		"c": 3.14,
		"d": 0.0,
		"e": "hello",
		"f": "/some/path",
	}
	for key, want := range cases {
		if got := data[key]; got != want {
			t.Errorf("data[%q] = %v (%T), want %v (%T)", key, got, got, want, want)
		}
	}
}
