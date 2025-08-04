package metrics

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileStats struct {
	Path          string `json:"path,omitempty"`
	Ext           string `json:"ext"`
	LinesTotal    int    `json:"lines_total"`
	LinesCode     int    `json:"lines_code"`
	LinesComments int    `json:"lines_comments"`
	LinesBlank    int    `json:"lines_blank"`
}

type ProjectStats struct {
	Files []FileStats `json:"files"`
}

var codeExts = map[string]bool{
	".go":   true,
	".py":   true,
	".js":   true,
	".java": true,
	".c":    true,
	".cpp":  true,
	".h":    true,
	".cs":   true,
	".rb":   true,
	".php":  true,
	".ts":   true,
	".txt":  true,
	".md":   true,
	".mod":  true,
	".yml":  true,
}

func isCommentAfterCode(line, ext string) bool {
	var commentToken string

	if ext == ".py" {
		commentToken = "#"
	} else {
		commentToken = "//"
	}

	idx := strings.Index(line, commentToken)
	if idx <= 0 {
		return false
	}

	count := 0
	for i := 0; i < idx; i++ {
		if line[i] == '"' || (ext == ".py" && line[i] == '\'') {
			count++
		}
	}
	return count%2 == 0
}

func TraversePath(path string) (*ProjectStats, error) {
	var stats ProjectStats

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(p)
		if !codeExts[ext] {
			return nil
		}

		file, err := os.Open(p)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Printf("error closing file %s: %v\n", p, err)
			}
		}(file)

		var total, blank, comments, code int
		inBlockComment := false

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			total++

			if line == "" {
				blank++
				continue
			}

			// Block comments /* ... */
			if inBlockComment {
				comments++
				if strings.Contains(line, "*/") {
					inBlockComment = false
				}
				continue
			}
			if strings.HasPrefix(line, "/*") {
				comments++
				if !strings.Contains(line, "*/") {
					inBlockComment = true
				}
				continue
			}

			// Single-line comments
			if strings.HasPrefix(line, "//") || (ext == ".py" && strings.HasPrefix(line, "#")) {
				comments++
				continue
			}

			// Inline comments after code
			if isCommentAfterCode(line, ext) {
				code++
				comments++
				continue
			}

			code++
		}

		stats.Files = append(stats.Files, FileStats{
			Path:          p,
			Ext:           ext,
			LinesTotal:    total,
			LinesCode:     code,
			LinesComments: comments,
			LinesBlank:    blank,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &stats, nil
}
