package http_transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sqardey/widget/internal/cache"
	"github.com/sqardey/widget/internal/domain"
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
	mux.HandleFunc("POST /api/widget/newuser", h.newUser)
	return mux
}

func (h *Handler) newUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user domain.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "ERROR: %v", err)
		return
	}

	if user.Login == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Login and password must be entered")
		return
	}

	classes, exam, p, err := h.Cache.ValidateCredentials(ctx, user.Login, user.Password)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Неверный логин или пароль (либо журнал недоступен)")
		return
	}

	key, err := h.Repo.NewUser(ctx, user.Login, user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERROR: %v", err)
		return
	}
	user.Key = key
	h.Cache.SeedUser(&user, p, classes, exam)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{"key": key})
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
		fmt.Fprintf(w, "ERROR: %v", err)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "ERROR: %v", err)
		return
	}
	force := r.URL.Query().Get("force") == "1"

	response, updated, err := h.Cache.GetOrUpdate(ctx, user, force)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "ERROR: %v", err)
		return
	}
	if updated == true {
		h.Repo.TouchUserUpdatedAt(ctx, user.Key, time.Now())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(response)
}
