// main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfxxfy/LintVision/logging"
	"github.com/rfxxfy/LintVision/metrics"
)

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}
	path = os.ExpandEnv(path)

	return path, nil
}

func main() {
	dir := flag.String("path", ".", "директория для анализа")
	logCfg := flag.String("log-config", "", "конфиг логгера")
	out := flag.String("out", "", "файл для сохранения результата JSON")
	flag.Parse()

	if *logCfg != "" {
		if err := logging.LoadConfig(*logCfg); err != nil {
			fmt.Fprintf(os.Stderr, "cannot load logging config: %v\n", err)
			os.Exit(1)
		}
	}

	if _, err := metrics.AnalyzeAndSave(*dir, *out); err != nil {
		logging.Fatal("analysis failed: %v", err)
	}
}
