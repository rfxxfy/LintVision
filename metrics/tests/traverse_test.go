package metrics_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rfxxfy/LintVision/metrics"
	"github.com/stretchr/testify/assert"
)

func createTestTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for relPath, content := range files {
		fullPath := filepath.Join(root, relPath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}
}

func TestScanDir(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		files        map[string]string
		wantFiles    []string // относительные пути
		wantHiddenF  int
		wantHiddenD  int
		wantNonHidD  int
	}{
		{
			name: "simple structure",
			files: map[string]string{
				"main.go":         "package main",
				".hiddenfile":     "secret",
				"dir/util.py":     "print('hi')",
				".hiddendir/a.go": "package a",
			},
			wantFiles:   []string{"main.go", ".hiddenfile", "dir/util.py", ".hiddendir/a.go"},
			wantHiddenF: 1,
			wantHiddenD: 1,
			wantNonHidD: 2, // dir
		},
		{
			name: "nested hidden dir",
			files: map[string]string{
				"main.go":                 "package main",
				"dir/.hiddendir/a.go":     "package a",
				"dir/visible/b.py":        "print('b')",
				"dir/visible/.c.py":       "print('c')",
			},
			wantFiles:   []string{"main.go", "dir/.hiddendir/a.go", "dir/visible/b.py", "dir/visible/.c.py"},
			wantHiddenF: 1, // .c.py
			wantHiddenD: 1, // .hiddendir
			wantNonHidD: 3, // dir, dir/visible
		},
		{
			name: "only hidden",
			files: map[string]string{
				".hidden": "data",
			},
			wantFiles:   []string{".hidden"},
			wantHiddenF: 1,
			wantHiddenD: 0,
			wantNonHidD: 1,
		},
		{
			name: "empty dir",
			files: map[string]string{},
			wantFiles:   []string{},
			wantHiddenF: 0,
			wantHiddenD: 0,
			wantNonHidD: 1,
		},
	}

	for _, tt := range tests {
		tt := tt // захват переменной для параллельного запуска
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			createTestTree(t, tmpDir, tt.files)

			gotFiles, gotHiddenF, gotHiddenD, gotNonHidD, err := metrics.ScanDir(tmpDir)
			assert.NoError(t, err)

			// Приводим к относительным путям
			for i, path := range gotFiles {
				rel, _ := filepath.Rel(tmpDir, path)
				gotFiles[i] = rel
			}

			assert.ElementsMatch(t, tt.wantFiles, gotFiles, "files mismatch")
			assert.Equal(t, tt.wantHiddenF, gotHiddenF, "hidden files mismatch")
			assert.Equal(t, tt.wantHiddenD, gotHiddenD, "hidden dirs mismatch")
			assert.Equal(t, tt.wantNonHidD, gotNonHidD, "non-hidden dirs mismatch")
			// TODO: logging here
		})
	}
}

func TestScanDir_Error(t *testing.T) {
	t.Parallel()
	// Передаём несуществующую директорию
	_, _, _, _, err := metrics.ScanDir("/nonexistent_dir_12345")
	assert.Error(t, err)
	// TODO: logging here
}

func TestComputeProjectStatsFromDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	files := map[string]string{
		"main.go":     "package main",
		".hidden.py":  "# hidden",
		"dir/a.py":    "print('a')",
	}
	createTestTree(t, tmpDir, files)

	ps, err := metrics.ComputeProjectStatsFromDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, ps.Files, 3)
	assert.Equal(t, 1, ps.HiddenFiles)
	assert.Equal(t, 2, ps.NonHiddenDirs)
	// TODO: logging here
}