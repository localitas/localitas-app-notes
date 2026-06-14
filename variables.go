package notes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type VariableRegistry struct {
	Version   string               `json:"version"`
	UpdatedAt time.Time            `json:"updated_at"`
	Variables map[string]*Variable `json:"variables"`
}

type Variable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Source      string      `json:"source"`
	Schema      *DataSchema `json:"schema,omitempty"`
	Value       interface{} `json:"value,omitempty"`
	DefinedIn   string      `json:"defined_in"`
	LastUpdated time.Time   `json:"last_updated"`
}

type DataSchema struct {
	Columns   []*ColumnSchema `json:"columns"`
	Rows      int             `json:"rows"`
	SizeBytes int64           `json:"size_bytes"`
}

type ColumnSchema struct {
	Name         string        `json:"name"`
	DType        string        `json:"dtype"`
	Nullable     bool          `json:"nullable"`
	SampleValues []interface{} `json:"sample_values,omitempty"`
	Min          interface{}   `json:"min,omitempty"`
	Max          interface{}   `json:"max,omitempty"`
	Mean         float64       `json:"mean,omitempty"`
	Sum          interface{}   `json:"sum,omitempty"`
}

func NewVariableRegistry() *VariableRegistry {
	return &VariableRegistry{
		Version:   "1.0",
		UpdatedAt: time.Now().UTC(),
		Variables: make(map[string]*Variable),
	}
}

func LoadVariableRegistry(dataDir string) (*VariableRegistry, error) {
	path := filepath.Join(dataDir, "variables.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return NewVariableRegistry(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var registry VariableRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}
	return &registry, nil
}

func (r *VariableRegistry) AddVariable(name, varType, source, definedIn string, schema *DataSchema, value interface{}) {
	r.Variables[name] = &Variable{
		Name:        name,
		Type:        varType,
		Source:      source,
		Schema:      schema,
		Value:       value,
		DefinedIn:   definedIn,
		LastUpdated: time.Now().UTC(),
	}
}

func (r *VariableRegistry) GetVariable(name string) (*Variable, bool) {
	v, ok := r.Variables[name]
	return v, ok
}

func (r *VariableRegistry) HasVariable(name string) bool {
	_, ok := r.Variables[name]
	return ok
}

func (r *VariableRegistry) RemoveVariable(name string) {
	delete(r.Variables, name)
}

func (r *VariableRegistry) GetVariablesByBlock(blockID string) []*Variable {
	vars := make([]*Variable, 0)
	for _, v := range r.Variables {
		if v.DefinedIn == blockID {
			vars = append(vars, v)
		}
	}
	return vars
}

func SaveVariableRegistry(dataDir string, registry *VariableRegistry) error {
	registry.UpdatedAt = time.Now().UTC()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "variables.json"), data, 0644)
}
