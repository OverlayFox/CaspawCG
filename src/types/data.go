package types

type DataSource interface {
	// Adds the specified locations to the data source for priming.
	// Location string example: "sheet1!A1" - currently only supports single fields
	Prime(locations []string) error
	// Removes the specified locations from the data source's primed data.
	RemovePrime(locations []string) error

	// Retrieves the data for the specified location.
	Get(location string) (any, error)

	Close()
}
