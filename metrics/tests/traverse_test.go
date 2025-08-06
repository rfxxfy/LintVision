package metrics_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rfxxfy/LintVision/metrics"
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
			wantNonHidD: 1, // dir
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
			wantNonHidD: 2, // dir, dir/visible
		},
		{
			name: "only hidden",
			files: map[string]string{
				".hidden": "data",
			},
			wantFiles:   []string{".hidden"},
			wantHiddenF: 1,
			wantHiddenD: 0,
			wantNonHidD: 0,
		},
		{
			name: "empty dir",
			files: map[string]string{},
			wantFiles:   []string{},
			wantHiddenF: 0,
			wantHiddenD: 0,
			wantNonHidD: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			createTestTree(t, tmpDir, tt.files)

			gotFiles, gotHiddenF, gotHiddenD, gotNonHidD, err := metrics.ScanDir(tmpDir)
			if err != nil {
				t.Fatalf("ScanDir error: %v", err)
			}

			// Приводим к относительным путям
			for i, path := range gotFiles {
				rel, _ := filepath.Rel(tmpDir, path)
				gotFiles[i] = rel
			}

			// Сортируем для сравнения, если порядок не важен
			want := make([]string, len(tt.wantFiles))
			copy(want, tt.wantFiles)
			if !reflect.DeepEqual(gotFiles, want) {
				t.Errorf("files = %v, want %v", gotFiles, want)
			}
			if gotHiddenF != tt.wantHiddenF {
				t.Errorf("hidden files = %d, want %d", gotHiddenF, tt.wantHiddenF)
			}
			if gotHiddenD != tt.wantHiddenD {
				t.Errorf("hidden dirs = %d, want %d", gotHiddenD, tt.wantHiddenD)
			}
			if gotNonHidD != tt.wantNonHidD {
				t.Errorf("non-hidden dirs = %d, want %d", gotNonHidD, tt.wantNonHidD)
			}
		})
	}
}

func TestScanDir_Error(t *testing.T) {
	// Передаём несуществующую директорию
	_, _, _, _, err := metrics.ScanDir("/nonexistent_dir_12345")
	if err == nil {
		t.Error("ScanDir should return error for nonexistent dir")
	}
}

func TestComputeProjectStatsFromDir(t *testing.T) {
	tmpDir := t.TempDir()
	files := map[string]string{
		"main.go":     "package main",
		".hidden.py":  "# hidden",
		"dir/a.py":    "print('a')",
	}
	createTestTree(t, tmpDir, files)

	ps, err := metrics.ComputeProjectStatsFromDir(tmpDir)
	if err != nil {
		t.Fatalf("ComputeProjectStatsFromDir error: %v", err)
	}
	if len(ps.Files) != 3 {
		t.Errorf("Files = %d, want 3", len(ps.Files))
	}
	if ps.HiddenFiles != 1 {
		t.Errorf("HiddenFiles = %d, want 1", ps.HiddenFiles)
	}
	if ps.NonHiddenDirs != 1 {
		t.Errorf("NonHiddenDirs = %d, want 1", ps.NonHiddenDirs)
	}
}