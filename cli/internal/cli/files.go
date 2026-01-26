package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type remoteFile struct {
	CommitID string `json:"commit_id"`
	Path     string `json:"file_path"`
	SHA256   string `json:"sha256"`
	Size     int    `json:"size"`
}

func runFiles(args []string) error {
	root, at, err := parseFilesArgs(args)
	if err != nil {
		return err
	}

	sess, err := ensureRemoteSession()
	if err != nil {
		return err
	}
	if strings.TrimSpace(sess.AccessToken) == "" {
		return errors.New("not logged in (run: sentra login)")
	}

	serverURL, err := serverURLFromEnv()
	if err != nil {
		return err
	}

	u, err := url.Parse(serverURL + "/files")
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("root", root)
	if at != "" {
		q.Set("at", at)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(sess.AccessToken))

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to fetch files")
	}

	var files []remoteFile
	if err := json.Unmarshal(body, &files); err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("✔ 0 files")
		return nil
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	for _, f := range files {
		fmt.Printf("%s\t%d\t%s\n", strings.TrimSpace(f.Path), f.Size, strings.TrimSpace(f.SHA256))
	}

	return nil
}

func parseFilesArgs(args []string) (root string, at string, err error) {
	if len(args) < 1 {
		return "", "", errors.New("usage: sentra files <project> [--at <commit>]")
	}

	root = projectRootFromPath(args[0])
	root = strings.TrimSpace(root)
	if root == "" {
		return "", "", errors.New("usage: sentra files <project> [--at <commit>]")
	}

	if len(args) == 1 {
		return root, "", nil
	}
	if len(args) != 3 || args[1] != "--at" {
		return "", "", errors.New("usage: sentra files <project> [--at <commit>]")
	}
	at = strings.TrimSpace(args[2])
	if at == "" {
		return "", "", errors.New("usage: sentra files <project> [--at <commit>]")
	}
	return root, at, nil
}
