package notes

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/localitas/localitas-go"
)

var logger = slog.Default().With("component", "notes")

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
			logger.Warn("install attempt failed", "attempt", attempt, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := applyEmbeddedMigrations(ctx, a.client, db.ID); err != nil {
			logger.Warn("install migrations attempt failed", "attempt", attempt, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		return db.ID, nil
	}
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(TemplatesFS, "templates/index.html")
	if err != nil {
		logger.Error("index template error", "error", err)
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
		logger.Error("index render error", "error", err)
	}
}

func (a *App) RegisterRoutes(mux *http.ServeMux) {
	h := &handler{app: a}
	p := &partialHandler{app: a}

	mux.HandleFunc("GET /{$}", a.handleIndex)
	mux.HandleFunc("GET /swagger.json", HandleSwagger)
	mux.HandleFunc("GET /help.md", handleHelpMarkdown)
	// Every note route acts for the user core vouches for; one without a
	// verified user is refused, and changing a note needs write scope.
	user := func(pattern string, handler http.HandlerFunc) {
		mux.Handle(pattern, client.AuthMiddleware(handler))
	}
	writer := func(pattern string, handler http.HandlerFunc) {
		user(pattern, client.RequireScopeFunc(client.ScopeWrite, handler))
	}
	user("GET /api/notes", h.handleList)
	writer("POST /api/notes", h.handleCreate)
	user("GET /api/notes/{id}", h.handleGet)
	writer("PUT /api/notes/{id}", h.handleUpdate)
	writer("DELETE /api/notes/{id}", h.handleDelete)
	user("GET /api/search", h.handleSearch)
	writer("POST /api/notes/{id}/execute", h.handleExecute)
	writer("POST /api/notes/{id}/execute/{blockID}", h.handleExecuteBlock)

	user("GET /partials/sidebar", p.handleSidebar)
	user("GET /partials/editor/{id}", p.handleEditor)
	user("GET /partials/empty", p.handleEmpty)
	writer("POST /partials/create", p.handleCreate)
	writer("POST /partials/save/{id}", p.handleSave)
	writer("DELETE /partials/delete/{id}", p.handleDelete)
	user("POST /partials/search", p.handleSearch)
	writer("POST /partials/execute/{id}", p.handleExecute)
}
