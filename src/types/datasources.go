package types

import (
	"fmt"
	"strconv"
	"unicode"
)

// DataType represents the type of data being handles
type DataType string

const (
	DataTypeInt    DataType = "int"
	DataTypeString DataType = "string"
	DataTypeFloat  DataType = "float"
	DataTypePath   DataType = "path"
	DataTypeURL    DataType = "url"
)

// Location represents a location in a data source, identified by a key and its data type.
type Location struct {
	Key  string
	Type DataType
}

// Range represents a range of data in a data source, identified by a key and its data type.
type Range struct {
	Locations []Location
}

func (r Range) Key() string {
	if len(r.Locations) == 0 {
		return ""
	}

	colToNum := func(col string) int {
		n := 0
		for _, ch := range col {
			n = n*26 + int(ch-'A'+1)
		}
		return n
	}

	numToCol := func(n int) string {
		if n <= 0 {
			return ""
		}
		out := ""
		for n > 0 {
			n--
			out = string(rune('A'+(n%26))) + out
			n /= 26
		}
		return out
	}

	parseCell := func(cell string) (string, int, bool) {
		if cell == "" {
			return "", 0, false
		}

		i := 0
		for i < len(cell) && unicode.IsLetter(rune(cell[i])) {
			i++
		}
		if i == 0 || i == len(cell) {
			return "", 0, false
		}

		col := ""
		for _, ch := range cell[:i] {
			if !unicode.IsLetter(ch) {
				return "", 0, false
			}
			if ch >= 'a' && ch <= 'z' {
				ch = ch - ('a' - 'A')
			}
			col += string(ch)
		}

		for _, ch := range cell[i:] {
			if !unicode.IsDigit(ch) {
				return "", 0, false
			}
		}

		row, err := strconv.Atoi(cell[i:])
		if err != nil || row <= 0 {
			return "", 0, false
		}

		return col, row, true
	}

	minCol, maxCol := 0, 0
	minRow, maxRow := 0, 0
	found := false

	for _, loc := range r.Locations {
		col, row, ok := parseCell(loc.Key)
		if !ok {
			continue
		}

		colNum := colToNum(col)
		if !found {
			minCol, maxCol = colNum, colNum
			minRow, maxRow = row, row
			found = true
			continue
		}

		if colNum < minCol {
			minCol = colNum
		}
		if colNum > maxCol {
			maxCol = colNum
		}
		if row < minRow {
			minRow = row
		}
		if row > maxRow {
			maxRow = row
		}
	}

	if !found {
		return ""
	}

	start := fmt.Sprintf("%s%d", numToCol(minCol), minRow)
	end := fmt.Sprintf("%s%d", numToCol(maxCol), maxRow)
	if start == end {
		return start
	}

	return start + ":" + end
}

// Data represents a piece of data retrieved from a data source, including its location and value.
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

	Close()
}
