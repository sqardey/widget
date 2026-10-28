package cache

import (
	"context"
	"sync"
	"time"

	"github.com/sqardey/widget/internal/domain"
	"github.com/sqardey/widget/internal/parser"
)

type ScheduleCache struct {
	mu        sync.RWMutex
	Classes   []domain.Class
	Exam      *domain.ExamInfo
	UpdatedAt time.Time
	Location  *time.Location
}

func New(loc *time.Location) *ScheduleCache {
	return &ScheduleCache{
		Location: loc,
	}
}

func (c *ScheduleCache) Update(classes []domain.Class, exam *domain.ExamInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Classes = classes
	c.Exam = exam
	c.UpdatedAt = time.Now().In(c.Location)
}

func (c *ScheduleCache) GetWidgetData() domain.WidgetResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()
	curTime := time.Now().In(c.Location).Format("15:04")
	response := domain.WidgetResponse{
		HasExam:  c.Exam != nil,
		Exam:     c.Exam,
		UpdateAt: c.UpdatedAt.Format("15:04"),
	}
	for _, cl := range c.Classes {
		if cl.StartTime <= curTime && cl.EndTime > curTime {
			response.CurrentClass = &cl
		} else if cl.StartTime > curTime {
			response.NextClass = &cl
			break
		}
	}
	return response
}

func (c *ScheduleCache) StartWorker(ctx context.Context,
	p *parser.Parser, interval time.Duration) {

}
