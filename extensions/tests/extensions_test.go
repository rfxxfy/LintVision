package extensions_test

import (
	"testing"

	"github.com/rfxxfy/LintVision/extensions"
	"github.com/rfxxfy/LintVision/extensions/categories"
	"github.com/rfxxfy/LintVision/extensions/languages"
)

func setupTestData() func() {
	// Сохраняем старое состояние
	oldConfigs := languages.Configs
	oldCats := categories.CategoryOfExt

	// Подменяем на тестовые значения
	languages.Configs = map[string]languages.LanguageConfig{
		".go": {Name: "Go", SingleLineComment: "//"},
		".py": {Name: "Python", SingleLineComment: "#"},
	}
	categories.CategoryOfExt = map[string]string{
		".md":  "markup",
		".yml": "config",
	}

	// Возвращаем старое состояние после теста
	return func() {
		languages.Configs = oldConfigs
		categories.CategoryOfExt = oldCats
	}
}

func TestIsCodeExtension(t *testing.T) {
	defer setupTestData()()
	tests := []struct {
		ext  string
		want bool
	}{
		{".go", true},
		{".py", true},
		{".md", false},
		{".txt", false},
	}
	for _, tt := range tests {
		got := extensions.IsCodeExtension(tt.ext)
		if got != tt.want {
			t.Errorf("IsCodeExtension(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func TestGetLanguageConfig(t *testing.T) {
	defer setupTestData()()
	tests := []struct {
		ext      string
		wantOk   bool
		wantName string
		wantComm string
	}{
		{".go", true, "Go", "//"},
		{".py", true, "Python", "#"},
		{".md", false, "", ""},
	}
	for _, tt := range tests {
		cfg, ok := extensions.GetLanguageConfig(tt.ext)
		if ok != tt.wantOk {
			t.Errorf("GetLanguageConfig(%q) ok = %v, want %v", tt.ext, ok, tt.wantOk)
		}
		if ok {
			if cfg.Name != tt.wantName {
				t.Errorf("GetLanguageConfig(%q) Name = %q, want %q", tt.ext, cfg.Name, tt.wantName)
			}
			if cfg.SingleLineComment != tt.wantComm {
				t.Errorf("GetLanguageConfig(%q) SingleLineComment = %q, want %q", tt.ext, cfg.SingleLineComment, tt.wantComm)
			}
		}
	}
}

func TestGetFileCategory(t *testing.T) {
	defer setupTestData()()
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "code"},
		{".py", "code"},
		{".md", "markup"},
		{".yml", "config"},
		{".txt", "unknown"},
	}
	for _, tt := range tests {
		got := extensions.GetFileCategory(tt.ext)
		if got != tt.want {
			t.Errorf("GetFileCategory(%q) = %q, want %q", tt.ext, got, tt.want)
		}
	}
}

func TestIsCommentAfterCode(t *testing.T) {
	defer setupTestData()()
	type args struct {
		line string
		ext  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Go: comment after code",
			args: args{line: `fmt.Println("hi") // comment`, ext: ".go"},
			want: true,
		},
		{
			name: "Go: comment at start",
			args: args{line: `// comment`, ext: ".go"},
			want: false,
		},
		{
			name: "Go: comment inside string",
			args: args{line: `fmt.Println("// not comment")`, ext: ".go"},
			want: false,
		},
		{
			name: "Go: comment after code with quote in code",
			args: args{line: `fmt.Println("\"hi\"") // comment`, ext: ".go"},
			want: true,
		},
		{
			name: "Py: comment after code",
			args: args{line: `print("hi") # comment`, ext: ".py"},
			want: true,
		},
		{
			name: "Py: comment at start",
			args: args{line: `# comment`, ext: ".py"},
			want: false,
		},
		{
			name: "Unknown ext",
			args: args{line: `print("hi") # comment`, ext: ".txt"},
			want: false,
		},
		{
			name: "Go: comment after code with odd quotes",
			args: args{line: `fmt.Println("\"hi\" // not comment") // comment`, ext: ".go"},
			want: true,
		},
		{
			name: "Go: comment after code with odd number of quotes before //",
			args: args{line: `fmt.Println("\"hi // not comment") // comment`, ext: ".go"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extensions.IsCommentAfterCode(tt.args.line, tt.args.ext)
			if got != tt.want {
				t.Errorf("IsCommentAfterCode(%q, %q) = %v, want %v", tt.args.line, tt.args.ext, got, tt.want)
			}
		})
	}
}