package parser

import (
	"context"
	"net/http"

	"github.com/sqardey/widget/internal/domain"
)

type Parser struct {
	parser *http.Client
}

func New() *Parser {
	return nil
}
func FetchTodaySchedule(ctx context.Context) ([]domain.Class,
	*domain.ExamInfo, error) {
	return []domain.Class{{NameSubject: "Высшмат", Audience: 12, Duration: 80}, {NameSubject: "Высшмат", Audience: 12, Duration: 80}},
		&domain.ExamInfo{NameSubject: "Дискретная математика", Audience: 15, Lesson: 5}, nil
}
