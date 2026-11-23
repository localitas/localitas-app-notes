package notes

import (
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
)

type JSExecutor struct {
	vm      *goja.Runtime
	console []string
}

func NewJSExecutor() *JSExecutor {
	exec := &JSExecutor{vm: goja.New()}
	console := exec.vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			parts[i] = arg.String()
		}
		exec.console = append(exec.console, strings.Join(parts, " "))
		return goja.Undefined()
	})
	exec.vm.Set("console", console)
	return exec
}

func ExecuteJSBlock(block *CodeBlock, registry *VariableRegistry) (*Variable, string, error) {
	executor := NewJSExecutor()

	for varName, variable := range registry.Variables {
		value := variable.Value
		if variable.Type == "dataframe" {
			df, err := executor.dataframeView(varName, variable.Schema)
			if err != nil {
				return nil, "", err
			}
			value = df
		}
		if err := executor.vm.Set(varName, value); err != nil {
			return nil, "", fmt.Errorf("set %s: %w", varName, err)
		}
	}

	result, err := executor.vm.RunString(block.Content)
	if err != nil {
		return nil, "", fmt.Errorf("JavaScript execution error: %w", err)
	}

	var resultValue interface{}
	if result != nil {
		resultValue = result.Export()
	}

	resultType := detectJSType(resultValue)

	var variable *Variable
	if block.VarName != "" {
		variable = &Variable{
			Name:        block.VarName,
			Type:        resultType,
			Source:      fmt.Sprintf("js:%s", block.ID),
			Value:       resultValue,
			DefinedIn:   block.ID,
			LastUpdated: time.Now().UTC(),
		}
	}

	output := formatJSOutput(resultValue)
	if len(executor.console) > 0 {
		consoleOut := strings.Join(executor.console, "\n")
		if output != "" && output != "undefined" && output != "null" {
			output = consoleOut + "\n" + output
		} else {
			output = consoleOut
		}
	}
	return variable, output, nil
}

// dataframeView is what JS sees of a dataframe: its shape, not its rows.
func (e *JSExecutor) dataframeView(name string, schema *DataSchema) (*goja.Object, error) {
	df := e.vm.NewObject()
	for _, p := range []struct {
		key   string
		value any
	}{
		{"_type", "dataframe"},
		{"_name", name},
		{"_rows", schema.Rows},
		{"_columns", len(schema.Columns)},
	} {
		if err := df.Set(p.key, p.value); err != nil {
			return nil, fmt.Errorf("dataframe %s: set %s: %w", name, p.key, err)
		}
	}
	return df, nil
}

// detectJSType and formatJSOutput read goja's Export of a JS object,
// which is a map[string]interface{}.
func detectJSType(value interface{}) string {
	if value == nil {
		return "null"
	}
	switch value.(type) {
	case bool:
		return "bool"
	case int, int8, int16, int32, int64:
		return "int"
	case float32, float64:
		return "float"
	case string:
		return "string"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	default:
		return "unknown"
	}
}

func formatJSOutput(value interface{}) string {
	if value == nil {
		return "null"
	}
	switch v := value.(type) {
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case string:
		return v
	case map[string]interface{}:
		return fmt.Sprintf("Object with %d keys", len(v))
	case []interface{}:
		return fmt.Sprintf("Array with %d elements", len(v))
	default:
		return fmt.Sprintf("%v", v)
	}
}
