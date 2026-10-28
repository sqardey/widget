package http_transport

import "github.com/sqardey/widget/internal/cache"

type Handler struct {
	Cache  *cache.ScheduleCache
	apiKey string
}

func NewHandler() *Handler {
	return nil
}
