package notes

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/localitas/localitas-go"
)

type handler struct {
	app *App
}

func (h *handler) handleList(w http.ResponseWriter, r *http.Request) {
	userID := client.UserIDFromRequest(r)
	limit := intParam(r, "limit", 50)
	offset := intParam(r, "offset", 0)

	notes, err := h.app.Store.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to list notes: %v", err)
		return
	}

	writeJSON(w, http.StatusOK, notes)
}

func (h *handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n, err := h.app.Store.Get(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "note not found: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID := client.UserIDFromRequest(r)
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body: %v", err)
		return
	}
	n, err := h.app.Store.Create(r.Context(), userID, req.Title, req.Content)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to create note: %v", err)
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body: %v", err)
		return
	}
	if err := h.app.Store.Update(r.Context(), id, req.Title, req.Content); err != nil {
		writeErr(w, http.StatusInternalServerError, "%v", err)
		return
	}
	n, err := h.app.Store.Get(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "saved but failed to reload: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.app.Store.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to delete note: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	userID := client.UserIDFromRequest(r)
	query := r.URL.Query().Get("q")
	if query == "" {
		writeErr(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	limit := intParam(r, "limit", 20)

	notes, err := h.app.Store.Search(r.Context(), userID, query, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "search failed: %v", err)
		return
	}

	writeJSON(w, http.StatusOK, notes)
}

func (h *handler) handleExecute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	results, err := h.app.Store.ExecuteAllBlocks(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "execution failed: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *handler) handleExecuteBlock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	blockID := r.PathValue("blockID")
	result, err := h.app.Store.ExecuteSingleBlock(r.Context(), id, blockID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "execution failed: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf(format, args...)})
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
