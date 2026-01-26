package cli

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

func serverURLFromEnv() (string, error) {
	v := strings.TrimSpace(os.Getenv("SENTRA_SERVER_URL"))
	if v == "" {
		port := strings.TrimSpace(os.Getenv("SERVER_PORT"))
		if port == "" {
			port = "8080"
		}
		return "http://127.0.0.1:" + port, nil
	}

	u, err := url.Parse(v)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid server URL scheme: %s", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("invalid server URL host")
	}

	if u.Scheme == "http" && !isLoopbackHost(host) {
		return "", fmt.Errorf("insecure connection: HTTP is only allowed for localhost connections")
	}

	return strings.TrimRight(v, "/"), nil
}

func isLoopbackHost(host string) bool {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil {
		return ip.IsLoopback()
	}
	return false
}
