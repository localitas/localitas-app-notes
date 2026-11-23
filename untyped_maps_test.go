package notes

import (
	"testing"

	"github.com/localitas/localitas-go/maptest"
)

func TestNoUntypedMaps(t *testing.T) {
	maptest.NoUntypedMaps(t, ".", map[string]maptest.Use{
		"js_executor.go": {Count: 2, Why: "goja's Value.Export returns a map[string]interface{} for a JS object; detectJSType and formatJSOutput switch on it"},
	})
}
