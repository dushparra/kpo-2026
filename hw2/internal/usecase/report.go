package usecase

type Report struct {
	Rows []ReportRow
}

type ReportRow struct {
	StudentID         int64   `json:"student_id"`
	StudentName       string  `json:"student_name"`
	Score             float64 `json:"score"`
	EvaluatedCriteria int     `json:"evaluated_criteria"`
	TotalCriteria     int     `json:"total_criteria"`
}
