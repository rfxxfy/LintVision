package stats

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func ComputeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file for hashing: %w", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("stat file for hashing: %w", err)
	}
	if !fi.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", path)
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("read file for hashing: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
