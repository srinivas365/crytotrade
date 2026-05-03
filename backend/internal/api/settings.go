package api

import (
	"encoding/json"
	"net/http"

	"github.com/cryptotrade/app/internal/db"
)

type SettingsHandler struct{ queries *db.Queries }

func NewSettingsHandler(q *db.Queries) *SettingsHandler { return &SettingsHandler{q} }

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromCtx(r.Context())
	s, err := h.queries.GetSettings(r.Context(), uid)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func (h *SettingsHandler) Put(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromCtx(r.Context())
	var s db.UserSettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	s.UserID = uid
	if err := h.queries.UpsertSettings(r.Context(), &s); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
