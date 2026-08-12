package update

import (
	"errors"

	"github.com/overlayfox/caspaw-cg/src/types"
)

type CasparMaps map[string]*Resolver // [casparKey]Resolver

func (c *CasparMaps) GetElementsAmount() int {
	maxIdx := 0
	for _, resolver := range *c {
		if resolver.slot > maxIdx {
			maxIdx = resolver.slot
		}
	}
	return maxIdx + 1 // +1 because offset is 0-indexed
}

func (c *CasparMaps) GetBySlot(slot int) (CasparMaps, error) {
	casparMapsBySlot := make(CasparMaps)
	for casparKey, resolver := range *c {
		if resolver.slot == slot {
			casparMapsBySlot[casparKey] = resolver
		}
	}

	if len(casparMapsBySlot) == 0 {
		return nil, errors.New("no caspar maps found for the given slot")
	}

	return casparMapsBySlot, nil
}

type Resolver struct {
	liveOffset int // the offset currently in the list of locations
	slot       int // immutable display-slot index (0..N-1) within a Schedule's visible window

	datasource types.DataSource
	dataRange  types.Range
}

func NewResolver(datasource types.DataSource, dataRange types.Range, offset int) Resolver {
	return Resolver{
		datasource: datasource,
		dataRange:  dataRange,
		liveOffset: offset,
		slot:       offset,
	}
}

func (r *Resolver) GetRowData(offset int) (any, error) {
	if offset < 0 || offset >= len(r.dataRange.Locations) {
		return nil, errors.New("offset out of range")
	}
	key := r.dataRange.Locations[offset].Key
	data, err := r.datasource.Get(key)
	if err != nil {
		return nil, err
	}
	return data.Value, nil
}

func (r *Resolver) GetData() (any, error) {
	if r.liveOffset < 0 || r.liveOffset >= len(r.dataRange.Locations) {
		return nil, errors.New("offset out of range")
	}
	key := r.dataRange.Locations[r.liveOffset].Key
	data, err := r.datasource.Get(key)
	if err != nil {
		return nil, err
	}
	return data.Value, nil
}

func (r *Resolver) GetRowAmount() int {
	return len(r.dataRange.Locations)
}

func (r *Resolver) GetSlot() int {
	return r.slot
}

func (r *Resolver) Advance() {
	r.liveOffset++
	if r.liveOffset >= len(r.dataRange.Locations) {
		r.liveOffset = 0 // Reset to the beginning if we reach the end
	}
}
