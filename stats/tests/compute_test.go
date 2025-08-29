package stats_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rfxxfy/LintVision/stats"
	"github.com/stretchr/testify/assert"
)

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
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		content  string
		ext      string
		want     stats.FileStats
	}{
		{
			name:     "Go file with code, comment, blank",
			filename: "main.go",
			content: `package main

// This is a comment
func main() {}`,
			ext: ".go",
			want: stats.FileStats{
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
			want: stats.FileStats{
				Ext:           ".py",
				Category:      "code",
				LinesTotal:    3,
				LinesCode:     1,
				LinesComments: 2,
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
			want: stats.FileStats{
				Ext:        ".md",
				Category:   "markup",
				LinesTotal: 4,
				LinesBlank: 2,
			},
		},
		{
			name:     "Unknown extension",
			filename: "data.bin",
			content:  "some data\n",
			ext:      ".bin",
			want: stats.FileStats{
				Ext:      ".bin",
				Category: "binary",
			},
		},
		{
			name:     "File does not exist",
			filename: "nofile.go",
			content:  "",
			ext:      ".go",
			want: stats.FileStats{
				Ext:      ".go",
				Category: "code",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			var path string
			if tt.name != "File does not exist" {
				path = createTempFile(t, tmpDir, tt.filename, tt.content)
			} else {
				path = filepath.Join(tmpDir, tt.filename)
			}
			got, _ := stats.ComputeFileStats(path)
			assert.Equal(t, tt.want.Ext, got.Ext)
			assert.Equal(t, tt.want.Category, got.Category)
			assert.Equal(t, tt.want.LinesTotal, got.LinesTotal)
			assert.Equal(t, tt.want.LinesCode, got.LinesCode)
			assert.Equal(t, tt.want.LinesComments, got.LinesComments)
			assert.Equal(t, tt.want.LinesBlank, got.LinesBlank)
		})
	}
}

func TestComputeProjectStats(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	files := map[string]string{
		"main.go":   "package main\n\n// comment\nfunc main() {}",
		"script.py": "# comment\nprint('hi')\n",
		"README.md": "# Title\n\nSome text\n",
		"empty.txt": "",
	}
	var paths []string
	for name, content := range files {
		paths = append(paths, createTempFile(t, tmpDir, name, content))
	}

	got, err := stats.ComputeProjectStats(paths)
	assert.NoError(t, err)
	assert.Equal(t, len(files), len(got.Files))

	wantCategories := map[string]int{
		"code":     2,
		"markup":   1,
		"document": 1,
	}
	assert.Equal(t, wantCategories, got.CategoryCounts)
}
