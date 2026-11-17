package exporter

import (
	"encoding/json"
	"os"

	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

type JSON struct{}

func (JSON) Export(report usecase.Report, path string) error {
	rows := report.Rows
	if rows == nil {
		rows = []usecase.ReportRow{}
	}
	output, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	output = append(output, '\n')
	return os.WriteFile(path, output, 0o644)
}
