package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type remoteProject struct {
	RootPath          string `json:"root_path"`
	LastCommitID      string `json:"last_commit_id"`
	LastCommitMessage string `json:"last_commit_message"`
	FileCount         int    `json:"file_count"`
}

func runProjects() error {
	sess, err := ensureRemoteSession()
	if err != nil {
		return err
	}
	if strings.TrimSpace(sess.AccessToken) == "" {
		return fmt.Errorf("not logged in (run: sentra login)")
	}

	serverURL, err := serverURLFromEnv()
	if err != nil {
		return err
	}

	endpoint := serverURL + "/projects"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(sess.AccessToken))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to fetch projects")
	}

	var projects []remoteProject
	if err := json.Unmarshal(body, &projects); err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Println("✔ 0 projects")
		return nil
	}

	for _, p := range projects {
		root := strings.TrimSpace(p.RootPath)
		if root == "" {
			root = "(unknown)"
		}

		last := strings.TrimSpace(p.LastCommitID)
		if last == "" {
			last = "-"
		} else if len(last) > 8 {
			last = last[:8]
		}

		msg := strings.TrimSpace(p.LastCommitMessage)
		if msg != "" {
			fmt.Printf("%s\t%s\t%d\t%s\n", root, last, p.FileCount, msg)
			continue
		}
		fmt.Printf("%s\t%s\t%d\n", root, last, p.FileCount)
	}

	return nil
}
