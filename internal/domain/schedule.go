package domain

type Class struct {
	NameSubject string
	Audience    string
	StartTime   string
	EndTime     string
	Duration    int
}

type ExamInfo struct {
	NameSubject string
	Date        string
}

type WidgetResponse struct {
	HasExam      bool
	Exam         *ExamInfo `json:"exam,omitempty"`
	NextClass    *Class    `json:"next_class,omitempty"`
	CurrentClass *Class    `json:"current_class,omitempty"`
	UpdateAt     string
}
