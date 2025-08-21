package metrics_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	//"strings"
	"testing"

	"github.com/rfxxfy/LintVision/metrics"
	"github.com/stretchr/testify/assert"
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
	t.Parallel()
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
	assert.Contains(t, out, `"main.go"`)
	assert.Contains(t, out, `"lines_total": 10`)
	// TODO: logging here
}

func TestSaveStats(t *testing.T) {
	t.Parallel()
	stats := sampleStats()
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "stats.json")

	err := metrics.SaveStats(stats, outFile)
	assert.NoError(t, err)

	// Проверяем, что файл создан и содержит корректный json
	data, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	var got metrics.ProjectStats
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)
	assert.Len(t, got.Files, 1)
	assert.Equal(t, "main.go", got.Files[0].Path)
	// TODO: logging here
}

func TestSaveStats_Error(t *testing.T) {
	t.Parallel()
	stats := sampleStats()
	// Пытаемся сохранить в несуществующую директорию
	err := metrics.SaveStats(stats, "/nonexistent_dir/stats.json")
	assert.Error(t, err)
	// TODO: logging here
}

func TestAnalyzeAndSave(t *testing.T) {
	t.Parallel()
	// Создаём временную директорию с одним файлом
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(filePath, []byte("package main\n"), 0644)
	assert.NoError(t, err)

	outFile := filepath.Join(tmpDir, "out.json")

	// AnalyzeAndSave должен вернуть ProjectStats и создать файл
	stats, err := metrics.AnalyzeAndSave(tmpDir, outFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, stats.Files)
	// Проверяем, что файл создан
	_, err = os.Stat(outFile)
	assert.NoError(t, err)
	// TODO: logging here
}