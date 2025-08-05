package languages

type LanguageConfig struct {
	Name              string `json:"name"`
	SingleLineComment string `json:"singleLineComment"`
	DoubleQuote       string `json:"doubleQuote"`
	SingleQuote       string `json:"singleQuote"`
}

var Configs map[string]LanguageConfig
