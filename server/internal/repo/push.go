package repo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mgeovany/sentra/server/internal/supabase"
)

type PushStore interface {
	Push(ctx context.Context, userID string, payload any) (PushResult, error)
}

type PushResult struct {
	ProjectID  string `json:"out_project_id"`
	CommitID   string `json:"out_commit_id"`
	ReceivedAt string `json:"received_at"`
	Deduped    bool   `json:"deduped"`
}

type DisabledPushStore struct{}

func (DisabledPushStore) Push(ctx context.Context, userID string, payload any) (PushResult, error) {
	return PushResult{}, ErrDBNotConfigured
}

type SupabasePushStore struct {
	client *supabase.Client
	fn     string
}

func NewSupabasePushStore(client *supabase.Client, fn string) SupabasePushStore {
	if fn == "" {
		fn = "sentra_push_v1"
	}
	return SupabasePushStore{client: client, fn: fn}
}

func (s SupabasePushStore) Push(ctx context.Context, userID string, payload any) (PushResult, error) {
	if s.client == nil {
		return PushResult{}, ErrDBNotConfigured
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return PushResult{}, fmt.Errorf("invalid push: missing user_id")
	}

	url := s.client.RPCURL(s.fn)
	body := map[string]any{
		"p_user_id": userID,
		"p_payload": payload,
	}

	headers := map[string]string{
		"Accept": "application/json",
		// We want a JSON response.
		"Prefer": "return=representation",
	}

	resp, respBody, err := s.client.PostJSON(ctx, url, body, headers)
	if err != nil {
		return PushResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return PushResult{}, fmt.Errorf("supabase rpc push failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	// PostgREST returns an array for RPC results.
	var out []PushResult
	if err := supabase.UnmarshalJSON(respBody, &out); err != nil {
		return PushResult{}, err
	}
	if len(out) == 0 {
		return PushResult{}, fmt.Errorf("supabase rpc push returned empty result")
	}
	return out[0], nil
}

// Supabase uses its own json package in this repo; keep this local helper.
var _ = http.MethodPost
