package parseurl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/rfxxfy/LintVision/metrics"
)

func AnalyzeRepoFromURL(repoURL string) (metrics.ProjectStats, error) {

	tempDir, err := createTempDir()
	if err != nil {
		//todo logging
		return metrics.ProjectStats{}, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	defer func() {
		if cleanupErr := cleanupTempDir(tempDir); cleanupErr != nil {
			//todo logging
		}
	}()

	if err := cloneRepo(repoURL, tempDir); err != nil {
		//todo logging
		return metrics.ProjectStats{}, fmt.Errorf("failed to clone repository: %w", err)
	}

	//todo logging - Repository successfully cloned

	stats, err := metrics.ComputeProjectStatsFromDir(tempDir)
	if err != nil {
		//todo logging
		return metrics.ProjectStats{}, fmt.Errorf("error analyzing files: %w", err)
	}

	//todo logging - Repository analysis completed

	return stats, nil
}

func createTempDir() (string, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "lintvision-repo-*")
	if err != nil {
		return "", err
	}
	//todo logging - Temporary directory created
	return tempDir, nil
}

func cloneRepo(repoURL, destDir string) error {
	//todo logging - Cloning repository

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", repoURL, destDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			//todo logging - Clone operation timed out
			return fmt.Errorf("git clone timed out after 2 minutes: %w", ctx.Err())
		}
		//todo logging
		return fmt.Errorf("git clone error: %w: %s", err, output)
	}

	return nil
}

func cleanupTempDir(dirPath string) error {
	//todo logging - Removing temporary directory
	err := os.RemoveAll(dirPath)
	if err != nil {
		return err
	}
	return nil
}
