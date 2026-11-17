package exporter_test

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dushparra/kpo-2026/hw2/internal/adapters/exporter"
	"github.com/dushparra/kpo-2026/hw2/internal/usecase"
)

func TestCSVAndJSONExportSameReportWithoutLosingText(t *testing.T) {
	name := "Иван, \"Петров\"\nБПИ255"
	report := usecase.Report{Rows: []usecase.ReportRow{{
		StudentID:         1,
		StudentName:       name,
		Score:             6.25,
		EvaluatedCriteria: 1,
		TotalCriteria:     2,
	}}}
	directory := t.TempDir()
	csvPath := filepath.Join(directory, "results.csv")
	jsonPath := filepath.Join(directory, "results.json")
	if err := (exporter.CSV{}).Export(report, csvPath); err != nil {
		t.Fatal(err)
	}
	if err := (exporter.JSON{}).Export(report, jsonPath); err != nil {
		t.Fatal(err)
	}
	csvData, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(strings.NewReader(string(csvData))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || len(records[1]) != 5 {
		t.Fatalf("unexpected CSV shape: %+v", records)
	}
	if records[0][0] != "student_id" || records[0][4] != "total_criteria" {
		t.Fatalf("incorrect CSV header: %+v", records[0])
	}
	if records[1][0] != "1" || records[1][1] != name || records[1][2] != "6.25" || records[1][3] != "1" || records[1][4] != "2" {
		t.Fatalf("CSV lost or changed report fields: %+v", records[1])
	}

	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		StudentID         int64   `json:"student_id"`
		StudentName       string  `json:"student_name"`
		Score             float64 `json:"score"`
		EvaluatedCriteria int     `json:"evaluated_criteria"`
		TotalCriteria     int     `json:"total_criteria"`
	}
	if err := json.Unmarshal(jsonData, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].StudentID != 1 || rows[0].StudentName != name || rows[0].Score != 6.25 || rows[0].EvaluatedCriteria != 1 || rows[0].TotalCriteria != 2 {
		t.Fatalf("JSON lost or changed report fields: %+v", rows)
	}
}

func TestExportersWriteAnEmptyReportAndReturnFileErrors(t *testing.T) {
	for _, test := range []struct {
		name     string
		exporter usecase.Exporter
	}{
		{name: "csv", exporter: exporter.CSV{}},
		{name: "json", exporter: exporter.JSON{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "empty."+test.name)
			if err := test.exporter.Export(usecase.Report{}, path); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if test.name == "json" && strings.TrimSpace(string(data)) != "[]" {
				t.Fatalf("empty JSON report must be [], got %s", data)
			}
			if test.name == "csv" {
				records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
				if err != nil || len(records) != 1 {
					t.Fatalf("empty CSV report must keep header: %q, error %v", data, err)
				}
			}
			if err := test.exporter.Export(usecase.Report{}, directory); err == nil {
				t.Fatal("expected error when output path is a directory")
			}
		})
	}
}
