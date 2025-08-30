package parseurl

import (
	"testing"
)

func TestValidateGitHubURL(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		expectedIsValid  bool
		expectedIsGitHub bool
		expectedError    string
	}{
		{
			name:             "Valid GitHub URL format",
			url:              "https://github.com/username/repository",
			expectedIsValid:  false,
			expectedIsGitHub: true,
			expectedError:    "Репозиторий не найден",
		},
		{
			name:             "Valid GitHub URL with .git suffix",
			url:              "https://github.com/username/repository.git",
			expectedIsValid:  false,
			expectedIsGitHub: true,
			expectedError:    "Репозиторий не найден",
		},
		{
			name:             "Invalid URL format",
			url:              "not-a-url",
			expectedIsValid:  false,
			expectedIsGitHub: false,
			expectedError:    "Неверный формат URL",
		},
		{
			name:             "Not GitHub URL",
			url:              "https://gitlab.com/username/repository",
			expectedIsValid:  false,
			expectedIsGitHub: false,
			expectedError:    "Это не GitHub URL",
		},
		{
			name:             "GitHub URL without repository",
			url:              "https://github.com/username",
			expectedIsValid:  false,
			expectedIsGitHub: false,
			expectedError:    "Это не GitHub URL",
		},
		{
			name:             "GitHub URL with valid username with hyphens",
			url:              "https://github.com/user-name/repository",
			expectedIsValid:  false,
			expectedIsGitHub: true,
			expectedError:    "Репозиторий не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateGitHubURL(tt.url)

			if result.IsValid != tt.expectedIsValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.expectedIsValid)
			}

			if result.IsGitHub != tt.expectedIsGitHub {
				t.Errorf("IsGitHub = %v, want %v", result.IsGitHub, tt.expectedIsGitHub)
			}

			if result.Error != tt.expectedError {
				t.Errorf("Error = %v, want %v", result.Error, tt.expectedError)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"Valid HTTP URL", "http://example.com", true},
		{"Valid HTTPS URL", "https://example.com", true},
		{"Valid URL with path", "https://example.com/path", true},
		{"Invalid URL - no scheme", "example.com", false},
		{"Invalid URL - no host", "https://", false},
		{"Invalid URL - empty string", "", false},
		{"Invalid URL - malformed", "://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestIsGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"Valid GitHub URL", "https://github.com/username/repository", true},
		{"Valid GitHub URL with www", "https://www.github.com/username/repository", true},
		{"Valid GitHub URL with .git", "https://github.com/username/repository.git", true},
		{"Not GitHub - GitLab", "https://gitlab.com/username/repository", false},
		{"Not GitHub - Bitbucket", "https://bitbucket.org/username/repository", false},
		{"GitHub URL without repository", "https://github.com/username", false},
		{"GitHub URL with empty repository", "https://github.com/username/", false},
		{"GitHub URL with valid username with hyphens", "https://github.com/user-name/repository", true},
		{"GitHub URL with valid repository with underscores", "https://github.com/username/repo_name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGitHubURL(tt.url)
			if result != tt.expected {
				t.Errorf("isGitHubURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}
