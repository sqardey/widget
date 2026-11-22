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
	Parser    *parser.Parser
	Login     string
	Password  string
}

func New(loc *time.Location, logger *slog.Logger) *ScheduleCache {
	return &ScheduleCache{
		users:    make(map[string]*UserCacheEntry),
		Location: loc,
		logger:   logger,
	}
}

// getOrCreateEntry безопасно достаёт или создаёт запись пользователя в мапе
func (c *ScheduleCache) getOrCreateEntry(u *domain.User) *UserCacheEntry {
	c.mu.RLock()
	entry, exists := c.users[u.Key]
	c.mu.RUnlock()

	if exists {
		return entry
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Повторная проверка на случай гонки между RUnlock и Lock
	if entry, exists = c.users[u.Key]; exists {
		return entry
	}

	entry = &UserCacheEntry{
		Parser:    parser.New(u.Login, u.Password),
		Login:     u.Login,
		Password:  u.Password,
		UpdatedAt: u.UpdatedAt,
	}
	c.users[u.Key] = entry
	return entry
}

// GetOrUpdate возвращает расписание пользователя (из памяти или обновляя из журнала).
// Возвращает: (ответ_виджету, было_ли_реальное_обновление, ошибка)
func (c *ScheduleCache) GetOrUpdate(ctx context.Context, u *domain.User, force bool) (domain.WidgetResponse, bool, error) {
	entry := c.getOrCreateEntry(u)

	// Блокируем только конкретного пользователя
	entry.mu.Lock()
	defer entry.mu.Unlock()

	// Если пользователь сменил логин/пароль в БД — пересоздаём его парсер
	if entry.Login != u.Login || entry.Password != u.Password {
		entry.Parser = parser.New(u.Login, u.Password)
		entry.Login = u.Login
		entry.Password = u.Password
		entry.UpdatedAt = time.Time{} // сбрасываем таймер для принудительного обновления
	}

	now := time.Now().In(c.Location)

	// Проверяем условия, нужно ли идти в журнал Академии:
	isFirstLoad := entry.UpdatedAt.IsZero() || entry.Classes == nil
	isNewDay := !isFirstLoad && (entry.UpdatedAt.YearDay() != now.YearDay() || entry.UpdatedAt.Year() != now.Year())
	isExpired := !isFirstLoad && now.Sub(entry.UpdatedAt) >= AutoRefreshTTL
	isForceAllowed := force && !isFirstLoad && now.Sub(entry.UpdatedAt) >= ManualRefreshCooldown

	needUpdate := isFirstLoad || isNewDay || isExpired || isForceAllowed

	updated := false
	if needUpdate {
		classes, exam, err := entry.Parser.FetchTodaySchedule(ctx, now)
		if err != nil {
			c.logger.Warn("ERROR fecth failed", "err", err)
			// Если в кэше еще вообще ничего нет — возвращаем ошибку
			if isFirstLoad {
				return domain.WidgetResponse{}, false, err
			}
			// Если старые данные за сегодня есть — не ломаем виджет, отдаём их
		} else {
			entry.Classes = classes
			entry.Exam = exam
			entry.UpdatedAt = now
			updated = true
			c.logger.Info("Schedule updated", "key", u.Key, "classes", len(classes), "exam", exam != nil)
		}
	}

	return entry.buildResponse(now), updated, nil
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
