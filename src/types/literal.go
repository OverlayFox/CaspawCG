package types

import "strconv"

// LiteralField is a single template field value entered or resolved on the frontend,
// still in its raw string form. BuildDataMap coerces it to the declared Type.
type LiteralField struct {
	CasparKey string
	Type      DataType
	RawValue  string
}

// BuildDataMap coerces each field's RawValue to its declared Type, falling back to
// the type's zero value on parse failure (matching how the field was always entered
// as plain text and may not yet represent a complete number while the user is typing).
func BuildDataMap(fields []LiteralField) map[string]any {
	data := make(map[string]any, len(fields))
	for _, f := range fields {
		switch f.Type {
		case DataTypeInt:
			n, err := strconv.Atoi(f.RawValue)
			if err != nil {
				n = 0
			}
			data[f.CasparKey] = n
		case DataTypeFloat:
			v, err := strconv.ParseFloat(f.RawValue, 64)
			if err != nil {
				v = 0.0
			}
			data[f.CasparKey] = v
		default:
			data[f.CasparKey] = f.RawValue
		}
	}
	return data
}
