package types

// DataType represents the type of data being handles
type DataType string

const (
	DataTypeInt    DataType = "int"
	DataTypeString DataType = "string"
	DataTypeFloat  DataType = "float"
	DataTypePath   DataType = "path"
	DataTypeURL    DataType = "url"
)

type Location struct {
	Key  string
	Type DataType
}

type Data struct {
	Location
	Value any
}

type DataSource interface {
	// Adds the specified locations to the data source for priming.
	// Location string example: "sheet1!A1" - currently only supports single fields
	Prime(locations []Location) error
	// Removes the specified locations from the data source's primed data.
	RemovePrime(keys []string) error

	// Retrieves the data for the specified location.
	Get(key string) (Data, error)

	Close()
}
