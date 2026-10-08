package notes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleSwagger_ReturnsAPIDoc(t *testing.T) {
	w := httptest.NewRecorder()
	HandleSwagger(w, httptest.NewRequest("GET", "/swagger.json", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var spec APIDoc
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if spec.AppName != "Notes" {
		t.Errorf("expected app_name Notes, got %q", spec.AppName)
	}
	hasNotes := false
	for _, ep := range spec.Endpoints {
		if ep.Path == "/api/notes" {
			hasNotes = true
		}
	}
	if !hasNotes {
		t.Error("expected /api/notes endpoint")
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
