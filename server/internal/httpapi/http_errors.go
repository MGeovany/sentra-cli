package httpapi

import (
	"io"
	"net/http"
	"os"
	"strings"
)

// writeHTTPError returns a minimal message by default.
// If SENTRA_DEBUG_HTTP_ERRORS=1, it returns the full error text to help local debugging.
func writeHTTPError(w http.ResponseWriter, status int, publicMsg string, err error) {
	w.WriteHeader(status)

	if os.Getenv("SENTRA_DEBUG_HTTP_ERRORS") == "1" && err != nil {
		msg := strings.TrimSpace(err.Error())
		if msg == "" {
			msg = publicMsg
		}
		// Avoid returning unbounded error strings.
		if len(msg) > 4000 {
			msg = msg[:4000]
		}
		_, _ = io.WriteString(w, msg)
		return
	}

	_, _ = io.WriteString(w, publicMsg)
}
