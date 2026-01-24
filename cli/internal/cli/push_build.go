package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mgeovany/sentra/cli/internal/auth"
	"github.com/mgeovany/sentra/cli/internal/commit"
)

func buildPushRequestV1(scanRoot, machineID, machineName string, c commit.Commit) ([]pushRequestV1, error) {
	pathsByRoot := map[string][]string{}
	for p := range c.Files {
		root := projectRootFromPath(p)
		if root == "" {
			continue
		}
		pathsByRoot[root] = append(pathsByRoot[root], p)
	}
	if len(pathsByRoot) == 0 {
		return nil, fmt.Errorf("cannot determine project.root")
	}

	roots := make([]string, 0, len(pathsByRoot))
	for root := range pathsByRoot {
		roots = append(roots, root)
	}
	sort.Strings(roots)

	clientID := strings.TrimSpace(c.ID)
	if _, err := uuid.Parse(clientID); err != nil {
		// Backward-compat: old commits used timestamp-based IDs.
		// Keep a stable idempotency key derived from the old ID.
		clientID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(clientID)).String()
	}

	out := make([]pushRequestV1, 0, len(roots))
	for _, root := range roots {
		paths := pathsByRoot[root]
		sort.Strings(paths)

		files := make([]pushFileV1, 0, len(paths))
		for _, p := range paths {
			abs := filepath.Join(scanRoot, filepath.FromSlash(p))
			plain, err := os.ReadFile(abs)
			if err != nil {
				return nil, fmt.Errorf("cannot read %s: %w", p, err)
			}

			shaPlain := auth.SHA256Hex(plain)
			cipherName, blob, size, err := auth.EncryptEnvBlob(plain)
			if err != nil {
				return nil, err
			}

			files = append(files, pushFileV1{
				Path:      p,
				SHA256:    shaPlain,
				Size:      size,
				Encrypted: true,
				Cipher:    cipherName,
				Blob:      blob,
			})
		}

		out = append(out, pushRequestV1{
			V:       1,
			Project: pushProjectV1{Root: strings.TrimSpace(root)},
			Machine: pushMachineV1{ID: machineID, Name: machineName},
			Commit: pushCommitV1{
				ClientID: clientID,
				Message:  strings.TrimSpace(c.Message),
			},
			Files: files,
		})
	}

	return out, nil
}

func projectRootFromPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return ""
	}
	parts := strings.Split(p, "/")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}
