package state

import (
	"os"
	"path/filepath"
)

func StateDir() string {
	dir := os.Getenv("ASEARCH_STATE_DIR")
	if dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".asearch")
	}
	return filepath.Join(home, ".asearch")
}

func SessionsDir() string {
	return filepath.Join(StateDir(), "sessions")
}

func ResultsDir() string {
	return filepath.Join(StateDir(), "results")
}

func EnsureDirs() error {
	for _, d := range []string{StateDir(), SessionsDir(), ResultsDir()} {
		if err := os.MkdirAll(d, 0700); err != nil {
			return err
		}
	}
	return nil
}
