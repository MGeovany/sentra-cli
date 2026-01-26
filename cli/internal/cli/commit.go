package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mgeovany/sentra/cli/internal/commit"
	"github.com/mgeovany/sentra/cli/internal/index"
)

func runCommit(args []string) error {
	verbosef("Starting commit operation...")
	message, err := parseCommitMessage(args)
	if err != nil {
		return err
	}
	verbosef("Commit message: %s", message)

	indexPath, err := index.DefaultPath()
	if err != nil {
		return err
	}
	verbosef("Index path: %s", indexPath)

	idx, ok, err := index.Load(indexPath)
	if err != nil {
		return err
	}
	if !ok || len(idx.Staged) == 0 {
		return errors.New("nothing to commit (no staged env files)")
	}
	verbosef("Found %d staged file(s)", len(idx.Staged))
	for path, hash := range idx.Staged {
		verbosef("  - %s (hash: %s)", path, hash)
	}

	cm := commit.New(message, idx.Staged)
	verbosef("Created commit: %s", cm.ID)
	if _, err := commit.Save(cm); err != nil {
		return err
	}
	verbosef("Commit saved to local storage")

	idx.Staged = map[string]string{}
	if err := index.Save(indexPath, idx); err != nil {
		return err
	}
	verbosef("Index cleared and saved")

	// Keep this one a bit more celebratory.
	shortID := cm.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	fmt.Println(c(ansiGreen, "✔ committed ") + c(ansiBoldCyan, shortID))
	verbosef("Commit %s created with %d file(s)", cm.ID, len(cm.Files))
	return nil
}

func parseCommitMessage(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("usage: sentra commit -m 'message'")
	}

	var message string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-m":
			if i+1 >= len(args) {
				return "", errors.New("usage: sentra commit -m 'message'")
			}
			message = args[i+1]
			i++
		default:
			return "", errors.New("usage: sentra commit -m 'message'")
		}
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return "", errors.New("commit message cannot be empty")
	}

	return message, nil
}
