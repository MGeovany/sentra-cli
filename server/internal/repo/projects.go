package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mgeovany/sentra/server/internal/supabase"
)

type ProjectInfo struct {
	RootPath          string `json:"root_path"`
	LastCommitID      string `json:"last_commit_id"`
	LastCommitMessage string `json:"last_commit_message"`
	FileCount         int    `json:"file_count"`
}

type ProjectStore interface {
	ListProjects(ctx context.Context, userID string) ([]ProjectInfo, error)
}

type DisabledProjectStore struct{}

func (DisabledProjectStore) ListProjects(ctx context.Context, userID string) ([]ProjectInfo, error) {
	return nil, ErrDBNotConfigured
}

type SupabaseProjectStore struct {
	client *supabase.Client
	fn     string
}

func NewSupabaseProjectStore(client *supabase.Client, fn string) SupabaseProjectStore {
	if fn == "" {
		fn = "sentra_projects_v1"
	}
	return SupabaseProjectStore{client: client, fn: fn}
}

func (s SupabaseProjectStore) ListProjects(ctx context.Context, userID string) ([]ProjectInfo, error) {
	if s.client == nil {
		return nil, ErrDBNotConfigured
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("invalid projects request: missing user_id")
	}

	url := s.client.RPCURL(s.fn)
	body := map[string]any{
		"p_user_id": userID,
	}
	headers := map[string]string{
		"Accept": "application/json",
		"Prefer": "return=representation",
	}

	resp, respBody, err := s.client.PostJSON(ctx, url, body, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("supabase rpc projects failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var out []ProjectInfo
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	return out, nil
}

var _ = http.MethodPost
