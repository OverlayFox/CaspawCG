package update

import (
	"errors"

	"github.com/overlayfox/caspaw-cg/src/types"
)

type Resolver struct {
	offset     int
	datasource types.DataSource
	dataRange  types.Range
}

func NewResolver(datasource types.DataSource, dataRange types.Range, offset int) Resolver {
	return Resolver{
		datasource: datasource,
		dataRange:  dataRange,
		offset:     offset,
	}
}

func (r *Resolver) GetData() (any, error) {
	if r.offset < 0 || r.offset >= len(r.dataRange.Locations) {
		return nil, errors.New("offset out of range")
	}
	key := r.dataRange.Locations[r.offset].Key
	data, err := r.datasource.Get(key)
	if err != nil {
		return nil, err
	}
	return data.Value, nil
}

func (r *Resolver) GetAllData() ([]any, error) {
	if r.offset < 0 || r.offset >= len(r.dataRange.Locations) {
		return nil, errors.New("offset out of range")
	}
	allData := make([]any, len(r.dataRange.Locations))
	for i, location := range r.dataRange.Locations {
		data, err := r.datasource.Get(location.Key)
		if err != nil {
			return nil, err
		}
		allData[i] = data.Value
	}
	return allData, nil
}

func (r *Resolver) Advance() {
	r.offset++
	if r.offset >= len(r.dataRange.Locations) {
		r.offset = 0 // Reset to the beginning if we reach the end
	}
}
