package domain

type Class struct {
	NameSubject string
	Audience    int
	StartTime   string
	EndTime     string
	Duration    int
}

type ExamInfo struct {
	NameSubject string
	Audience    int
	Lesson      int
}

type WidgetResponse struct {
	HasExam      bool
	Exam         *ExamInfo `json:"exam"`
	NextClass    *Class    `json:"next_class"`
	CurrentClass *Class    `json:"current_class"`
	UpdateAt     string
}
