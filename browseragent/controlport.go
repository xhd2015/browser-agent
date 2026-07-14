package browseragent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ComposeControlAddr joins host and port into host:port, applying product defaults.
func ComposeControlAddr(host string, port int) string {
	host = strings.TrimSpace(host)
	if host == "" {
		host = DefaultControlHost
	}
	if port <= 0 {
		port = DefaultControlPort
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// ResolveEnsureAddr returns the control listen address for EnsureDaemon/spawn paths.
func ResolveEnsureAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr != "" {
		return addr
	}
	return DefaultAddr
}

// ForeignPortError returns the Q3 foreign-listener error text.
func ForeignPortError(host string, port int) error {
	addr := ComposeControlAddr(host, port)
	msg := fmt.Sprintf("browser-agent: error: control port %s is in use by another process (not browser-agent daemon)", addr)
	hint := "stop the other process, or use --server-port <port> with matching serve --port <port>"
	return fmt.Errorf("%s\nhint: %s", msg, hint)
}

// CheckForeignControlPort reports an error when host:port is occupied by a non-browser-agent listener.
func CheckForeignControlPort(addr string) error {
	host, portStr, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}
	baseURL := "http://" + net.JoinHostPort(host, portStr)
	if isBrowserAgentHealth(baseURL) {
		return nil
	}
	if portReachable(baseURL) {
		return ForeignPortError(host, port)
	}
	return nil
}

func portReachable(baseURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/v1/health", nil)
	if err != nil {
		return false
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	return true
}

type healthPayload struct {
	OK            bool   `json:"ok"`
	Product       string `json:"product"`
	DaemonVersion string `json:"daemon_version"`
	BaseDir       string `json:"base_dir"`
}

func fetchHealthPayload(baseURL string) (healthPayload, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/v1/health", nil)
	if err != nil {
		return healthPayload{}, false
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return healthPayload{}, false
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return healthPayload{}, false
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return healthPayload{}, false
	}
	var hp healthPayload
	if err := json.Unmarshal(body, &hp); err != nil {
		return healthPayload{}, false
	}
	if !hp.OK {
		return healthPayload{}, false
	}
	return hp, true
}

func isBrowserAgentHealth(baseURL string) bool {
	hp, ok := fetchHealthPayload(baseURL)
	if !ok {
		return false
	}
	if hp.Product != "" && hp.Product != ProductName {
		return false
	}
	return true
}

func daemonHealthOK(baseURL string) bool {
	return isBrowserAgentHealth(baseURL)
}

func fetchDaemonVersion(baseURL string) string {
	hp, ok := fetchHealthPayload(baseURL)
	if !ok {
		return ""
	}
	return hp.DaemonVersion
}