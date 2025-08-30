package parseurl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type ValidationResult struct {
	IsValid     bool
	IsGitHub    bool
	Exists      bool
	IsPublic    bool
	Error       string
	Suggestions []string
}

func ValidateGitHubURL(repoURL string) ValidationResult {
	result := ValidationResult{}

	if !isValidURL(repoURL) {
		result.Error = "Неверный формат URL"
		result.Suggestions = []string{
			"Убедитесь, что URL начинается с http:// или https://",
			"Проверьте правильность написания URL",
		}
		return result
	}

	if !isGitHubURL(repoURL) {
		result.Error = "Это не GitHub URL"
		result.Suggestions = []string{
			"Используйте ссылку на GitHub репозиторий",
			"Формат: https://github.com/username/repository",
		}
		return result
	}
	result.IsGitHub = true

	if exists, err := checkRepoExists(repoURL); err != nil {
		result.Error = fmt.Sprintf("Ошибка проверки репозитория: %v", err)
		result.Suggestions = []string{
			"Проверьте интернет-соединение",
			"Попробуйте позже",
		}
		return result
	} else if !exists {
		result.Error = "Репозиторий не найден"
		result.Suggestions = []string{
			"Проверьте правильность названия репозитория",
			"Убедитесь, что репозиторий существует",
			"Проверьте правильность имени пользователя",
		}
		return result
	}
	result.Exists = true

	// 4. Проверка доступности (открытости) репозитория
	if isPublic, err := checkRepoAccessibility(repoURL); err != nil {
		result.Error = fmt.Sprintf("Ошибка проверки доступности: %v", err)
		result.Suggestions = []string{
			"Попробуйте позже",
			"Проверьте интернет-соединение",
		}
		return result
	} else if !isPublic {
		result.Error = "Репозиторий недоступен"
		result.Suggestions = []string{
			"Репозиторий может быть приватным",
			"Требуется авторизация",
			"Нет прав доступа",
		}
		return result
	}
	result.IsPublic = true

	// Все проверки пройдены
	result.IsValid = true
	return result
}

func isValidURL(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func isGitHubURL(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if parsedURL.Host != "github.com" && parsedURL.Host != "www.github.com" {
		return false
	}

	path := strings.Trim(parsedURL.Path, "/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) < 2 {
		return false
	}

	if pathParts[0] == "" || pathParts[1] == "" {
		return false
	}

	pathParts[1] = strings.TrimSuffix(pathParts[1], ".git")

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`)
	repoRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

	return usernameRegex.MatchString(pathParts[0]) && repoRegex.MatchString(pathParts[1])
}

func checkRepoExists(repoURL string) (bool, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return false, err
	}

	path := strings.Trim(parsedURL.Path, "/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		return false, fmt.Errorf("неверный формат GitHub URL")
	}

	username := pathParts[0]
	repoName := strings.TrimSuffix(pathParts[1], ".git")

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", username, repoName)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", "LintVision/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	case 403:
		return true, nil
	default:
		return false, fmt.Errorf("неожиданный статус ответа: %d", resp.StatusCode)
	}
}

func checkRepoAccessibility(repoURL string) (bool, error) {
	tempDir, err := createTempDir()
	if err != nil {
		return false, fmt.Errorf("не удалось создать временную директорию: %w", err)
	}
	defer cleanupTempDir(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", "--quiet", repoURL, tempDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		outputStr := string(output)

		if strings.Contains(outputStr, "Repository not found") || strings.Contains(outputStr, "404") {
			return false, fmt.Errorf("репозиторий не найден")
		}
		if strings.Contains(outputStr, "Authentication failed") || strings.Contains(outputStr, "403") {
			return false, fmt.Errorf("репозиторий приватный или требует аутентификации")
		}
		if strings.Contains(outputStr, "Permission denied") {
			return false, fmt.Errorf("нет доступа к репозиторию")
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return false, fmt.Errorf("превышено время ожидания при проверке доступности")
		}

		return false, fmt.Errorf("ошибка проверки доступности: %s", strings.TrimSpace(outputStr))
	}

	return true, nil
}
