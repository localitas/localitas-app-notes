package notes

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type APIEndpoint struct {
	Method      string     `json:"method"`
	Path        string     `json:"path"`
	Summary     string     `json:"summary"`
	QueryParams []APIParam `json:"query_params,omitempty"`
	RequestBody *APIBody   `json:"request_body,omitempty"`
	Response    *APIBody   `json:"response,omitempty"`
}

type APIParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIBody struct {
	ContentType string `json:"content_type"`
	Example     string `json:"example"`
}

type APIFieldDoc struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

type APIDoc struct {
	AppName     string        `json:"app_name"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Keywords    []string      `json:"keywords,omitempty"`
	Fields      []APIFieldDoc `json:"fields,omitempty"`
	Endpoints   []APIEndpoint `json:"endpoints"`
}

var NotesAPIDoc = APIDoc{
	AppName:     "Notes",
	Version:     "0.1.0",
	Description: "Markdown notes with full-text search and code execution (JavaScript)",
	Keywords:    []string{"notes", "note", "memo", "markdown", "write", "document", "journal", "notebook", "snippet", "text"},
	Fields: []APIFieldDoc{
		{Name: "Headings & Text", Description: "Standard markdown formatting", Example: "# Heading 1\n## Heading 2\n\n**bold** *italic* ~~strikethrough~~\n[link](https://example.com)"},
		{Name: "Lists & Tasks", Description: "Bullet lists, numbered lists, task lists", Example: "- Bullet item\n  - Nested\n\n1. First\n2. Second\n\n- [ ] Todo\n- [x] Done"},
		{Name: "Code Blocks", Description: "Fenced code blocks with optional variable name", Example: "```js:myVar\nconst x = 42;\nx\n```"},
		{Name: "Tables & Quotes", Description: "Blockquotes and tables", Example: "> Blockquote\n\n| Col 1 | Col 2 |\n|-------|-------|\n| A     | B     |"},
	},
	Endpoints: []APIEndpoint{
		{
			Method:  "GET",
			Path:    "/api/notes",
			Summary: "List all notes",
			QueryParams: []APIParam{
				{Name: "limit", Type: "integer", Description: "Max results (default 50)"},
				{Name: "offset", Type: "integer", Description: "Pagination offset (default 0)"},
			},
			Response: &APIBody{ContentType: "application/json", Example: `[{"id":"abc...","title":"My Note","content":"# Hello\n...","summary":"Hello world..."}]`},
		},
		{
			Method:      "POST",
			Path:        "/api/notes",
			Summary:     "Create a note",
			RequestBody: &APIBody{ContentType: "application/json", Example: `{"title":"My Note","content":"# Hello\nWorld"}`},
			Response:    &APIBody{ContentType: "application/json", Example: `{"id":"abc...","title":"My Note","content":"# Hello\nWorld"}`},
		},
		{
			Method:   "GET",
			Path:     "/api/notes/{id}",
			Summary:  "Get a note by ID",
			Response: &APIBody{ContentType: "application/json", Example: `{"id":"abc...","title":"My Note","content":"# Hello\n..."}`},
		},
		{
			Method:      "PUT",
			Path:        "/api/notes/{id}",
			Summary:     "Update a note",
			RequestBody: &APIBody{ContentType: "application/json", Example: `{"title":"Updated","content":"# Updated\nNew content"}`},
			Response:    &APIBody{ContentType: "application/json", Example: `{"id":"abc...","title":"Updated","content":"# Updated\n..."}`},
		},
		{
			Method:   "DELETE",
			Path:     "/api/notes/{id}",
			Summary:  "Delete a note",
			Response: &APIBody{ContentType: "application/json", Example: `{"success":true}`},
		},
		{
			Method:  "GET",
			Path:    "/api/search",
			Summary: "Search notes (FTS5)",
			QueryParams: []APIParam{
				{Name: "q", Type: "string", Required: true, Description: "Search query"},
				{Name: "limit", Type: "integer", Description: "Max results (default 20)"},
			},
			Response: &APIBody{ContentType: "application/json", Example: `[{"id":"abc...","title":"My Note","content":"..."}]`},
		},
		{
			Method:   "POST",
			Path:     "/api/notes/{id}/execute",
			Summary:  "Execute all code blocks in a note",
			Response: &APIBody{ContentType: "application/json", Example: `{"id":"abc...","total_blocks":2,"success_count":2,"results":{"block_0":{"success":true,"output":"42"}}}`},
		},
	},
}

func HandleSwagger(w http.ResponseWriter, r *http.Request) {
	spec := generateSwaggerSpec(NotesAPIDoc)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}

func generateSwaggerSpec(doc APIDoc) map[string]interface{} {
	paths := make(map[string]interface{})
	for _, ep := range doc.Endpoints {
		methodKey := strings.ToLower(ep.Method)
		opID := methodKey + "_" + strings.NewReplacer("/", "_", "{", "", "}", "").Replace(ep.Path)
		operation := map[string]interface{}{
			"summary":     ep.Summary,
			"operationId": opID,
			"responses":   map[string]interface{}{"200": map[string]interface{}{"description": "Success"}},
		}
		if len(ep.QueryParams) > 0 {
			params := make([]map[string]interface{}, 0)
			for _, p := range ep.QueryParams {
				params = append(params, map[string]interface{}{"name": p.Name, "in": "query", "required": p.Required, "description": p.Description, "schema": map[string]string{"type": p.Type}})
			}
			operation["parameters"] = params
		}
		if ep.RequestBody != nil {
			operation["requestBody"] = map[string]interface{}{"content": map[string]interface{}{ep.RequestBody.ContentType: map[string]interface{}{"example": json.RawMessage(ep.RequestBody.Example)}}}
		}
		if ep.Response != nil {
			operation["responses"].(map[string]interface{})["200"] = map[string]interface{}{"description": "Success", "content": map[string]interface{}{ep.Response.ContentType: map[string]interface{}{"example": json.RawMessage(ep.Response.Example)}}}
		}
		if _, exists := paths[ep.Path]; !exists {
			paths[ep.Path] = make(map[string]interface{})
		}
		paths[ep.Path].(map[string]interface{})[methodKey] = operation
	}
	return map[string]interface{}{"openapi": "3.0.3", "info": map[string]interface{}{"title": doc.AppName, "version": doc.Version, "description": doc.Description}, "paths": paths}
}

func RenderDocsHTML(doc APIDoc) template.HTML {
	var sb strings.Builder
	if len(doc.Fields) > 0 {
		sb.WriteString(`<h3 style="font-size: 0.875rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; color: var(--color-text-secondary); margin-bottom: 1rem;">Markdown Guide</h3><div class="accordion-list">`)
		for _, f := range doc.Fields {
			sb.WriteString(fmt.Sprintf(`<details class="glass-panel" style="border-radius: 0.5rem; margin-bottom: 0.5rem;"><summary style="padding: 0.75rem 1rem; cursor: pointer; font-weight: 500; color: var(--color-text-primary);">%s</summary><div style="padding: 0 1rem 0.75rem; font-size: 0.875rem; color: var(--color-text-secondary);"><p style="margin-bottom: 0.5rem;">%s</p><pre style="background: var(--color-bg-base); padding: 0.75rem; border-radius: 0.375rem; overflow-x: auto; font-size: 0.8125rem;">%s</pre></div></details>`, template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Description), template.HTMLEscapeString(f.Example)))
		}
		sb.WriteString(`</div><hr style="border-color: var(--color-glass-border); margin: 1.5rem 0;">`)
	}
	sb.WriteString(`<h3 style="font-size: 0.875rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; color: var(--color-text-secondary); margin-bottom: 1rem;">API Endpoints</h3><div class="accordion-list">`)
	for _, ep := range doc.Endpoints {
		title := fmt.Sprintf("%s %s — %s", ep.Method, ep.Path, ep.Summary)
		sb.WriteString(fmt.Sprintf(`<details class="glass-panel" style="border-radius: 0.5rem; margin-bottom: 0.5rem;"><summary style="padding: 0.75rem 1rem; cursor: pointer; font-weight: 500; color: var(--color-text-primary);">%s</summary><div style="padding: 0 1rem 0.75rem; font-size: 0.875rem; color: var(--color-text-secondary);">`, template.HTMLEscapeString(title)))
		var ex strings.Builder
		if ep.RequestBody != nil {
			ex.WriteString("# Request\n")
			ex.WriteString(prettyJSON(ep.RequestBody.Example))
			ex.WriteString("\n\n")
		}
		if ep.Response != nil {
			ex.WriteString("# Response\n")
			ex.WriteString(prettyJSON(ep.Response.Example))
		}
		if ex.Len() > 0 {
			sb.WriteString(fmt.Sprintf(`<pre style="background: var(--color-bg-base); padding: 0.75rem; border-radius: 0.375rem; overflow-x: auto; font-size: 0.8125rem;">%s</pre>`, template.HTMLEscapeString(ex.String())))
		}
		sb.WriteString(`</div></details>`)
	}
	sb.WriteString(`</div>`)
	return template.HTML(sb.String())
}

func prettyJSON(s string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return s
	}
	return string(b)
}
