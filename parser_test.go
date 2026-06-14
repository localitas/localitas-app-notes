package notes

import (
	"testing"
)

func TestParseCodeBlocks(t *testing.T) {
	markdown := "# Test Note\n\nSome text here.\n\n```csv:sales_data\nfile: data/sales.csv\n```\n\nMore text.\n\n```js:config\nconst x = 42;\nreturn x;\n```\n\n```python\nprint(\"hello\")\n```\n"

	blocks := ParseCodeBlocks(markdown)

	if len(blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(blocks))
	}

	if blocks[0].Language != "csv" {
		t.Errorf("Expected language 'csv', got '%s'", blocks[0].Language)
	}
	if blocks[0].VarName != "sales_data" {
		t.Errorf("Expected varname 'sales_data', got '%s'", blocks[0].VarName)
	}
	if blocks[0].ID != "block_0" {
		t.Errorf("Expected ID 'block_0', got '%s'", blocks[0].ID)
	}

	if blocks[1].Language != "js" {
		t.Errorf("Expected language 'js', got '%s'", blocks[1].Language)
	}
	if blocks[1].VarName != "config" {
		t.Errorf("Expected varname 'config', got '%s'", blocks[1].VarName)
	}

	if blocks[2].Language != "python" {
		t.Errorf("Expected language 'python', got '%s'", blocks[2].Language)
	}
	if blocks[2].VarName != "" {
		t.Errorf("Expected empty varname, got '%s'", blocks[2].VarName)
	}
}

func TestParseCSVBlock(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectPath  string
		expectError bool
	}{
		{"valid csv block", "file: data/sales.csv\nschema: auto", "data/sales.csv", false},
		{"file only", "file: data/test.csv", "data/test.csv", false},
		{"missing file directive", "schema: auto", "", true},
		{"empty content", "", "", true},
		{"file with spaces", "file:    data/my file.csv   ", "data/my file.csv", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ParseCSVBlock(tt.content)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if path != tt.expectPath {
					t.Errorf("Expected path '%s', got '%s'", tt.expectPath, path)
				}
			}
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"simple identifier", "myVar", true},
		{"with underscore", "my_var", true},
		{"starts with underscore", "_private", true},
		{"with numbers", "var123", true},
		{"starts with number", "123var", false},
		{"with hyphen", "my-var", false},
		{"empty string", "", false},
		{"single char", "x", true},
		{"single underscore", "_", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidIdentifier(tt.input)
			if result != tt.valid {
				t.Errorf("isValidIdentifier('%s') = %v, expected %v", tt.input, result, tt.valid)
			}
		})
	}
}

func TestGetBlockByID(t *testing.T) {
	blocks := []*CodeBlock{
		{ID: "block_0", Language: "csv", VarName: "data"},
		{ID: "block_1", Language: "js", VarName: "config"},
		{ID: "block_2", Language: "python"},
	}

	block := GetBlockByID(blocks, "block_1")
	if block == nil {
		t.Fatalf("Expected to find block_1")
	}
	if block.Language != "js" {
		t.Errorf("Expected language 'js', got '%s'", block.Language)
	}

	block = GetBlockByID(blocks, "block_99")
	if block != nil {
		t.Errorf("Expected nil for non-existent block")
	}
}

func TestGetBlocksByLanguage(t *testing.T) {
	blocks := []*CodeBlock{
		{ID: "block_0", Language: "csv"},
		{ID: "block_1", Language: "js"},
		{ID: "block_2", Language: "csv"},
		{ID: "block_3", Language: "python"},
	}

	csvBlocks := GetBlocksByLanguage(blocks, "csv")
	if len(csvBlocks) != 2 {
		t.Errorf("Expected 2 CSV blocks, got %d", len(csvBlocks))
	}

	pythonBlocks := GetBlocksByLanguage(blocks, "python")
	if len(pythonBlocks) != 1 {
		t.Errorf("Expected 1 Python block, got %d", len(pythonBlocks))
	}

	rustBlocks := GetBlocksByLanguage(blocks, "rust")
	if len(rustBlocks) != 0 {
		t.Errorf("Expected 0 Rust blocks, got %d", len(rustBlocks))
	}
}
