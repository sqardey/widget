package http_transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sqardey/widget/internal/cache"
	"github.com/sqardey/widget/internal/repository"
)

type Handler struct {
	Cache *cache.ScheduleCache
	Repo  *repository.Storage
}

func NewHandler(cache *cache.ScheduleCache, repo *repository.Storage) *Handler {
	return &Handler{Cache: cache, Repo: repo}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/widget", h.handleWidget)
	return mux
}

func (h *Handler) handleWidget(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("key")
	if query == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "api key not be has empty")
		return
	}

	ctx := r.Context()
	user, err := h.Repo.GetUserByKey(ctx, query)
	if err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	force := r.URL.Query().Get("force") == "1"

	response, updated, err := h.Cache.GetOrUpdate(ctx, user, force)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if updated == true {
		h.Repo.TouchUserUpdatedAt(ctx, user.Key, time.Now())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(response)
}
