package languages

type LanguageConfig struct {
    SingleLineCommentToken string `json:"singleLineCommentToken"`
    DoubleQuote            string `json:"doubleQuote"`
    SingleQuote            string `json:"singleQuote"`
}

var Configs map[string]LanguageConfig
