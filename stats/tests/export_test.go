package stats_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rfxxfy/LintVision/stats"
	"github.com/stretchr/testify/assert"
)

func sampleStats() stats.ProjectStats {
	return stats.ProjectStats{
		Files: []stats.FileStats{
			{
				Path:          "main.go",
				Ext:           ".go",
				Category:      "code",
				LinesTotal:    10,
				LinesCode:     8,
				LinesComments: 1,
				LinesBlank:    1,
			},
		},
		CategoryCounts: map[string]int{"code": 1},
	}
}

func TestPrintStats(t *testing.T) {
	t.Parallel()
	testStats := sampleStats()

	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	stats.PrintStats(testStats)

	w.Close()
	os.Stdout = stdout
	buf.ReadFrom(r)

	out := buf.String()
	assert.Contains(t, out, `"main.go"`)
	assert.Contains(t, out, `"lines_total": 10`)
}

func TestSaveStats(t *testing.T) {
	t.Parallel()
	testStats := sampleStats()
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "stats.json")

	err := stats.SaveStats(testStats, outFile)
	assert.NoError(t, err)

	data, err := os.ReadFile(outFile)
	assert.NoError(t, err)
	var got stats.ProjectStats
	err = json.Unmarshal(data, &got)
	assert.NoError(t, err)
	assert.Len(t, got.Files, 1)
	assert.Equal(t, "main.go", got.Files[0].Path)
}

func TestSaveStats_Error(t *testing.T) {
	t.Parallel()
	testStats := sampleStats()
	err := stats.SaveStats(testStats, "/nonexistent_dir/stats.json")
	assert.Error(t, err)
}

func TestAnalyzeAndSave(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(filePath, []byte("package main\n"), 0644)
	assert.NoError(t, err)

	outFile := filepath.Join(tmpDir, "out.json")

	result, err := stats.AnalyzeAndSave(tmpDir, outFile)
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Files)
	_, err = os.Stat(outFile)
	assert.NoError(t, err)
}
