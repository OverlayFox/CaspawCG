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
	// GetName returns the identifier of the Datasource
	GetName() string

	// Prime adds the specified locations to the data source for priming.
	// Location string example: "sheet1!A1" - currently only supports single fields
	Prime(locations []Location) error
	// RemovePrime removes the specified locations from the data source's primed data.
	RemovePrime(keys []string) error

	// Get retrieves the data for the specified location.
	Get(key string) (Data, error)

	// Close closes the datasource
	Close()
}

type DatasourceManager interface {
	// AddDataSource adds a datasource
	AddDataSource(ds DataSource) error
	// RemoveDataSource removes a datasource
	RemoveDataSource(name string) error
	// GetDataSource returns a datasource by name
	GetDataSource(name string) (DataSource, error)

	// UI functions
	// GetDataSourceNames returns the names of all datasources
	GetDataSourceNames() []string
}
