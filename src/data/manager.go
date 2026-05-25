package data

import (
	"fmt"
	"slices"
	"sync"
)

// manager manages all datasources given to the application
type manager struct {
	cfg *Config

	dataSources []DataSource
	mtx         sync.RWMutex
}

func NewManager(cfg *Config) DatasourceManager {
	return &manager{
		cfg:         cfg,
		dataSources: make([]DataSource, 0),
	}
}

func (m *manager) AddDataSource(ds DataSource) error {
	names := m.GetDataSourceNames()
	if slices.Contains(names, ds.GetName()) {
		return fmt.Errorf("datasource with name '%s' already exists", ds.GetName())
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.dataSources = append(m.dataSources, ds)

	return nil
}

func (m *manager) RemoveDataSource(name string) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for i, ds := range m.dataSources {
		if ds.GetName() == name {
			m.dataSources = append(m.dataSources[:i], m.dataSources[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("datasource with name '%s' not found", name)
}

func (m *manager) GetDataSource(name string) (DataSource, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	for _, ds := range m.dataSources {
		if ds.GetName() == name {
			return ds, nil
		}
	}

	return nil, fmt.Errorf("datasource with name '%s' not found", name)
}

func (m *manager) GetDataSourceNames() []string {
	m.mtx.RLock()
	names := make([]string, 0, len(m.dataSources))
	for _, ds := range m.dataSources {
		names = append(names, ds.GetName())
	}
	m.mtx.RUnlock()

	return names
}
