package httpapi

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/mgeovany/sentra/server/internal/auth"
	"github.com/mgeovany/sentra/server/internal/repo"
)

func requireDeviceSignature(store repo.MachineStore, next http.Handler) http.Handler {
	if store == nil {
		store = repo.DisabledMachineStore{}
	}
	if next == nil {
		next = http.NotFoundHandler()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || strings.TrimSpace(user.ID) == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "unauthorized")
			return
		}

		machineID := strings.TrimSpace(r.Header.Get("X-Sentra-Machine-ID"))
		ts := strings.TrimSpace(r.Header.Get("X-Sentra-Timestamp"))
		sig := strings.TrimSpace(r.Header.Get("X-Sentra-Signature"))
		if machineID == "" || ts == "" || sig == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "unauthorized")
			return
		}

		pub, pubOK, err := store.DevicePubKey(r.Context(), user.ID, machineID)
		if err != nil {
			if err == repo.ErrDBNotConfigured {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = io.WriteString(w, "db not configured")
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "unauthorized")
			return
		}
		if !pubOK || strings.TrimSpace(pub) == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "unauthorized")
			return
		}

		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewReader(body))

		if err := auth.VerifyDeviceSignature(pub, machineID, ts, r.Method, r.URL.Path, body, sig); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}
