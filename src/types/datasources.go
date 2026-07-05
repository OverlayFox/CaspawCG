package types

import (
	"fmt"
	"strconv"
	"strings"
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

func NewRange(input string) (Range, error) {
	sheet, body, ok := splitSheetPrefix(input)
	if !ok {
		return Range{}, fmt.Errorf("invalid range format: %s (missing sheet name, expected e.g. 'Sheet1!A1:A10')", input)
	}

	splitStr := strings.Split(body, ":")
	if len(splitStr) != 2 {
		return Range{}, fmt.Errorf("invalid range format: %s", input)
	}

	startCol, startRow, ok := parseCellRef(splitStr[0])
	if !ok {
		return Range{}, fmt.Errorf("invalid range format: %s", input)
	}

	endCol, endRow, ok := parseCellRef(splitStr[1])
	if !ok {
		return Range{}, fmt.Errorf("invalid range format: %s", input)
	}

	if startCol != endCol {
		return Range{}, fmt.Errorf("invalid range format: %s (range must span a single column)", input)
	}

	if endRow < startRow {
		return Range{}, fmt.Errorf("invalid range format: %s (end row before start row)", input)
	}

	locations := make([]Location, 0, endRow-startRow+1)
	for row := startRow; row <= endRow; row++ {
		locations = append(locations, Location{Key: sheetQualifiedKey(sheet, fmt.Sprintf("%s%d", startCol, row))})
	}

	return Range{Locations: locations}, nil
}

// splitSheetPrefix splits a "Sheet1!A1:A10"-style (optionally quoted, e.g. "'CG-System'!A1:A10")
// string into its bare (unquoted) sheet name and the remaining cell-reference body.
// ok is false if no "!"-qualified sheet name is present.
func splitSheetPrefix(input string) (sheet string, body string, ok bool) {
	idx := strings.Index(input, "!")
	if idx <= 0 {
		return "", "", false
	}
	sheet = input[:idx]
	if len(sheet) >= 2 && sheet[0] == '\'' && sheet[len(sheet)-1] == '\'' {
		sheet = sheet[1 : len(sheet)-1]
	}
	return sheet, input[idx+1:], true
}

// sheetQualifiedKey builds a Sheets-API-safe A1 range/cell reference, quoting the sheet
// name since Google rejects unquoted sheet names containing characters like "-" or spaces.
func sheetQualifiedKey(sheet, cellRef string) string {
	return fmt.Sprintf("'%s'!%s", sheet, cellRef)
}

// parseCellRef parses a cell reference such as "A1" into its column letters and row number.
func parseCellRef(cell string) (string, int, bool) {
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

	var colSb strings.Builder
	for _, ch := range cell[:i] {
		if ch >= 'a' && ch <= 'z' {
			ch = ch - ('a' - 'A')
		}
		colSb.WriteRune(ch)
	}
	col := colSb.String()

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

	minCol, maxCol := 0, 0
	minRow, maxRow := 0, 0
	found := false
	sheet := ""

	for _, loc := range r.Locations {
		locSheet, body, ok := splitSheetPrefix(loc.Key)
		if !ok {
			continue
		}
		sheet = locSheet

		col, row, ok := parseCellRef(body)
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
		return sheetQualifiedKey(sheet, start)
	}

	return sheetQualifiedKey(sheet, start+":"+end)
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
