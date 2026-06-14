package notes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleSwagger_ReturnsValidOpenAPI(t *testing.T) {
	req := httptest.NewRequest("GET", "/swagger.json", nil)
	w := httptest.NewRecorder()
	HandleSwagger(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var spec map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if spec["openapi"] != "3.0.3" {
		t.Errorf("expected openapi 3.0.3, got %v", spec["openapi"])
	}
	paths, _ := spec["paths"].(map[string]interface{})
	if paths["/api/notes"] == nil {
		t.Error("expected /api/notes path")
	}
}

func TestRenderDocsHTML_ContainsContent(t *testing.T) {
	html := string(RenderDocsHTML(NotesAPIDoc))
	if !strings.Contains(html, "API Endpoints") {
		t.Error("expected API Endpoints heading")
	}
	if !strings.Contains(html, "GET /api/notes") {
		t.Error("expected GET /api/notes")
	}
}
