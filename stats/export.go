package stats

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rfxxfy/LintVision/logging"
)

func PrintStats(stats ProjectStats) {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		logging.Error("PrintStats: JSON marshal failed: %v", err)
		return
	}
	fmt.Println(string(data))
}

func SaveStats(stats ProjectStats, filePath string) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		logging.Error("SaveStats: JSON marshal failed: %v", err)
		return err
	}
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		logging.Error("SaveStats: cannot write to %s: %v", filePath, err)
		return err
	}
	logging.Info("SaveStats: written result to %s", filePath)
	return nil
}

func AnalyzeAndSave(root, outPath string) (ProjectStats, error) {
	stats, err := ComputeProjectStatsFromDir(root)
	if err != nil {
		logging.Error("AnalyzeAndSave: ComputeProjectStatsFromDir failed: %v", err)
		return stats, err
	}
	PrintStats(stats)
	if outPath != "" {
		if err := SaveStats(stats, outPath); err != nil {
			return stats, err
		}
	}
	return stats, nil
}
