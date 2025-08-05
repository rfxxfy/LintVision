package logging

import (
	"fmt"
	"strings"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

func (l Level) String() string {
	if int(l) < 0 || int(l) >= len(levelNames) {
		return fmt.Sprintf("Level(%d)", l)
	}
	return levelNames[l]
}

func ParseLevel(s string) (Level, error) {
	us := strings.ToUpper(strings.TrimSpace(s))
	for i, name := range levelNames {
		if name == us {
			return Level(i), nil
		}
	}
	return Level(0), fmt.Errorf("logging: unknown level %q", s)
}
