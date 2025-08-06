package metrics_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rfxxfy/LintVision/metrics"
)

// Вспомогательная функция для создания временного файла с содержимым
func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	return path
}

func TestComputeFileStats(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
		ext      string
		want     metrics.FileStats
	}{
		{
			name:     "Go file with code, comment, blank",
			filename: "main.go",
			content: `package main

// This is a comment
func main() {}`,
			ext: ".go",
			want: metrics.FileStats{
				Ext:           ".go",
				Category:      "code",
				LinesTotal:    4,
				LinesCode:     2,
				LinesComments: 1,
				LinesBlank:    1,
			},
		},
		{
			name:     "Python file with code, comment, blank",
			filename: "script.py",
			content: `
# comment
print("hi")  # inline comment
`,
			ext: ".py",
			want: metrics.FileStats{
				Ext:           ".py",
				Category:      "code",
				LinesTotal:    4,
				LinesCode:     1,
				LinesComments: 2, // одна строка - только комментарий, одна - код с комментом
				LinesBlank:    1,
			},
		},
		{
			name:     "Markdown file (markup)",
			filename: "README.md",
			content: `
# Title

Some text
`,
			ext: ".md",
			want: metrics.FileStats{
				Ext:        ".md",
				Category:   "markup",
				LinesTotal: 4,
				LinesBlank: 1,
			},
		},
		{
			name:     "Unknown extension",
			filename: "data.bin",
			content:  "some data\n",
			ext:      ".bin",
			want: metrics.FileStats{
				Ext:      ".bin",
				Category: "",
			},
		},
		{
			name:     "File does not exist",
			filename: "nofile.go",
			content:  "",
			ext:      ".go",
			want: metrics.FileStats{
				Ext:      ".go",
				Category: "code",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			var path string
			if tt.name != "File does not exist" {
				path = createTempFile(t, tmpDir, tt.filename, tt.content)
			} else {
				path = filepath.Join(tmpDir, tt.filename)
			}
			got, _ := metrics.ComputeFileStats(path)
			// Сравниваем только интересующие поля
			if got.Ext != tt.want.Ext ||
				got.Category != tt.want.Category ||
				got.LinesTotal != tt.want.LinesTotal ||
				got.LinesCode != tt.want.LinesCode ||
				got.LinesComments != tt.want.LinesComments ||
				got.LinesBlank != tt.want.LinesBlank {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestComputeProjectStats(t *testing.T) {
	tmpDir := t.TempDir()
	files := map[string]string{
		"main.go":    "package main\n\n// comment\nfunc main() {}",
		"script.py":  "# comment\nprint('hi')\n",
		"README.md":  "# Title\n\nSome text\n",
		"empty.txt":  "",
	}
	var paths []string
	for name, content := range files {
		paths = append(paths, createTempFile(t, tmpDir, name, content))
	}

	got, err := metrics.ComputeProjectStats(paths)
	if err != nil {
		t.Fatalf("ComputeProjectStats error: %v", err)
	}

	// Проверяем количество файлов
	if len(got.Files) != len(files) {
		t.Errorf("got %d files, want %d", len(got.Files), len(files))
	}

	// Проверяем, что категории подсчитаны корректно
	wantCategories := map[string]int{
		"code":   2,
		"markup": 1,
		"":       1,
	}
	if !reflect.DeepEqual(got.CategoryCounts, wantCategories) {
		t.Errorf("got categories %v, want %v", got.CategoryCounts, wantCategories)
	}
}