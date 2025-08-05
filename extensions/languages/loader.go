package languages

import (
	_ "embed"
	"encoding/json"

	"github.com/rfxxfy/LintVision/logging"
)

var (
	//go:embed config.json
	configData []byte
)

func init() {
	Configs = make(map[string]LanguageConfig)
	if err := json.Unmarshal(configData, &Configs); err != nil {
		logging.Fatal("languages: cannot unmarshal config.json: %v", err)
	}
}
