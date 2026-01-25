package repo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mgeovany/sentra/server/internal/supabase"
)

type FileInfo struct {
	CommitID string `json:"commit_id"`
	FilePath string `json:"file_path"`
	SHA256   string `json:"sha256"`
	Size     int    `json:"size"`
}

type FileStore interface {
	ListFiles(ctx context.Context, userID string, root string, at string) ([]FileInfo, error)
}

type DisabledFileStore struct{}

func (DisabledFileStore) ListFiles(ctx context.Context, userID string, root string, at string) ([]FileInfo, error) {
	return nil, ErrDBNotConfigured
}

type SupabaseFileStore struct {
	client *supabase.Client
	fn     string
}

func NewSupabaseFileStore(client *supabase.Client, fn string) SupabaseFileStore {
	if fn == "" {
		fn = "sentra_files_v1"
	}
	return SupabaseFileStore{client: client, fn: fn}
}

func (s SupabaseFileStore) ListFiles(ctx context.Context, userID string, root string, at string) ([]FileInfo, error) {
	if s.client == nil {
		return nil, ErrDBNotConfigured
	}
	userID = strings.TrimSpace(userID)
	root = strings.TrimSpace(root)
	at = strings.TrimSpace(at)
	if userID == "" || root == "" {
		return nil, fmt.Errorf("invalid files request")
	}

	url := s.client.RPCURL(s.fn)
	body := map[string]any{
		"p_user_id": userID,
		"p_root":    root,
		"p_at":      at,
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
		return nil, fmt.Errorf("supabase rpc files failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var out []FileInfo
	if err := supabase.UnmarshalJSON(respBody, &out); err != nil {
		return nil, err
	}
	return out, nil
}

var _ = http.MethodPost
