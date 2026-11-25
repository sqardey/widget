package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/sqardey/widget/internal/domain"
	"github.com/sqardey/widget/internal/parser"
)

const (
	AutoRefreshTTL        = 3 * time.Hour
	ManualRefreshCooldown = 1 * time.Hour
)

// ScheduleCache — общий кэш всех пользователей
type ScheduleCache struct {
	mu       sync.RWMutex
	users    map[string]*UserCacheEntry
	Location *time.Location
	logger   *slog.Logger
}

// UserCacheEntry — состояние и расписание одного конкретного пользователя
type UserCacheEntry struct {
	mu        sync.Mutex
	Classes   []domain.Class
	Exam      *domain.ExamInfo
	UpdatedAt time.Time
}

func New(loc *time.Location, logger *slog.Logger) *ScheduleCache {
	return &ScheduleCache{
		users:    make(map[string]*UserCacheEntry),
		Location: loc,
		logger:   logger,
	}
}

func (c *ScheduleCache) getOrCreateEntry(u *domain.User) *UserCacheEntry {
	c.mu.RLock()
	entry, exists := c.users[u.Key]
	c.mu.RUnlock()

	if exists {
		return entry
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists = c.users[u.Key]; exists {
		return entry
	}

	entry = &UserCacheEntry{}
	c.users[u.Key] = entry
	return entry
}

func (c *ScheduleCache) GetOrUpdate(
	ctx context.Context,
	u *domain.User,
	force bool,
) (domain.WidgetResponse, bool, error) {
	entry := c.getOrCreateEntry(u)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now().In(c.Location)

	isFirstLoad := entry.UpdatedAt.IsZero() || entry.Classes == nil
	isNewDay := !isFirstLoad &&
		entry.UpdatedAt.Format("2006-01-02") != now.Format("2006-01-02")
	isExpired := !isFirstLoad &&
		now.Sub(entry.UpdatedAt) >= AutoRefreshTTL
	isForceAllowed := force && !isFirstLoad &&
		now.Sub(entry.UpdatedAt) >= ManualRefreshCooldown

	needUpdate := isFirstLoad || isNewDay || isExpired || isForceAllowed

	if !needUpdate {
		return entry.buildResponse(now), false, nil
	}

	// u.Login и u.Password уже расшифрованы в handleWidget.
	// Парсер существует только во время этого вызова.
	p := parser.New(u.Login, u.Password)

	classes, exam, err := p.FetchTodaySchedule(ctx, now)
	if err != nil {
		c.logger.Warn("Failed to fetch schedule", "error", err)

		// Вчерашнее расписание нельзя отдавать как сегодняшнее.
		if isFirstLoad || isNewDay {
			return domain.WidgetResponse{}, false, err
		}

		return entry.buildResponse(now), false, nil
	}

	entry.Classes = classes
	entry.Exam = exam
	entry.UpdatedAt = now

	c.logger.Info(
		"Schedule updated",
		"classes", len(classes),
		"exam", exam != nil,
	)

	return entry.buildResponse(now), true, nil
}

// buildResponse вычисляет текущую и следующую пару на момент времени now
func (e *UserCacheEntry) buildResponse(now time.Time) domain.WidgetResponse {
	curTime := now.Format("15:04")
	resp := domain.WidgetResponse{
		HasExam:  e.Exam != nil,
		Exam:     e.Exam,
		UpdateAt: e.UpdatedAt.Format("15:04"),
	}

	for i := range e.Classes {
		cl := e.Classes[i]
		if cl.StartTime <= curTime && cl.EndTime > curTime {
			resp.CurrentClass = &cl
		} else if cl.StartTime > curTime {
			resp.NextClass = &cl
			break
		}
	}

	return resp
}

func (c *ScheduleCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.users, key)
}
