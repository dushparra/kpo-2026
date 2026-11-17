package exporter

import (
	"bytes"
	"encoding/csv"
	"os"
	"strconv"

	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

type CSV struct{}

func (CSV) Export(report usecase.Report, path string) error {
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write([]string{
		"student_id", "student_name", "score", "evaluated_criteria", "total_criteria",
	}); err != nil {
		return err
	}
	for _, row := range report.Rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.StudentID, 10),
			row.StudentName,
			strconv.FormatFloat(row.Score, 'f', 2, 64),
			strconv.Itoa(row.EvaluatedCriteria),
			strconv.Itoa(row.TotalCriteria),
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return err
	}
	return os.WriteFile(path, output.Bytes(), 0o644)
}
