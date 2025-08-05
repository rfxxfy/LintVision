package metrics

import (
	"LintVision/extensions"
	"bufio"
	"io/fs"
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
	stats := &ProjectStats{}

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(p)
		if !extensions.IsCodeExtension(ext) {
			return nil
		}

		fileStats, err := processFile(p, ext)
		if err != nil {
			return err
		}

		stats.Files = append(stats.Files, *fileStats)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return stats, nil
}

func processFile(path string, ext string) (*FileStats, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	total, code, comments, blank := countLines(file, ext)

	return &FileStats{
		Path:          path,
		Ext:           ext,
		LinesTotal:    total,
		LinesCode:     code,
		LinesComments: comments,
		LinesBlank:    blank,
	}, nil
}

func countLines(file *os.File, ext string) (total, code, comments, blank int) {
	scanner := bufio.NewScanner(file)
	inBlockComment := false

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

	return total, code, comments, blank
}
