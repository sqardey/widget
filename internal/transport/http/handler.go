package http_transport

import (
	"encoding/json"
	"net/http"

	"github.com/sqardey/widget/internal/cache"
)

type Handler struct {
	Cache  *cache.ScheduleCache
	apiKey string
}

func NewHandler(cache *cache.ScheduleCache, key string) *Handler {
	return &Handler{Cache: cache, apiKey: key}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/widget", h.handleWidget)
	return mux
}

func (h *Handler) handleWidget(w http.ResponseWriter, r *http.Request) {
	querry := r.URL.Query().Get("key")
	if querry != h.apiKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := h.Cache.GetWidgetData()
	_ = json.NewEncoder(w).Encode(resp)
}
