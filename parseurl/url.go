package parseurl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rfxxfy/LintVision/logging"
	"github.com/rfxxfy/LintVision/stats"
)

func AnalyzeRepoFromURL(repoURL string) (stats.ProjectStats, error) {
	logging.Info("AnalyzeRepoFromURL: starting analysis for %s", repoURL)

	tempDir, err := createTempDir()
	if err != nil {
		logging.Error("AnalyzeRepoFromURL: failed to create temporary directory for %s: %v", repoURL, err)
		return stats.ProjectStats{}, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	defer func() {
		if cleanupErr := cleanupTempDir(tempDir); cleanupErr != nil {
			logging.Error("AnalyzeRepoFromURL: failed to cleanup temporary directory %s: %v", tempDir, cleanupErr)
		} else {
			logging.Info("AnalyzeRepoFromURL: cleaned up temporary directory %s", tempDir)
		}
	}()

	if err := cloneRepo(repoURL, tempDir); err != nil {
		logging.Error("AnalyzeRepoFromURL: failed to clone repository %s into %s: %v", repoURL, tempDir, err)
		return stats.ProjectStats{}, fmt.Errorf("failed to clone repository: %w", err)
	}
	logging.Info("AnalyzeRepoFromURL: repository %s successfully cloned into %s", repoURL, tempDir)

	result, err := stats.ComputeProjectStatsFromDir(tempDir)
	if err != nil {
		logging.Error("AnalyzeRepoFromURL: error analyzing files in %s for %s: %v", tempDir, repoURL, err)
		return stats.ProjectStats{}, fmt.Errorf("error analyzing files: %w", err)
	}
	logging.Info("AnalyzeRepoFromURL: repository analysis completed for %s", repoURL)

	return result, nil
}

func createTempDir() (string, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "lintvision-repo-*")
	if err != nil {
		logging.Error("createTempDir: failed to create temporary directory: %v", err)
		return "", err
	}
	logging.Info("createTempDir: created temporary directory %s", tempDir)
	return tempDir, nil
}

func cloneRepo(repoURL, destDir string) error {
	logging.Info("cloneRepo: cloning repository %s into %s", repoURL, destDir)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", repoURL, destDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logging.Error("cloneRepo: git clone timed out for %s: %v", repoURL, ctx.Err())
			return fmt.Errorf("git clone timed out after 2 minutes: %w", ctx.Err())
		}

		outputStr := string(output)
		if strings.Contains(outputStr, "Repository not found") || strings.Contains(outputStr, "404") {
			return fmt.Errorf("репозиторий не найден или не существует: %s", repoURL)
		}
		if strings.Contains(outputStr, "Authentication failed") || strings.Contains(outputStr, "403") {
			return fmt.Errorf("репозиторий закрытый или требует аутентификации: %s", repoURL)
		}
		if strings.Contains(outputStr, "Permission denied") {
			return fmt.Errorf("нет доступа к репозиторию: %s", repoURL)
		}
		if strings.Contains(outputStr, "fatal:") {
			return fmt.Errorf("ошибка git: %s", strings.TrimSpace(outputStr))
		}

		logging.Error("cloneRepo: git clone error for %s: %v: %s", repoURL, err, outputStr)
		return fmt.Errorf("ошибка клонирования: %s", strings.TrimSpace(outputStr))
	}

	logging.Info("cloneRepo: git clone completed for %s", repoURL)
	return nil
}

func cleanupTempDir(dirPath string) error {
	logging.Info("cleanupTempDir: removing temporary directory %s", dirPath)
	err := os.RemoveAll(dirPath)
	if err != nil {
		logging.Error("cleanupTempDir: failed to remove %s: %v", dirPath, err)
		return err
	}
	logging.Info("cleanupTempDir: removed temporary directory %s", dirPath)
	return nil
}
