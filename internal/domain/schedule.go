package domain

import "time"

type Class struct {
	NameSubject string `json:"name_subject"`
	Audience    string `json:"audience"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Duration    int    `json:"duration"`
}

type ExamInfo struct {
	NameSubject string `json:"name_subject"`
	Date        string `json:"date"`
}

type WidgetResponse struct {
	HasExam      bool      `json:"has_exam"`
	Exam         *ExamInfo `json:"exam,omitempty"`
	NextClass    *Class    `json:"next_class,omitempty"`
	CurrentClass *Class    `json:"current_class,omitempty"`
	UpdateAt     string    `json:"updated_at"`
}

type User struct {
	Key       string
	Login     string
	Password  string
	UpdatedAt time.Time
}
