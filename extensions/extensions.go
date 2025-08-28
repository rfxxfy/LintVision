package extensions

import (
	"strings"

	"github.com/rfxxfy/LintVision/extensions/categories"
	"github.com/rfxxfy/LintVision/extensions/languages"
)

func IsCodeExtension(ext string) bool {
	_, ok := languages.Configs[ext]
	return ok
}

func GetLanguageConfig(ext string) (languages.LanguageConfig, bool) {
	cfg, ok := languages.Configs[ext]
	return cfg, ok
}

func GetFileCategory(ext string) string {
	if IsCodeExtension(ext) {
		return "code"
	}
	if cat, ok := categories.CategoryOfExt[ext]; ok {
		return cat
	}
	return "unknown"
}

func IsCommentAfterCode(line, ext string) bool {
	cfg, ok := GetLanguageConfig(ext)
	if !ok {
		return false
	}
	token := cfg.SingleLineCommentToken
	idx := strings.Index(line, token)
	if idx <= 0 {
		return false
	}
	count := 0
	for i := 0; i < idx; i++ {
		ch := line[i]
		if ch == '"' || ch == '\'' {
			count++
		}
	}
	return count%2 == 0
}
