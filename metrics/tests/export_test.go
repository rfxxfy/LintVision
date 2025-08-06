package metrics_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rfxxfy/LintVision/metrics"
)

// Вспомогательная функция для создания ProjectStats для тестов
func sampleStats() metrics.ProjectStats {
	return metrics.ProjectStats{
		Files: []metrics.FileStats{
			{
				Path:         "main.go",
				Ext:          ".go",
				Category:     "code",
				LinesTotal:   10,
				LinesCode:    8,
				LinesComments: 1,
				LinesBlank:   1,
			},
		},
		CategoryCounts: map[string]int{"code": 1},
	}
}

func TestPrintStats(t *testing.T) {
	stats := sampleStats()

	// Перехватываем stdout
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	metrics.PrintStats(stats)

	w.Close()
	os.Stdout = stdout
	buf.ReadFrom(r)

	// Проверяем, что вывод содержит json с нужными полями
	out := buf.String()
	if !strings.Contains(out, `"main.go"`) || !strings.Contains(out, `"LinesTotal": 10`) {
		t.Errorf("PrintStats output = %q, want json with file info", out)
	}
}

func TestSaveStats(t *testing.T) {
	stats := sampleStats()
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "stats.json")

	err := metrics.SaveStats(stats, outFile)
	if err != nil {
		t.Fatalf("SaveStats error: %v", err)
	}

	// Проверяем, что файл создан и содержит корректный json
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("cannot read output file: %v", err)
	}
	var got metrics.ProjectStats
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("output is not valid json: %v", err)
	}
	if len(got.Files) != 1 || got.Files[0].Path != "main.go" {
		t.Errorf("SaveStats output = %+v, want file info", got)
	}
}

func TestSaveStats_Error(t *testing.T) {
	stats := sampleStats()
	// Пытаемся сохранить в несуществующую директорию
	err := metrics.SaveStats(stats, "/nonexistent_dir/stats.json")
	if err == nil {
		t.Error("SaveStats should fail for invalid path")
	}
}

func TestAnalyzeAndSave(t *testing.T) {
	// Создаём временную директорию с одним файлом
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "main.go")
	os.WriteFile(filePath, []byte("package main\n"), 0644)

	outFile := filepath.Join(tmpDir, "out.json")

	// AnalyzeAndSave должен вернуть ProjectStats и создать файл
	stats, err := metrics.AnalyzeAndSave(tmpDir, outFile)
	if err != nil {
		t.Fatalf("AnalyzeAndSave error: %v", err)
	}
	if len(stats.Files) == 0 {
		t.Error("AnalyzeAndSave: no files found")
	}
	// Проверяем, что файл создан
	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("AnalyzeAndSave: output file not created: %v", err)
	}
}