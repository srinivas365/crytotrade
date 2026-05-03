package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cryptotrade/app/internal/db"
)

type HistoryHandler struct{ queries *db.Queries }

func NewHistoryHandler(q *db.Queries) *HistoryHandler { return &HistoryHandler{q} }

type historyResponse struct {
	Records []*db.AlertRecord `json:"records"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	Limit   int               `json:"limit"`
}

func (h *HistoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid := userIDFromCtx(r.Context())

	limit := 20
	page := 1
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	offset := (page - 1) * limit

	total, err := h.queries.CountAlertHistory(r.Context(), uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	records, err := h.queries.GetAlertHistory(r.Context(), uid, limit, offset)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if records == nil {
		records = []*db.AlertRecord{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(historyResponse{Records: records, Total: total, Page: page, Limit: limit})
}
