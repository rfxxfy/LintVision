package metrics

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rfxxfy/LintVision/logging"
)

func ScanDir(root string) ([]string, int, int, int, error) {
	var paths []string
	hiddenFiles := 0
	hiddenDirs := 0
	nonHiddenDirs := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logging.Error("ScanDir: error accessing %s: %v", path, err)
			return err
		}

		name := d.Name()
		isHidden := strings.HasPrefix(name, ".")

		if d.IsDir() {
			if isHidden {
				hiddenDirs++
				return fs.SkipDir
			} else {
				nonHiddenDirs++
			}
			return nil
		}

		if isHidden {
			hiddenFiles++
		}
		paths = append(paths, path)
		return nil
	})

	if err != nil {
		logging.Error("ScanDir: walk error on %s: %v", root, err)
	} else {
		logging.Info("ScanDir: found %d files (%d hidden files) under %s; dirs: %d hidden, %d non-hidden",
			len(paths), hiddenFiles, root, hiddenDirs, nonHiddenDirs)
	}

	return paths, hiddenFiles, hiddenDirs, nonHiddenDirs, err
}

func ComputeProjectStatsFromDir(root string) (ProjectStats, error) {
	files, hf, hd, nhd, err := ScanDir(root)
	if err != nil {
		return ProjectStats{}, err
	}

	ps, err := ComputeProjectStats(files)
	if err != nil {
		return ps, err
	}

	ps.HiddenFiles = hf
	ps.HiddenDirs = hd
	ps.NonHiddenDirs = nhd
	return ps, nil
}
