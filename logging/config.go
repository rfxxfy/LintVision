package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Level      string `json:"level"`       // DEBUG, INFO, …
	Output     string `json:"output"`      // "stdout", "stderr" или путь к файлу
	TimeFormat string `json:"time_format"` // e.g. "2006-01-02 15:04:05"
	Format     string `json:"format"`      // "text" или "json"
	Caller     bool   `json:"caller"`      // включить вывод caller
}

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("logging: cannot read config file %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("logging: invalid JSON in %q: %w", path, err)
	}

	lvl, err := ParseLevel(cfg.Level)
	if err != nil {
		return err
	}
	SetLevel(lvl)

	switch strings.ToLower(cfg.Output) {
	case "", "stdout":
		SetOutput(os.Stdout)
	case "stderr":
		SetOutput(os.Stderr)
	default:
		f, err := os.OpenFile(cfg.Output,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("logging: cannot open output file %q: %w", cfg.Output, err)
		}
		SetOutput(f)
	}

	if cfg.TimeFormat != "" {
		std.timeFormat = cfg.TimeFormat
	}

	if strings.EqualFold(cfg.Format, "json") {
		SetFormat(JSONFormat)
	} else {
		SetFormat(TextFormat)
	}

	SetCaller(cfg.Caller)
	return nil
}
