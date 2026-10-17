package notes

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/localitas/localitas-go"
	"github.com/localitas/localitas-go/httputil"
)

type handler struct {
	app *App
}

func (h *handler) handleList(w http.ResponseWriter, r *http.Request) {
	userID := client.GetUserID(r.Context())
	limit := intParam(r, "limit", 50)
	offset := intParam(r, "offset", 0)

	notes, err := h.app.Store.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "failed to list notes: %v", err)
		return
	}

	writeJSON(w, r, http.StatusOK, notes)
}

func (h *handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n, err := h.app.Store.Get(r.Context(), client.GetUserID(r.Context()), id)
	if err != nil {
		writeStoreErr(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, n)
}

func (h *handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID := client.GetUserID(r.Context())
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid request body: %v", err)
		return
	}
	n, err := h.app.Store.Create(r.Context(), userID, req.Title, req.Content)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "failed to create note: %v", err)
		return
	}
	writeJSON(w, r, http.StatusCreated, n)
}

func (h *handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, r, http.StatusBadRequest, "invalid request body: %v", err)
		return
	}
	if err := h.app.Store.Update(r.Context(), client.GetUserID(r.Context()), id, req.Title, req.Content); err != nil {
		writeStoreErr(w, r, err)
		return
	}
	n, err := h.app.Store.Get(r.Context(), client.GetUserID(r.Context()), id)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "saved but failed to reload: %v", err)
		return
	}
	writeJSON(w, r, http.StatusOK, n)
}

func (h *handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.app.Store.Delete(r.Context(), client.GetUserID(r.Context()), id); err != nil {
		writeStoreErr(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	userID := client.GetUserID(r.Context())
	query := r.URL.Query().Get("q")
	if query == "" {
		writeErr(w, r, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	limit := intParam(r, "limit", 20)

	notes, err := h.app.Store.Search(r.Context(), userID, query, limit)
	if err != nil {
		writeErr(w, r, http.StatusInternalServerError, "search failed: %v", err)
		return
	}

	writeJSON(w, r, http.StatusOK, notes)
}

func (h *handler) handleExecute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	results, err := h.app.Store.ExecuteAllBlocks(r.Context(), client.GetUserID(r.Context()), id)
	if err != nil {
		writeStoreErr(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, results)
}

func (h *handler) handleExecuteBlock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	blockID := r.PathValue("blockID")
	result, err := h.app.Store.ExecuteSingleBlock(r.Context(), client.GetUserID(r.Context()), id, blockID)
	if err != nil {
		writeStoreErr(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, v interface{}) {
	httputil.WriteResponse(w, r, status, v)
}

func writeErr(w http.ResponseWriter, r *http.Request, status int, format string, args ...interface{}) {
	httputil.WriteError(w, r, status, format, args...)
}

// writeStoreErr answers 404 for a note that is not the caller's, 500 for
// anything else.
func writeStoreErr(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrNotFound) {
		writeErr(w, r, http.StatusNotFound, "note not found")
		return
	}
	writeErr(w, r, http.StatusInternalServerError, "%v", err)
}

func intParam(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func escapeHTML(s string) string {
	return template.HTMLEscapeString(s)
}
