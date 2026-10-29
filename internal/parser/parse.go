package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sort"
	"time"

	"github.com/sqardey/widget/internal/domain"
)

const baseURL = "https://journalka.top-academy.ru"

type Parser struct {
	client   *http.Client
	login    string
	password string
	loggedIn bool
}

func New(login, password string) *Parser {
	jar, _ := cookiejar.New(nil)

	return &Parser{
		client: &http.Client{
			Jar:     jar,
			Timeout: 15 * time.Second,
		},
		login:    login,
		password: password,
	}
}

// Внутренние структуры для парсинга JSON от journalka
type scheduleResponse struct {
	Items []scheduleItem `json:"items"`
}

type scheduleItem struct {
	Date            string `json:"date"`
	Lesson          int    `json:"lesson"`
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
	StartMinutes    int    `json:"startMinutes"`
	EndMinutes      int    `json:"endMinutes"`
	SubjectFallback string `json:"subjectFallback"`
	SubjectKey      string `json:"subjectKey"`
	Teacher         string `json:"teacher"`
	Room            string `json:"room"`
}

type examsResponse struct {
	Items []examItem `json:"items"`
}

type examItem struct {
	Date string `json:"date"`
	Spec string `json:"spec"`
}

// auth выполняет POST /api/auth/login и сохраняет сессионные куки в client.Jar
func (p *Parser) auth(ctx context.Context) error {
	payload := map[string]string{
		"username": p.login,
		"password": p.password,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/auth/login", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest ||
		resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusForbidden ||
		resp.StatusCode == http.StatusUnprocessableEntity {
		return errors.New("journal auth failed: invalid login or password")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("journal login unexpected status: %d", resp.StatusCode)
	}

	p.loggedIn = true
	return nil
}

// getJSON делает GET-запрос с авто-релогином при 401/403
func (p *Parser) getJSON(ctx context.Context, path string, target any) error {
	for attempt := 0; attempt < 2; attempt++ {
		if !p.loggedIn || attempt == 1 {
			if err := p.auth(ctx); err != nil {
				return err
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := p.client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			resp.Body.Close()
			p.loggedIn = false
			continue
		}

		if resp.StatusCode >= 400 {
			resp.Body.Close()
			return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, path)
		}

		err = json.NewDecoder(resp.Body).Decode(target)
		resp.Body.Close()
		return err
	}

	return errors.New("unauthorized after retry")
}

// FetchTodaySchedule забирает пары и экзамены на дату now
func (p *Parser) FetchTodaySchedule(ctx context.Context, now time.Time) ([]domain.Class, *domain.ExamInfo, error) {
	todayStr := now.Format("2006-01-02")

	// Вычисляем понедельник и воскресенье текущей недели (как делает бот)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Воскресенье = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1)).Format("2006-01-02")
	sunday := now.AddDate(0, 0, 7-weekday).Format("2006-01-02")

	// 1. Запрашиваем расписание на неделю
	schedPath := fmt.Sprintf("/api/dashboard/schedule?scope=week&dateStart=%s&dateEnd=%s", monday, sunday)
	var schedResp scheduleResponse
	if err := p.getJSON(ctx, schedPath, &schedResp); err != nil {
		return nil, nil, fmt.Errorf("fetch schedule: %w", err)
	}

	// Фильтруем только сегодняшние пары и сортируем по времени начала
	var todayItems []scheduleItem
	for _, it := range schedResp.Items {
		if it.Date == todayStr {
			todayItems = append(todayItems, it)
		}
	}
	sort.Slice(todayItems, func(i, j int) bool {
		return todayItems[i].StartMinutes < todayItems[j].StartMinutes
	})

	classes := make([]domain.Class, 0, len(todayItems))
	for _, it := range todayItems {
		subj := it.SubjectFallback
		if subj == "" {
			subj = it.SubjectKey
		}
		if subj == "" {
			subj = "—"
		}

		dur := 90
		if it.EndMinutes > it.StartMinutes {
			dur = it.EndMinutes - it.StartMinutes
		}

		classes = append(classes, domain.Class{
			NameSubject: subj,
			Audience:    it.Room,
			StartTime:   it.StartTime,
			EndTime:     it.EndTime,
			Duration:    dur,
		})
	}

	// 2. Запрашиваем будущие экзамены (ошибка экзаменов не должна ломать отдачу расписания)
	var exam *domain.ExamInfo
	var examsResp examsResponse
	if err := p.getJSON(ctx, "/api/dashboard/home/future-exams", &examsResp); err == nil {
		for _, ex := range examsResp.Items {
			if ex.Date == todayStr {
				spec := ex.Spec
				if spec == "" {
					spec = "Экзамен"
				}
				exam = &domain.ExamInfo{
					NameSubject: spec,
					Date:        ex.Date,
				}
				break
			}
		}
	}

	return classes, exam, nil
}
