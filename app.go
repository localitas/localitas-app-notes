package notes

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/localitas/localitas-go"
)

type App struct {
	Store    *Store
	BasePath string
	client   *client.Client
}

func New(c *client.Client, basePath string) *App {
	if basePath == "" {
		basePath = "/"
	}
	return &App{
		BasePath: basePath,
		client:   c,
	}
}

func (a *App) InitStore(coreURL, dbID, token string) error {
	store, err := OpenStore(coreURL, dbID, token)
	if err != nil {
		return err
	}
	a.Store = store
	return nil
}

func (a *App) Install(ctx context.Context) (string, error) {
	for attempt := 1; ; attempt++ {
		db, err := a.client.CreateSystemDatabase(ctx, DatabaseName)
		if err != nil {
			log.Printf("install: attempt %d failed (retrying): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := applyEmbeddedMigrations(ctx, a.client, db.ID); err != nil {
			log.Printf("install: migrations attempt %d failed (retrying): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		return db.ID, nil
	}
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(TemplatesFS, "templates/index.html")
	if err != nil {
		log.Printf("notes index template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	data := map[string]string{
		"BasePath":   a.BasePath,
		"SelectedID": r.URL.Query().Get("id"),
	}
	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("notes index render error: %v", err)
	}
}

func (a *App) RegisterRoutes(mux *http.ServeMux) {
	h := &handler{app: a}
	p := &partialHandler{app: a}

	mux.HandleFunc("GET /{$}", a.handleIndex)
	mux.HandleFunc("GET /swagger.json", HandleSwagger)
	mux.HandleFunc("GET /help.md", handleHelpMarkdown)
	mux.HandleFunc("GET /api/notes", h.handleList)
	mux.HandleFunc("POST /api/notes", h.handleCreate)
	mux.HandleFunc("GET /api/notes/{id}", h.handleGet)
	mux.HandleFunc("PUT /api/notes/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/notes/{id}", h.handleDelete)
	mux.HandleFunc("GET /api/search", h.handleSearch)
	mux.HandleFunc("POST /api/notes/{id}/execute", h.handleExecute)
	mux.HandleFunc("POST /api/notes/{id}/execute/{blockID}", h.handleExecuteBlock)

	mux.HandleFunc("GET /partials/sidebar", p.handleSidebar)
	mux.HandleFunc("GET /partials/editor/{id}", p.handleEditor)
	mux.HandleFunc("GET /partials/empty", p.handleEmpty)
	mux.HandleFunc("POST /partials/create", p.handleCreate)
	mux.HandleFunc("POST /partials/save/{id}", p.handleSave)
	mux.HandleFunc("DELETE /partials/delete/{id}", p.handleDelete)
	mux.HandleFunc("POST /partials/search", p.handleSearch)
	mux.HandleFunc("POST /partials/execute/{id}", p.handleExecute)
}
