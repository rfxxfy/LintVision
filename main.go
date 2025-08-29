package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rfxxfy/LintVision/logging"
	"github.com/rfxxfy/LintVision/stats"
)

func main() {
	guiMode := flag.Bool("gui", false, "Запустить в GUI режиме")
	dir := flag.String("path", ".", "директория для анализа")
	logCfg := flag.String("log-config", "", "конфиг логгера")
	out := flag.String("out", "", "файл для сохранения результата JSON")
	flag.Parse()

	if *guiMode {
		gui := NewLintVisionGUI()
		gui.Run()
		return
	}

	if *logCfg != "" {
		if err := logging.LoadConfig(*logCfg); err != nil {
			fmt.Fprintf(os.Stderr, "cannot load logging config: %v\n", err)
			os.Exit(1)
		}
	}

	if _, err := stats.AnalyzeAndSave(*dir, *out); err != nil {
		logging.Fatal("analysis failed: %v", err)
	}
}
