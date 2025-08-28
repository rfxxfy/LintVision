package extensions_test

import (
	"testing"

	"github.com/rfxxfy/LintVision/extensions"
	"github.com/stretchr/testify/assert"
)

// TODO: заменить на моки и интерфейсы, когда появится возможность
// Сейчас тесты используют реальные данные из extensions

func TestIsCodeExtension(t *testing.T) {
	t.Parallel()
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
		tt := tt // захват переменной для параллельного запуска
		t.Run(tt.ext, func(t *testing.T) {
			t.Parallel()
			got := extensions.IsCodeExtension(tt.ext)
			assert.Equal(t, tt.want, got, "IsCodeExtension(%q)", tt.ext)
			// TODO: logging here
		})
	}
}

func TestGetLanguageConfig(t *testing.T) {
	t.Parallel()
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
		tt := tt
		t.Run(tt.ext, func(t *testing.T) {
			t.Parallel()
			cfg, ok := extensions.GetLanguageConfig(tt.ext)
			assert.Equal(t, tt.wantOk, ok, "GetLanguageConfig(%q) ok", tt.ext)
			if ok {
				assert.Equal(t, tt.wantComm, cfg.SingleLineCommentToken, "GetLanguageConfig(%q) SingleLineCommentToken", tt.ext)
			}
			// TODO: logging here
		})
	}
}

func TestGetFileCategory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "code"},
		{".py", "code"},
		{".md", "markup"},
		{".yml", "markup"},
		{".txt", "document"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.ext, func(t *testing.T) {
			t.Parallel()
			got := extensions.GetFileCategory(tt.ext)
			assert.Equal(t, tt.want, got, "GetFileCategory(%q)", tt.ext)
			// TODO: logging here
		})
	}
}

func TestIsCommentAfterCode(t *testing.T) {
	t.Parallel()
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
			want: false,
		},
		{
			name: "Go: comment after code with odd number of quotes before //",
			args: args{line: `fmt.Println("\"hi // not comment") // comment`, ext: ".go"},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extensions.IsCommentAfterCode(tt.args.line, tt.args.ext)
			assert.Equal(t, tt.want, got, "IsCommentAfterCode(%q, %q)", tt.args.line, tt.args.ext)
			// TODO: logging here
		})
	}
}