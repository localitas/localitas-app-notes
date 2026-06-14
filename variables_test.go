package notes

import (
	"testing"
	"time"
)

func TestNewVariableRegistry(t *testing.T) {
	registry := NewVariableRegistry()
	if registry.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", registry.Version)
	}
	if registry.Variables == nil {
		t.Error("Variables map should be initialized")
	}
	if len(registry.Variables) != 0 {
		t.Error("Variables map should be empty initially")
	}
}

func TestAddVariable(t *testing.T) {
	registry := NewVariableRegistry()
	schema := &DataSchema{
		Columns:   []*ColumnSchema{{Name: "id", DType: "int64"}, {Name: "name", DType: "string"}},
		Rows:      100,
		SizeBytes: 5000,
	}
	registry.AddVariable("test_data", "dataframe", "csv:data/test.csv", "block_0", schema, nil)

	if len(registry.Variables) != 1 {
		t.Fatalf("Expected 1 variable, got %d", len(registry.Variables))
	}
	v := registry.Variables["test_data"]
	if v.Name != "test_data" {
		t.Errorf("Expected name 'test_data', got '%s'", v.Name)
	}
	if v.Type != "dataframe" {
		t.Errorf("Expected type 'dataframe', got '%s'", v.Type)
	}
	if v.Schema == nil || len(v.Schema.Columns) != 2 {
		t.Error("Schema should have 2 columns")
	}
}

func TestGetVariable(t *testing.T) {
	registry := NewVariableRegistry()
	registry.AddVariable("my_var", "int", "js:block_1", "block_1", nil, 42)

	v, exists := registry.GetVariable("my_var")
	if !exists {
		t.Fatal("Variable 'my_var' should exist")
	}
	if v.Value != 42 {
		t.Errorf("Expected value 42, got %v", v.Value)
	}

	_, exists = registry.GetVariable("nonexistent")
	if exists {
		t.Error("Variable 'nonexistent' should not exist")
	}
}

func TestHasVariable(t *testing.T) {
	registry := NewVariableRegistry()
	registry.AddVariable("exists", "string", "js:block_0", "block_0", nil, "value")

	if !registry.HasVariable("exists") {
		t.Error("Variable 'exists' should exist")
	}
	if registry.HasVariable("not_exists") {
		t.Error("Variable 'not_exists' should not exist")
	}
}

func TestRemoveVariable(t *testing.T) {
	registry := NewVariableRegistry()
	registry.AddVariable("to_remove", "int", "js:block_0", "block_0", nil, 123)

	registry.RemoveVariable("to_remove")
	if registry.HasVariable("to_remove") {
		t.Error("Variable should not exist after removal")
	}
}

func TestGetVariablesByBlock(t *testing.T) {
	registry := NewVariableRegistry()
	registry.AddVariable("var1", "int", "js:block_0", "block_0", nil, 1)
	registry.AddVariable("var2", "int", "js:block_0", "block_0", nil, 2)
	registry.AddVariable("var3", "int", "js:block_1", "block_1", nil, 3)

	vars := registry.GetVariablesByBlock("block_0")
	if len(vars) != 2 {
		t.Errorf("Expected 2 variables from block_0, got %d", len(vars))
	}
	vars = registry.GetVariablesByBlock("block_1")
	if len(vars) != 1 {
		t.Errorf("Expected 1 variable from block_1, got %d", len(vars))
	}
	vars = registry.GetVariablesByBlock("block_99")
	if len(vars) != 0 {
		t.Errorf("Expected 0 variables from non-existent block, got %d", len(vars))
	}
}

func TestVariableRegistryPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	registry := NewVariableRegistry()
	registry.AddVariable("data", "dataframe", "csv:data/test.csv", "block_0", &DataSchema{
		Columns:   []*ColumnSchema{{Name: "id", DType: "int64"}},
		Rows:      10,
		SizeBytes: 100,
	}, nil)

	if err := SaveVariableRegistry(tmpDir, registry); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	loaded, err := LoadVariableRegistry(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}
	if len(loaded.Variables) != 1 {
		t.Fatalf("Expected 1 variable, got %d", len(loaded.Variables))
	}
	v := loaded.Variables["data"]
	if v == nil {
		t.Fatal("Variable 'data' should exist")
	}
	if v.Type != "dataframe" {
		t.Errorf("Expected type 'dataframe', got '%s'", v.Type)
	}
	if v.Schema == nil {
		t.Error("Schema should not be nil")
	}
}

func TestDataSchema(t *testing.T) {
	schema := &DataSchema{
		Columns: []*ColumnSchema{
			{Name: "id", DType: "int64", Nullable: false, Min: 1, Max: 100},
			{Name: "price", DType: "float64", Nullable: false, Mean: 125.50},
			{Name: "name", DType: "string", Nullable: true, SampleValues: []interface{}{"Alice", "Bob", "Charlie"}},
		},
		Rows:      1000,
		SizeBytes: 50000,
	}

	if len(schema.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(schema.Columns))
	}
	if schema.Columns[0].DType != "int64" {
		t.Errorf("Expected dtype 'int64', got '%s'", schema.Columns[0].DType)
	}
	if schema.Columns[1].Mean != 125.50 {
		t.Errorf("Expected mean 125.50, got %v", schema.Columns[1].Mean)
	}
	if !schema.Columns[2].Nullable {
		t.Error("String column should be nullable")
	}
}

func TestVariableTimestamps(t *testing.T) {
	registry := NewVariableRegistry()
	before := time.Now()
	registry.AddVariable("test", "int", "js:block_0", "block_0", nil, 42)
	after := time.Now()

	v := registry.Variables["test"]
	if v.LastUpdated.Before(before) || v.LastUpdated.After(after) {
		t.Error("LastUpdated timestamp not in expected range")
	}
}
