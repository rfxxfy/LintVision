package metrics

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfxxfy/LintVision/extensions"
	"github.com/rfxxfy/LintVision/logging"
)

func ComputeFileStats(path string) (FileStats, error) {
	ext := filepath.Ext(path)
	cat := extensions.GetFileCategory(ext)

	fs := FileStats{Path: path, Ext: ext, Category: cat}
	if cat != "code" && cat != "markup" {
		return fs, nil
	}

	f, err := os.Open(path)
	if err != nil {
		logging.Error("ComputeFileStats: cannot open %s: %v", path, err)
		return fs, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		fs.LinesTotal++

		switch cat {
		case "code":
			cfg, _ := extensions.GetLanguageConfig(ext)
			token := cfg.SingleLineCommentToken
			switch {
			case trimmed == "":
				fs.LinesBlank++
			case token != "" && strings.HasPrefix(trimmed, token):
				fs.LinesComments++
			case token != "" && extensions.IsCommentAfterCode(line, ext):
				fs.LinesComments++
				fs.LinesCode++
			default:
				fs.LinesCode++
			}

		case "markup":
			if trimmed == "" {
				fs.LinesBlank++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logging.Error("ComputeFileStats: scanner error in %s: %v", path, err)
		return fs, err
	}
	return fs, nil
}

// ComputeProjectStats аккумулирует FileStats по списку путей и
// возвращает готовый ProjectStats (без данных о скрытых, они заполняются ниже).
func ComputeProjectStats(paths []string) (ProjectStats, error) {
	ps := ProjectStats{
		CategoryCounts: make(map[string]int),
	}
	for _, p := range paths {
		stat, err := ComputeFileStats(p)
		if err != nil {
			logging.Error("ComputeProjectStats: error computing %s: %v", p, err)
			return ps, err
		}
		ps.Files = append(ps.Files, stat)
		ps.CategoryCounts[stat.Category]++
	}
	logging.Info("ComputeProjectStats: processed %d files", len(ps.Files))
	return ps, nil
}
