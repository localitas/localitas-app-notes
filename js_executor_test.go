package notes

import (
	"testing"
)

func TestNewJSExecutor(t *testing.T) {
	executor := NewJSExecutor()
	if executor == nil {
		t.Fatal("NewJSExecutor should return non-nil executor")
	}
	if executor.vm == nil {
		t.Fatal("Executor should have initialized VM")
	}
}

func TestExecuteJSBlock_BasicTypes(t *testing.T) {
	registry := NewVariableRegistry()

	tests := []struct {
		name         string
		code         string
		varName      string
		expectedType string
	}{
		{"integer result", "42", "result", "int"},
		{"float result", "3.14", "pi", "float"},
		{"string result", "'hello world'", "greeting", "string"},
		{"boolean result", "true", "flag", "bool"},
		{"array result", "[1, 2, 3]", "numbers", "array"},
		{"object result", "({a: 1, b: 2})", "obj", "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := &CodeBlock{
				ID:       "block_0",
				Language: "js",
				VarName:  tt.varName,
				Content:  tt.code,
			}

			variable, output, err := ExecuteJSBlock(block, registry)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if output == "" {
				t.Error("Expected non-empty output")
			}
			if variable == nil {
				t.Fatal("Expected variable to be set")
			}
			if variable.Type != tt.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tt.expectedType, variable.Type)
			}
			if variable.Name != tt.varName {
				t.Errorf("Expected name '%s', got '%s'", tt.varName, variable.Name)
			}
		})
	}
}

func TestExecuteJSBlock_NoVarName(t *testing.T) {
	registry := NewVariableRegistry()
	block := &CodeBlock{
		ID:       "block_0",
		Language: "js",
		Content:  "1 + 1",
	}

	variable, output, err := ExecuteJSBlock(block, registry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if output != "2" {
		t.Errorf("Expected output '2', got '%s'", output)
	}
	if variable != nil {
		t.Error("Expected nil variable when no varName")
	}
}

func TestExecuteJSBlock_VariableInjection(t *testing.T) {
	registry := NewVariableRegistry()
	registry.Variables["x"] = &Variable{Name: "x", Type: "int", Value: 10}

	block := &CodeBlock{
		ID:       "block_0",
		Language: "js",
		VarName:  "doubled",
		Content:  "x * 2",
	}

	variable, _, err := ExecuteJSBlock(block, registry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if variable == nil {
		t.Fatal("Expected variable")
	}
	if val, ok := variable.Value.(int64); !ok || val != 20 {
		t.Errorf("Expected value 20, got %v (%T)", variable.Value, variable.Value)
	}
}

func TestExecuteJSBlock_SyntaxError(t *testing.T) {
	registry := NewVariableRegistry()
	block := &CodeBlock{
		ID:       "block_0",
		Language: "js",
		Content:  "this is not valid javascript {{{}}}",
	}

	_, _, err := ExecuteJSBlock(block, registry)
	if err == nil {
		t.Error("Expected error for invalid JavaScript")
	}
}
