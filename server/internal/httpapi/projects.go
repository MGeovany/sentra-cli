package httpapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/mgeovany/sentra/server/internal/auth"
	"github.com/mgeovany/sentra/server/internal/repo"
)

func projectsHandler(store repo.ProjectStore) http.Handler {
	if store == nil {
		store = repo.DisabledProjectStore{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		user, ok := auth.UserFromContext(r.Context())
		if !ok || strings.TrimSpace(user.ID) == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		projects, err := store.ListProjects(r.Context(), user.ID)
		if err != nil {
			log.Printf("projects list failed user_id=%s err=%v", user.ID, err)
			switch err {
			case repo.ErrDBNotConfigured:
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = io.WriteString(w, "db not configured")
			default:
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = io.WriteString(w, "projects failed")
			}
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(projects)
	})
}
