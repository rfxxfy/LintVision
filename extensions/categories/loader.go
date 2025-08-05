package categories

import (
	_ "embed"
	"encoding/json"

	"github.com/rfxxfy/LintVision/logging"
)

//go:embed config.json
var configData []byte

func init() {
	if err := json.Unmarshal(configData, &Definitions); err != nil {
		logging.Fatal("categories: cannot unmarshal config.json: %v", err)
	}

	CategoryOfExt = make(map[string]string, len(Definitions)*4)
	for category, exts := range Definitions {
		for _, ext := range exts {
			CategoryOfExt[ext] = category
		}
	}

	logging.Info("categories: loaded %d categories", len(Definitions))
}
