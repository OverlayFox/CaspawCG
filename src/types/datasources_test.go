package types

import "testing"

func TestFieldSubscriptionResolveLocations(t *testing.T) {
	t.Run("datasource input type resolves a single normalized location", func(t *testing.T) {
		sub := FieldSubscription{
			Source:    "sheet1",
			Type:      DataTypeString,
			InputType: "datasource",
			Location:  "Sheet1!A1",
		}
		identifier, locations, err := sub.ResolveLocations()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "'Sheet1'!A1"
		if identifier != want {
			t.Errorf("identifier = %q, want %q", identifier, want)
		}
		if len(locations) != 1 || locations[0].Key != want {
			t.Errorf("locations = %+v, want a single location with key %q", locations, want)
		}
	})

	t.Run("range input type resolves the location at the given offset", func(t *testing.T) {
		sub := FieldSubscription{
			Source:    "sheet1",
			Type:      DataTypeInt,
			InputType: "range",
			Range:     "Sheet1!A1:A3",
			Offset:    1,
		}
		identifier, locations, err := sub.ResolveLocations()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if identifier != "'Sheet1'!A2" {
			t.Errorf("identifier = %q, want 'Sheet1'!A2", identifier)
		}
		if len(locations) != 3 {
			t.Errorf("locations = %+v, want 3 locations primed for the whole range", locations)
		}
	})

	t.Run("offset out of range is an error", func(t *testing.T) {
		sub := FieldSubscription{
			Source:    "sheet1",
			InputType: "range",
			Range:     "Sheet1!A1:A2",
			Offset:    5,
		}
		if _, _, err := sub.ResolveLocations(); err == nil {
			t.Error("expected an error for an out-of-range offset, got nil")
		}
	})

	t.Run("invalid range propagates the parse error", func(t *testing.T) {
		sub := FieldSubscription{InputType: "range", Range: "not-a-range"}
		if _, _, err := sub.ResolveLocations(); err == nil {
			t.Error("expected an error for a malformed range, got nil")
		}
	})
}
