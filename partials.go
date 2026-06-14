package notes

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/localitas/localitas-go"
)

type SidebarData struct {
	Notes          []NoteListItem
	SelectedNoteID string
}

type EditorData struct {
	Note     *Note
	BasePath string
	Results  *ExecutionResults
}

type partialHandler struct {
	app *App
}

func (p *partialHandler) tmpl() (*template.Template, error) {
	tmpl := template.New("")
	partials := []string{
		"templates/partials/_sidebar_list.html",
		"templates/partials/_editor.html",
		"templates/partials/_empty.html",
		"templates/partials/_execution_results.html",
	}
	for _, file := range partials {
		content, err := TemplatesFS.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		if _, err := tmpl.Parse(string(content)); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
	}
	return tmpl, nil
}

func (p *partialHandler) tmplOOB() (*template.Template, error) {
	tmpl := template.New("")
	for _, file := range []string{
		"templates/partials/_editor.html",
		"templates/partials/_empty.html",
		"templates/partials/_execution_results.html",
	} {
		content, err := TemplatesFS.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		if _, err := tmpl.Parse(string(content)); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
	}
	sidebarContent, err := TemplatesFS.ReadFile("templates/partials/_sidebar_list.html")
	if err != nil {
		return nil, err
	}
	oobSidebar := strings.Replace(string(sidebarContent),
		`id="notes-sidebar-list"`,
		`id="notes-sidebar-list" hx-swap-oob="true"`, 1)
	if _, err := tmpl.Parse(oobSidebar); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (p *partialHandler) sidebarItems(r *http.Request) []NoteListItem {
	userID := client.UserIDFromRequest(r)
	notes, _ := p.app.Store.List(r.Context(), userID, 100, 0)
	items := make([]NoteListItem, 0, len(notes))
	for _, n := range notes {
		if n.ID == "" {
			continue
		}
		title := n.Title
		if title == "" {
			title = "Untitled"
		}
		items = append(items, NoteListItem{
			ID:        n.ID,
			Title:     title,
			Summary:   n.Summary,
			UpdatedAt: n.UpdatedAt,
		})
	}
	return items
}

func (p *partialHandler) handleSidebar(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmpl()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	selectedID := r.URL.Query().Get("selected")
	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Notes:          p.sidebarItems(r),
		SelectedNoteID: selectedID,
	})
}

func (p *partialHandler) handleEditor(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmpl()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	note, err := p.app.Store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Note not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "editor", EditorData{Note: note})
}

func (p *partialHandler) handleEmpty(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmpl()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "empty", nil)
}

func (p *partialHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmplOOB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userID := client.UserIDFromRequest(r)
	note, err := p.app.Store.Create(r.Context(), userID, "Untitled", "")
	if err != nil {
		http.Error(w, "Failed to create note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Notes:          p.sidebarItems(r),
		SelectedNoteID: note.ID,
	})
	tmpl.ExecuteTemplate(w, "editor", EditorData{Note: note})
}

func (p *partialHandler) handleSave(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmplOOB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	content := r.FormValue("content")

	if err := p.app.Store.Update(r.Context(), id, title, content); err != nil {
		http.Error(w, "Failed to save: "+err.Error(), http.StatusBadRequest)
		return
	}
	note, err := p.app.Store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to reload note after save", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Notes:          p.sidebarItems(r),
		SelectedNoteID: id,
	})
	tmpl.ExecuteTemplate(w, "editor", EditorData{Note: note})
}

func (p *partialHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmplOOB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Note ID required", http.StatusBadRequest)
		return
	}
	if err := p.app.Store.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "empty", nil)
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Notes: p.sidebarItems(r),
	})
}

func (p *partialHandler) handleExecute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	content := r.FormValue("content")

	if err := p.app.Store.Update(r.Context(), id, title, content); err != nil {
		http.Error(w, "Failed to save: "+err.Error(), http.StatusBadRequest)
		return
	}

	results, err := p.app.Store.ExecuteAllBlocks(r.Context(), id)
	if err != nil {
		http.Error(w, "Execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	note, err := p.app.Store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to reload note", http.StatusInternalServerError)
		return
	}

	tmpl, err := p.tmplOOB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Notes:          p.sidebarItems(r),
		SelectedNoteID: id,
	})
	tmpl.ExecuteTemplate(w, "editor", EditorData{Note: note, Results: results})
}

func (p *partialHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmpl()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	query := r.FormValue("query")
	if query == "" {
		p.handleSidebar(w, r)
		return
	}
	userID := client.UserIDFromRequest(r)
	results, err := p.app.Store.Search(r.Context(), userID, query, 100)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}
	items := make([]NoteListItem, 0, len(results))
	for _, n := range results {
		title := n.Title
		if title == "" {
			title = "Untitled"
		}
		items = append(items, NoteListItem{
			ID:        n.ID,
			Title:     title,
			Summary:   n.Summary,
			UpdatedAt: n.UpdatedAt,
		})
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{Notes: items})
}
