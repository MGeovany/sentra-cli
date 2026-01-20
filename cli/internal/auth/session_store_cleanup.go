package auth

import (
	"errors"
	"os"
)

func removeLegacySessionFiles() error {
	// Best-effort cleanup; ignore errors where practical.
	if p, err := sessionPath(); err == nil {
		_ = os.Remove(p)
	}
	if p, err := sessionKeyPath(); err == nil {
		_ = os.Remove(p)
	}
	return nil
}

func removeLegacySessionFilesStrict() error {
	if p, err := sessionPath(); err == nil {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if p, err := sessionKeyPath(); err == nil {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
