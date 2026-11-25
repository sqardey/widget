package http_transport

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sqardey/widget/internal/cache"
	"github.com/sqardey/widget/internal/domain"
	"github.com/sqardey/widget/internal/repository"
)

//go:embed widget.js
var widgetTemplate string

//go:embed widget_v2.js
var widgetTemplateV2 string

type Handler struct {
	Cache             *cache.ScheduleCache
	Repo              *repository.Storage
	Logger            *slog.Logger
	credentialsCipher cipher.AEAD
	BotSecret         string
}

func NewHandler(cache *cache.ScheduleCache, repo *repository.Storage,
	logger *slog.Logger, botSecret string) (*Handler, error) {
	h := &Handler{
		Cache:     cache,
		Repo:      repo,
		Logger:    logger,
		BotSecret: botSecret,
	}
	if err := h.InitEncryption(); err != nil {
		return nil, fmt.Errorf("init encryption: %w", err)
	}

	return h, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/widget", h.handleWidget)
	mux.HandleFunc("POST /api/widget/newuser", h.newUserHandler)
	mux.HandleFunc("GET /api/widget/script", h.handleWidgetScript)
	mux.HandleFunc("GET /api/widget/script/v2", h.handleWidgetScriptv2)
	return h.requestIDMiddleware(h.loggingMiddleware(mux))
}

func (h *Handler) handleWidget(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("key")
	if query == "" {
		http.Error(w, "API key is required", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	user, err := h.Repo.GetUserByKey(ctx, query)
	if errors.Is(err, pgx.ErrNoRows) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		h.Logger.Error(
			"Failed to get user",
			"error", err,
			"request_id", getRequestID(ctx),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	login, err := h.decrypt(user.Login)
	if err != nil {
		h.Logger.Error(
			"Failed to decrypt login",
			"error", err,
			"request_id", getRequestID(ctx),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	password, err := h.decrypt(user.Password)
	if err != nil {
		h.Logger.Error(
			"Failed to decrypt password",
			"error", err,
			"request_id", getRequestID(ctx),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user.Login = login
	user.Password = password

	force := r.URL.Query().Get("force") == "1"

	response, updated, err := h.Cache.GetOrUpdate(ctx, user, force)
	if err != nil {
		h.Logger.Error(
			"Failed to get schedule",
			"error", err,
			"request_id", getRequestID(ctx),
		)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if updated {
		h.Repo.TouchUserUpdatedAt(ctx, user.Key, time.Now())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) newUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Bot-Secret") != h.BotSecret || h.BotSecret == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.Logger.Warn("Failed decode JSON", "err", err)
		return
	}

	if user.Login == "" || user.Password == "" || user.TelegramID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	encryptedLogin := h.encrypt(user.Login)
	encryptedPassword := h.encrypt(user.Password)

	key, err := h.Repo.PutNewUser(
		r.Context(),
		user.TelegramID,
		encryptedLogin,
		encryptedPassword,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.Logger.Error("Failed write new user to database", "error", err)
		return
	}
	h.Cache.Invalidate(key)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"key": key})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

type ctxKey string

const requestIDKey ctxKey = "request_id"

func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func (h *Handler) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		level := slog.LevelInfo
		switch {
		case rw.status >= 500:
			level = slog.LevelError
		case rw.status >= 400:
			level = slog.LevelWarn
		}
		h.Logger.Log(r.Context(), level, "http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.status),
			slog.String("duration", time.Since(startTime).String()),
			slog.String("request_id", getRequestID(r.Context())),
		)
	})
}

func (h *Handler) handleWidgetScript(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	script := strings.ReplaceAll(widgetTemplate, "{{API_KEY}}", key)
	script = strings.ReplaceAll(
		script,
		"{{BASE_URL}}",
		"https://widget.sqardey.xyz",
	)

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="IT_TOP_Widget.js"`)
	_, _ = w.Write([]byte(script))
}

func (h *Handler) handleWidgetScriptv2(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	script := strings.ReplaceAll(widgetTemplateV2, "{{API_KEY}}", key)
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="IT_TOP_Widget.js"`)
	_, _ = w.Write([]byte(script))
}
